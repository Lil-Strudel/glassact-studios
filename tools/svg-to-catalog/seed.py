import json
import math
import os
import re
import sys
from io import BytesIO
from pathlib import Path

import cairosvg
import requests
from PIL import Image

API_BASE = os.environ.get("API_BASE", "http://localhost:4100")
AUTH_TOKEN = os.environ.get("AUTH_TOKEN", "")

PROGRESS_FILE = Path(__file__).parent / "seed_progress.json"

# Every seeded item defaults to PG-1. Prices are refined later per-item/per-proof.
DEFAULT_PRICE_GROUP_ID = 1
# Aspect-derived sizing: the longer physical side of each inlay, in inches.
LONG_SIDE_INCHES = 4.0
# Cap the rasterization width when measuring the content bbox, for speed on
# sources with very large viewBoxes.
MAX_RENDER_WIDTH = 2000
# Fallback grout target when analyze leaves the grout region unmatched: the grout
# nearest to black (reliably a dark granite), mirroring the implicit-black grout.
BLACK = "#000000"


# ---------------------------------------------------------------------------
# HTTP helpers
# ---------------------------------------------------------------------------

def headers() -> dict:
    return {"Authorization": f"Bearer {AUTH_TOKEN}"}


def upload_bytes(name: str, data: bytes, content_type: str) -> str | None:
    resp = requests.post(
        f"{API_BASE}/api/upload",
        headers=headers(),
        files={"file": (name, data, content_type)},
    )
    if not resp.ok:
        print(f"  Error uploading {name}: {resp.status_code} {resp.text}", file=sys.stderr)
        return None
    return resp.json()["url"]


def analyze(svg_url: str) -> dict | None:
    resp = requests.post(
        f"{API_BASE}/api/catalog/analyze",
        headers={**headers(), "Content-Type": "application/json"},
        json={"svg_url": svg_url},
    )
    if not resp.ok:
        print(f"  Error analyzing {svg_url}: {resp.status_code} {resp.text}", file=sys.stderr)
        return None
    return resp.json()


def create_catalog_item(body: dict) -> tuple[bool, str | None]:
    resp = requests.post(
        f"{API_BASE}/api/catalog",
        headers={**headers(), "Content-Type": "application/json"},
        json=body,
    )
    if not resp.ok:
        print(
            f"  Error creating catalog item {body['catalog_code']}: {resp.status_code} {resp.text}",
            file=sys.stderr,
        )
        return False, None
    return True, resp.json().get("uuid")


def fetch_palette(path: str) -> list[dict]:
    resp = requests.get(f"{API_BASE}{path}", headers=headers())
    resp.raise_for_status()
    items = resp.json() or []
    # Only active colors are valid at bake time (the server bakes against the
    # active palettes), so an id we assign must resolve there.
    return [{"id": c["id"], "hex": c["hex"]} for c in items if c.get("is_active", True)]


# ---------------------------------------------------------------------------
# Color matching — ports apps/api/svg/colormatch.go (sRGB -> Lab, ΔE76), with no
# distance threshold so it always returns a nearest id.
# ---------------------------------------------------------------------------

_HEX6 = re.compile(r"^#[0-9a-fA-F]{6}$")
_HEX3 = re.compile(r"^#[0-9a-fA-F]{3}$")


def normalize_hex(s: str) -> str | None:
    s = s.strip().lower()
    if _HEX6.match(s):
        return s
    if _HEX3.match(s):
        return "#" + s[1] * 2 + s[2] * 2 + s[3] * 2
    return None


def _srgb_to_linear(c: float) -> float:
    if c <= 0.04045:
        return c / 12.92
    return ((c + 0.055) / 1.055) ** 2.4


def _lab_f(t: float) -> float:
    delta = 6.0 / 29.0
    if t > delta ** 3:
        return t ** (1.0 / 3.0)
    return t / (3 * delta * delta) + 4.0 / 29.0


def hex_to_lab(hex_str: str) -> tuple[float, float, float] | None:
    h = normalize_hex(hex_str)
    if h is None:
        return None
    r = int(h[1:3], 16) / 255.0
    g = int(h[3:5], 16) / 255.0
    b = int(h[5:7], 16) / 255.0
    rl, gl, bl = _srgb_to_linear(r), _srgb_to_linear(g), _srgb_to_linear(b)

    x = rl * 0.4124564 + gl * 0.3575761 + bl * 0.1804375
    y = rl * 0.2126729 + gl * 0.7151522 + bl * 0.0721750
    z = rl * 0.0193339 + gl * 0.1191920 + bl * 0.9503041

    xn, yn, zn = 0.95047, 1.00000, 1.08883
    fx, fy, fz = _lab_f(x / xn), _lab_f(y / yn), _lab_f(z / zn)
    return (116 * fy - 16, 500 * (fx - fy), 200 * (fy - fz))


def _delta_e76(a: tuple, b: tuple) -> float:
    return math.sqrt((a[0] - b[0]) ** 2 + (a[1] - b[1]) ** 2 + (a[2] - b[2]) ** 2)


