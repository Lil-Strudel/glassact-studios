import json
import os
import sys
from pathlib import Path

import requests

API_BASE = os.environ.get("API_BASE", "http://localhost:4100")
AUTH_TOKEN = os.environ.get("AUTH_TOKEN", "")

PROGRESS_FILE = Path(__file__).parent / "seed_progress.json"


def load_progress() -> set[str]:
    if PROGRESS_FILE.exists():
        return set(json.loads(PROGRESS_FILE.read_text()))
    return set()


def save_progress(done: set[str]) -> None:
    PROGRESS_FILE.write_text(json.dumps(sorted(done), indent=2))


def headers() -> dict:
    return {"Authorization": f"Bearer {AUTH_TOKEN}"}


def upload_svg(svg_path: Path) -> str | None:
    with svg_path.open("rb") as f:
        resp = requests.post(
            f"{API_BASE}/api/upload",
            headers=headers(),
            files={"file": (svg_path.name, f, "image/svg+xml")},
        )
    if not resp.ok:
        print(f"  Error uploading {svg_path.name}: {resp.status_code} {resp.text}", file=sys.stderr)
        return None
    return resp.json()["url"]


def create_catalog_item(data: dict, svg_url: str) -> str | None:
    body = {
        "catalog_code": data["catalog_code"],
        "category": data["category"],
        "name": data["name"],
        "description": data.get("description"),
        "svg_url": svg_url,
        "default_width": 1,
        "default_height": 1,
        "min_width": 1,
        "min_height": 1,
        "default_price_group_id": 1,
        "is_active": True,
    }
    resp = requests.post(
        f"{API_BASE}/api/catalog",
        headers={**headers(), "Content-Type": "application/json"},
        json=body,
    )
    if not resp.ok:
        print(f"  Error creating catalog item {data['catalog_code']}: {resp.status_code} {resp.text}", file=sys.stderr)
        return None
    return resp.json()["uuid"]


def add_tags(uuid: str, tags: list[str]) -> bool:
    ok = True
    for tag in tags:
        resp = requests.post(
            f"{API_BASE}/api/catalog/{uuid}/tags",
            headers={**headers(), "Content-Type": "application/json"},
            json={"tag": tag},
        )
        if not resp.ok:
            print(f"  Error adding tag '{tag}' to {uuid}: {resp.status_code} {resp.text}", file=sys.stderr)
            ok = False
    return ok


def process_item(json_path: Path) -> bool:
    data = json.loads(json_path.read_text())
    catalog_code = data["catalog_code"]
    svg_path = json_path.with_suffix(".svg")

    if not svg_path.exists():
        print(f"  Skipping {catalog_code}: no matching SVG found", file=sys.stderr)
        return False

    svg_url = upload_svg(svg_path)
    if svg_url is None:
        return False

    uuid = create_catalog_item(data, svg_url)
    if uuid is None:
        return False

    tags_ok = add_tags(uuid, data.get("tags", []))

    if tags_ok:
        print(f"  OK — {catalog_code} ({len(data.get('tags', []))} tags)")
    else:
        print(f"  Partial — {catalog_code} created but some tags failed")

    return True


def main():
    if not AUTH_TOKEN:
        print("Error: AUTH_TOKEN environment variable is required.", file=sys.stderr)
        sys.exit(1)

    input_folder = Path(sys.argv[1]) if len(sys.argv) > 1 else Path(__file__).parent / "input"

    if not input_folder.exists() or not input_folder.is_dir():
        print(f"Error: '{input_folder}' is not a valid directory.", file=sys.stderr)
        sys.exit(1)

    json_files = sorted(input_folder.rglob("*.json"))
    total = len(json_files)

    if total == 0:
        print(f"No JSON files found in '{input_folder}'.")
        sys.exit(0)

    done = load_progress()
    skipped = 0
    errors = 0

    print(f"Found {total} JSON files. {len(done)} already seeded.")

    for i, json_path in enumerate(json_files, 1):
        catalog_code = json_path.stem
        print(f"[{i}/{total}] {json_path.relative_to(input_folder)}")

        if catalog_code in done:
            print(f"  Skipping {catalog_code} (already seeded)")
            skipped += 1
            continue

        if process_item(json_path):
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