def nearest_id(hex_str: str, palette: list[dict]) -> int | None:
    src = hex_to_lab(hex_str)
    if src is None or not palette:
        return None
    best_id, best_dist = None, math.inf
    for c in palette:
        lab = hex_to_lab(c["hex"])
        if lab is None:
            continue
        d = _delta_e76(src, lab)
        if d < best_dist:
            best_id, best_dist = c["id"], d
    return best_id


# ---------------------------------------------------------------------------
# Manifest + geometry — the browser's job, done headlessly.
# ---------------------------------------------------------------------------

def fill_manifest(manifest: dict, glass_palette: list[dict], default_grout_id: int) -> list[str]:
    """Assign every null glass/grout id in place. Returns human-readable notes.

    analyze best-guesses ids within ΔE76 ≤ 25 and leaves far colors null; the
    create endpoint rejects any null, so fill them with an un-thresholded nearest
    match (glass regions carry their source_hex; the grout region does not, so it
    falls back to the nearest-black grout).
    """
    notes: list[str] = []

    for key, region in (manifest.get("glass_regions") or {}).items():
        if region.get("glass_color_id") is not None:
            continue
        source_hex = region.get("source_hex")
        match = nearest_id(source_hex, glass_palette) if source_hex else None
        if match is None:
            match = glass_palette[0]["id"] if glass_palette else None
            notes.append(f"glass group {key} (hex {source_hex}) had no source match; defaulted")
        region["glass_color_id"] = match

    grout = manifest.get("grout_region") or {}
    if grout.get("grout_id") is None:
        grout["grout_id"] = default_grout_id
        notes.append("grout region unmatched; defaulted to nearest-black grout")
    manifest["grout_region"] = grout

    return notes


_VIEWBOX_RE = re.compile(r'viewBox\s*=\s*"([^"]+)"', re.IGNORECASE)
_DIM_RE = re.compile(r"-?\d*\.?\d+")


def parse_view_box(svg_text: str) -> tuple[float, float, float, float] | None:
    m = _VIEWBOX_RE.search(svg_text)
    if m:
        parts = _DIM_RE.findall(m.group(1))
        if len(parts) == 4:
            x, y, w, h = (float(p) for p in parts)
            if w > 0 and h > 0:
                return x, y, w, h
    # Fall back to width/height attributes.
    wm = re.search(r'\bwidth\s*=\s*"([^"]+)"', svg_text, re.IGNORECASE)
    hm = re.search(r'\bheight\s*=\s*"([^"]+)"', svg_text, re.IGNORECASE)
    if wm and hm:
        wp = _DIM_RE.search(wm.group(1))
        hp = _DIM_RE.search(hm.group(1))
        if wp and hp:
            w, h = float(wp.group()), float(hp.group())
            if w > 0 and h > 0:
                return 0.0, 0.0, w, h
    return None


def content_bbox(structure_svg: bytes) -> dict:
    """Measure the opaque content bounding box (a headless getBBox).

    Renders the structure SVG and measures the non-transparent pixel box, mapped
    back to user units. Falls back to the full viewBox when rendering yields
    nothing (e.g. embedded raster or gradient sources).
    """
    vb = parse_view_box(structure_svg.decode("utf-8", errors="replace"))
    if vb is None:
        vb = (0.0, 0.0, 100.0, 100.0)
    vbx, vby, vbw, vbh = vb
    full = {"x": vbx, "y": vby, "width": vbw, "height": vbh}

    render_w = min(round(vbw), MAX_RENDER_WIDTH)
    if render_w < 1:
        render_w = 1
    try:
        png = cairosvg.svg2png(bytestring=structure_svg, output_width=render_w)
        img = Image.open(BytesIO(png))
        box = img.getchannel("A").getbbox()
    except Exception as e:  # noqa: BLE001 — any render failure -> viewBox fallback
        print(f"  Warning: content bbox render failed ({e}); using viewBox", file=sys.stderr)
        return full

    if box is None:
        return full

    left, top, right, bottom = box
    sx = vbw / img.width
    sy = vbh / img.height
    return {
        "x": vbx + left * sx,
        "y": vby + top * sy,
        "width": (right - left) * sx,
        "height": (bottom - top) * sy,
    }


def dimensions(content_w: float, content_h: float) -> tuple[float, float, float, float]:
    """Aspect-derived physical size: longer side = LONG_SIDE_INCHES, min = default.

    Derived from the measured content bounding box (not the raw viewBox) so the
    physical tile aspect matches the artwork and the bake fits it tight rather
    than letterboxing it into an artboard-shaped tile.
    """
    w, h = content_w, content_h
    if w <= 0 or h <= 0:
        w = h = 1.0
    if w >= h:
        dw = LONG_SIDE_INCHES
        dh = round(LONG_SIDE_INCHES * h / w, 2)
    else:
        dh = LONG_SIDE_INCHES
        dw = round(LONG_SIDE_INCHES * w / h, 2)
    dw = max(dw, 0.01)
    dh = max(dh, 0.01)
    return dw, dh, dw, dh


# ---------------------------------------------------------------------------
# Progress
# ---------------------------------------------------------------------------

def load_progress() -> set[str]:
    if PROGRESS_FILE.exists():
        return set(json.loads(PROGRESS_FILE.read_text()))
    return set()


def save_progress(done: set[str]) -> None:
    PROGRESS_FILE.write_text(json.dumps(sorted(done), indent=2))


# ---------------------------------------------------------------------------
# Per-item pipeline
# ---------------------------------------------------------------------------

def process_item(svg_path: Path, glass_palette: list[dict], default_grout_id: int) -> bool:
    catalog_code = svg_path.stem
    json_path = svg_path.with_suffix(".json")

    if not json_path.exists():
        print(f"  Skipping {catalog_code}: no matching metadata JSON", file=sys.stderr)
        return False

    meta = json.loads(json_path.read_text())

    # 1. Upload the source SVG so analyze can fetch it from S3.
    source_bytes = svg_path.read_bytes()
    source_url = upload_bytes(svg_path.name, source_bytes, "image/svg+xml")
    if source_url is None:
        return False

    # 2. Analyze -> structure SVG + best-guess manifest.
    result = analyze(source_url)
    if result is None:
        return False

    structure_svg = result["structure_svg"]
    manifest = result["manifest"]
    for warning in result.get("warnings", []):
        print(f"  analyze: {warning}")

    # 3. Fill any unmatched color ids (the human editor's job).
    for note in fill_manifest(manifest, glass_palette, default_grout_id):
        print(f"  manifest: {note}")

    # 4. Measure the content bbox (the browser's getBBox).
    bbox = content_bbox(structure_svg.encode("utf-8"))

    # 5. Upload the structure SVG — this is what the create step bakes.
    structure_url = upload_bytes(
        f"{catalog_code}-structure.svg", structure_svg.encode("utf-8"), "image/svg+xml"
    )
    if structure_url is None:
        return False

    # 6. Aspect-derived physical dimensions, from the measured content aspect.
    dw, dh, mw, mh = dimensions(bbox["width"], bbox["height"])

    # 7. Create the catalog item (server bakes + swaps svg_url to the baked asset).
    body = {
        "catalog_code": catalog_code,
        "name": meta["name"],
        "description": meta.get("description"),
        "category": meta.get("category", svg_path.parent.name),
        "default_width": dw,
        "default_height": dh,
        "min_width": mw,
        "min_height": mh,
        "default_price_group_id": DEFAULT_PRICE_GROUP_ID,
        "svg_url": structure_url,
        "manifest": manifest,
        "content_bbox": bbox,
        "is_active": True,
        "tags": meta.get("tags", []),
    }

    ok, _ = create_catalog_item(body)
    if ok:
        print(f"  OK — {catalog_code} ({len(body['tags'])} tags, {dw}x{dh}in)")
    return ok


def main():
    if not AUTH_TOKEN:
        print("Error: AUTH_TOKEN environment variable is required.", file=sys.stderr)
        sys.exit(1)

    input_folder = Path(sys.argv[1]) if len(sys.argv) > 1 else Path(__file__).parent / "input"
    if not input_folder.exists() or not input_folder.is_dir():
        print(f"Error: '{input_folder}' is not a valid directory.", file=sys.stderr)
        sys.exit(1)

    try:
        glass_palette = fetch_palette("/api/glass-colors")
        grout_palette = fetch_palette("/api/grouts")
    except requests.RequestException as e:
        print(f"Error: failed to fetch color palettes — {e}", file=sys.stderr)
        sys.exit(1)

    if not glass_palette:
        print("Error: no active glass colors found. Run `pnpm db:seed` first.", file=sys.stderr)
        sys.exit(1)
    if not grout_palette:
        print("Error: no active grouts found. Run `pnpm db:seed` first.", file=sys.stderr)
        sys.exit(1)

    default_grout_id = nearest_id(BLACK, grout_palette) or grout_palette[0]["id"]

    svg_files = sorted(input_folder.rglob("*.svg"))
    total = len(svg_files)
    if total == 0:
        print(f"No SVG files found in '{input_folder}'.")
        sys.exit(0)

    done = load_progress()
    skipped = 0
    errors = 0

    print(f"Found {total} SVG files. {len(done)} already seeded.")

    for i, svg_path in enumerate(svg_files, 1):
        catalog_code = svg_path.stem
        print(f"[{i}/{total}] {svg_path.relative_to(input_folder)}")

        if catalog_code in done:
            print(f"  Skipping {catalog_code} (already seeded)")
            skipped += 1
            continue

        if process_item(svg_path, glass_palette, default_grout_id):
            done.add(catalog_code)
            save_progress(done)
        else:
            errors += 1

    succeeded = total - skipped - errors
    print(f"\nDone. {succeeded} seeded, {skipped} skipped, {errors} errors.")
    if errors:
        sys.exit(1)


if __name__ == "__main__":
    main()
