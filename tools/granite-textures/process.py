import argparse
import json
import sys
from dataclasses import dataclass
from io import BytesIO
from pathlib import Path

import numpy as np
from PIL import Image
from pypdf import PdfReader

TOOL_DIR = Path(__file__).parent
INPUT_DIR = TOOL_DIR / "input"
TINT_SOURCE = TOOL_DIR / "tint-source" / "granite-slab.jpg"
OUTPUT_DIR = TOOL_DIR.parent.parent / "apps" / "webapp" / "src" / "assets" / "images" / "granite"

PDF_POINTS_PER_INCH = 72.0
WEBP_QUALITY = 82
SWATCH_PIXELS = 96
SWATCH_INCHES = 3.0
# The CC0 slab has no known physical size; 6" makes its grain read like a
# medium-grained monument granite next to the real samples.
TINT_TILE_INCHES = 6.0
# Below this OKLab chroma a stone reads as black/gray rather than colored.
NEUTRAL_CHROMA = 0.012


@dataclass
class Sample:
    key: str
    name: str
    tile_inches: float
    jpeg: bytes
    is_gray_gamma: bool


def extract_sample(pdf_path: Path) -> Sample:
    """Pull the tiled JPEG and its physical size out of a sample PDF.

    Each sample PDF is a 24"x12" page painted with a tiling /Pattern whose single
    image XObject is the stone photo. Taking the JPEG bytes directly avoids
    rasterizing the page (and resampling the photo) entirely.
    """
    page = PdfReader(pdf_path).pages[0]
    patterns = page["/Resources"]["/Pattern"]
    if len(patterns) != 1:
        raise ValueError(f"{pdf_path.name}: expected 1 pattern, found {len(patterns)}")
    pattern = next(iter(patterns.values())).get_object()

    images = [x.get_object() for x in pattern["/Resources"]["/XObject"].values()]
    if len(images) != 1:
        raise ValueError(f"{pdf_path.name}: expected 1 image, found {len(images)}")
    image = images[0]

    filters = image["/Filter"]
    filters = list(filters) if isinstance(filters, list) else [filters]
    if filters != ["/DCTDecode"]:
        raise ValueError(f"{pdf_path.name}: expected a bare JPEG, got filters {filters}")

    color_space = image["/ColorSpace"]
    profile = color_space[1].get_object().get_data()
    is_gray_gamma = b"Gray Gamma 2.2" in profile
    if not is_gray_gamma and b"sRGB" not in profile:
        raise ValueError(f"{pdf_path.name}: image is neither sRGB nor Gray Gamma 2.2")

    bbox_width = float(pattern["/BBox"][2]) - float(pattern["/BBox"][0])
    if abs(float(pattern["/XStep"]) - bbox_width) > 1e-3:
        raise ValueError(f"{pdf_path.name}: pattern step does not match its tile")
    scale = float(pattern["/Matrix"][0])
    tile_inches = round(bbox_width * scale / PDF_POINTS_PER_INCH, 3)

    jpeg = image.get_data()
    if jpeg[:2] != b"\xff\xd8":
        raise ValueError(f"{pdf_path.name}: image stream is not a JPEG")

    name = pdf_path.stem.removesuffix(" 2x1")
    return Sample(
        key=name.lower().replace(" ", "-"),
        name=name,
        tile_inches=tile_inches,
        jpeg=jpeg,
        is_gray_gamma=is_gray_gamma,
    )


def srgb_to_linear(srgb: np.ndarray) -> np.ndarray:
    return np.where(srgb <= 0.04045, srgb / 12.92, ((srgb + 0.055) / 1.055) ** 2.4)


def linear_to_srgb(linear: np.ndarray) -> np.ndarray:
    linear = np.clip(linear, 0.0, 1.0)
    return np.where(
        linear <= 0.0031308, linear * 12.92, 1.055 * linear ** (1 / 2.4) - 0.055
    )


def decode_linear(sample: Sample) -> np.ndarray:
    image = Image.open(BytesIO(sample.jpeg))
    if image.mode not in ("L", "RGB"):
        raise ValueError(f"{sample.name}: unsupported JPEG mode {image.mode}")
    encoded = np.asarray(image.convert("RGB"), dtype=np.float64) / 255.0
    if sample.is_gray_gamma:
        return encoded**2.2
    return srgb_to_linear(encoded)


def to_srgb8(linear: np.ndarray) -> np.ndarray:
    return np.round(linear_to_srgb(linear) * 255.0).astype(np.uint8)


def luminance(linear: np.ndarray) -> np.ndarray:
    return linear @ np.array([0.2126, 0.7152, 0.0722])


def gaussian_kernel(sigma: float) -> np.ndarray:
    radius = int(np.ceil(3 * sigma))
    x = np.arange(-radius, radius + 1)
    kernel = np.exp(-(x**2) / (2 * sigma**2))
    return kernel / kernel.sum()


def blur_1d(signal: np.ndarray, sigma: float, mode: str) -> np.ndarray:
    kernel = gaussian_kernel(sigma)
    return np.convolve(np.pad(signal, len(kernel) // 2, mode=mode), kernel, mode="valid")


def blur(plane: np.ndarray, sigma: float, mode: str) -> np.ndarray:
    rows = np.apply_along_axis(blur_1d, 1, plane, sigma, mode)
    return np.apply_along_axis(blur_1d, 0, rows, sigma, mode)


def destripe(linear: np.ndarray) -> np.ndarray:
    """Even out brightness that is constant along whole rows or columns.

    Some samples (Galaxy Black most of all) carry scanner striping: bands a few
    dozen pixels wide that are uniformly darker top to bottom. Stone has no
    structure that runs the full height of a 12" tile, so a row or column
    average that departs from the image mean is an artifact. Tiling would also
    repeat each band every tile.
    """
    lum = luminance(linear)
    target = lum.mean()
    for axis in (0, 1):
        profile = blur_1d(lum.mean(axis=axis), sigma=3.0, mode="reflect")
        gain = np.clip(target / profile, 0.5, 2.0)
        shape = (1, -1, 1) if axis == 0 else (-1, 1, 1)
        linear = linear * gain.reshape(shape)
        lum = luminance(linear)
    return linear


def flatten_lighting(linear: np.ndarray) -> tuple[np.ndarray, float, float]:
    """Divide out the photo's low-frequency lighting (vignetting, gradients).

    Once tiled, any brightness drift across the photo repeats as a visible
    light/dark checkerboard even when the edges themselves match. Only the
    luminance gain is touched, so the stone's color is untouched; sigma is a
    quarter tile, far coarser than any crystal.
    """
    height, width = linear.shape[:2]
    grid = 32
    block_y, block_x = height // grid, width // grid
    lum = luminance(linear)
    coarse = (
        lum[: block_y * grid, : block_x * grid]
        .reshape(grid, block_y, grid, block_x)
        .mean(axis=(1, 3))
    )
    coarse = blur(coarse, sigma=grid / 4, mode="reflect")
    lighting = np.asarray(
        Image.fromarray(coarse.astype(np.float32)).resize(
            (width, height), Image.Resampling.BICUBIC
        ),
        dtype=np.float64,
    )
    gain = np.clip(lum.mean() / lighting, 0.5, 2.0)
    return linear * gain[..., None], float(gain.min()), float(gain.max())


def periodic_min_cut(cost: np.ndarray) -> np.ndarray:
    """Minimum-cost top-to-bottom path through `cost` that ends where it starts.

    Returns the column of the path in every row. The path must wrap cleanly
    because the tile repeats vertically, so a jog between the last and first
    row would itself be a seam. All start columns are solved at once.
    """
    rows, cols = cost.shape
    starts = np.arange(cols)
    total = np.full((cols, cols), np.inf)
    total[starts, starts] = cost[0]
    moves = np.zeros((rows, cols, cols), dtype=np.int8)
    for row in range(1, rows):
        left = np.concatenate([np.full((cols, 1), np.inf), total[:, :-1]], axis=1)
        right = np.concatenate([total[:, 1:], np.full((cols, 1), np.inf)], axis=1)
        stacked = np.stack([left, total, right])
        choice = stacked.argmin(axis=0)
        total = stacked.min(axis=0) + cost[row]
        moves[row] = choice - 1

    best_start = int(np.argmin(total[starts, starts]))
    path = np.empty(rows, dtype=int)
    path[-1] = best_start
    for row in range(rows - 1, 0, -1):
        path[row - 1] = path[row] + moves[row, best_start, path[row]]
    return path


def heal_vertical_seam(linear: np.ndarray) -> np.ndarray:
    """Hide the seam where the tile's right edge meets its left edge.

    The same image shifted half a tile sideways is continuous across that edge.
    A band around the seam is swapped for the shifted copy, bounded by
    minimum-error cuts (image quilting) so the switch-over follows the grain
    instead of a straight line, then lightly feathered. Unlike a cross-fade,
    nothing is averaged, so there is no ghosting or loss of contrast.
    """
    height, width = linear.shape[:2]
    half_band = width // 16
    guard = 3

    centered = np.roll(linear, half_band, axis=1)
    shifted = np.roll(centered, width // 2, axis=1)
    band = slice(0, 2 * half_band)
    # Scored in sRGB, not linear light: in linear light every difference between
    # near-black pixels is tiny, so cuts would wander through dark artifacts.
    error = (
        (linear_to_srgb(centered[:, band]) - linear_to_srgb(shifted[:, band])) ** 2
    ).sum(axis=2)

    left_cut = periodic_min_cut(error[:, : half_band - guard])
    right_cut = periodic_min_cut(error[:, half_band + guard :]) + half_band + guard

    columns = np.arange(width)[None, :]
    use_shifted = (columns >= left_cut[:, None]) & (columns <= right_cut[:, None])
    weight = blur(use_shifted.astype(np.float64), sigma=1.0, mode="wrap")[..., None]
    healed = centered * (1 - weight) + shifted * weight
    return np.roll(healed, -half_band, axis=1)


def make_seamless(linear: np.ndarray) -> np.ndarray:
    # Healing the vertical seam first leaves a single horizontal seam that is
    # straight across the whole tile, so the transposed pass can heal it with a
    # source that is already seamless side to side.
    healed = heal_vertical_seam(linear)
    return heal_vertical_seam(healed.transpose(1, 0, 2)).transpose(1, 0, 2)


def seam_ratio(srgb8: np.ndarray) -> float:
    """Wrap-edge pixel difference relative to the tile's interior; ~1.0 is invisible.

    The baseline is the difference across interior 8px JPEG block edges, not
    across any neighboring pair: the wrap edge always falls on a block edge,
    and in the heavily compressed black samples block edges alone differ
    noticeably more than pixels inside a block.
    """
    pixels = srgb8.astype(np.float64)
    edge = (
        np.abs(pixels[:, -1] - pixels[:, 0]).mean()
        + np.abs(pixels[-1, :] - pixels[0, :]).mean()
    )
    block_edges = np.arange(8, pixels.shape[0], 8)
    interior = (
        np.abs(pixels[:, block_edges] - pixels[:, block_edges - 1]).mean()
        + np.abs(pixels[block_edges, :] - pixels[block_edges - 1, :]).mean()
    )
    return float(edge / interior)


def resize_linear(linear: np.ndarray, size: int) -> np.ndarray:
    channels = [
        np.asarray(
            Image.fromarray(linear[..., c].astype(np.float32)).resize(
                (size, size), Image.Resampling.LANCZOS
            ),
            dtype=np.float64,
        )
        for c in range(3)
    ]
    return np.stack(channels, axis=2)


def swatch(linear: np.ndarray, tile_inches: float) -> np.ndarray:
    width = linear.shape[1]
    crop = min(width, round(SWATCH_INCHES * width / tile_inches))
    start = (width - crop) // 2
    patch = linear[start : start + crop, start : start + crop]
    return resize_linear(patch, SWATCH_PIXELS)


def mean_hex(linear: np.ndarray) -> str:
    srgb8 = to_srgb8(linear.reshape(-1, 3).mean(axis=0))
    return "#" + "".join(f"{v:02x}" for v in srgb8)


def oklch(linear_rgb: np.ndarray) -> tuple[float, float, float]:
    lms = (
        np.array(
            [
                [0.4122214708, 0.5363325363, 0.0514459929],
                [0.2119034982, 0.6806995451, 0.1073969566],
                [0.0883024619, 0.2817188376, 0.6299787005],
            ]
        )
        @ linear_rgb
    )
    l_, m_, s_ = np.cbrt(lms)
    lightness = 0.2104542553 * l_ + 0.7936177850 * m_ - 0.0040720468 * s_
    a = 1.9779984951 * l_ - 2.4285922050 * m_ + 0.4505937099 * s_
    b = 0.0259040371 * l_ + 0.7827717662 * m_ - 0.8086757660 * s_
    return float(lightness), float(np.hypot(a, b)), float(np.degrees(np.arctan2(b, a)) % 360)


def picker_order(entry: dict) -> tuple:
    lightness, chroma, hue = entry["_oklch"]
    if chroma < NEUTRAL_CHROMA:
        return (0, 0.0, lightness)
    # Start the hue wheel at olive so greens, blues, then reds/pinks run in one
    # sweep instead of the greens being split across 0/360.
    return (1, (hue - 100) % 360, lightness)


def save_webp(srgb8: np.ndarray, path: Path) -> int:
    Image.fromarray(srgb8).save(
        path, "WEBP", quality=WEBP_QUALITY, method=6, exact=False
    )
    return path.stat().st_size


def contact_sheet(srgb8: np.ndarray, path: Path) -> None:
    tiled = np.tile(srgb8, (3, 3, 1))
    sheet = Image.fromarray(tiled)
    sheet.thumbnail((1800, 1800), Image.Resampling.LANCZOS)
    sheet.save(path, "JPEG", quality=90)


def process_tint_base(review_dir: Path | None) -> int:
    """Neutral tile for the custom-color backdrop, centered on 50% gray.

    The webapp composites it over the picked color with `overlay`, which leaves
    a 50% gray pixel showing exactly the picked color, so the backdrop averages
    to what the user chose rather than a darker version of it.
    """
    gray = np.asarray(Image.open(TINT_SOURCE).convert("L"), dtype=np.float64) / 255.0
    centered = np.clip(gray - gray.mean() + 0.5, 0.0, 1.0)
    srgb8 = np.round(centered * 255.0).astype(np.uint8)
    rgb = np.repeat(srgb8[..., None], 3, axis=2)
    if review_dir:
        contact_sheet(rgb, review_dir / "tint-base.jpg")
    return save_webp(rgb, OUTPUT_DIR / "tint-base.webp")


def main() -> None:
    parser = argparse.ArgumentParser(
        description="Generate the webapp's granite backdrop tiles from the sample PDFs."
    )
    parser.add_argument(
        "--review",
        type=Path,
        help="write a 3x3 tiled contact sheet per texture here for eyeballing seams",
    )
    args = parser.parse_args()

    pdfs = sorted(INPUT_DIR.glob("*.pdf"))
    if not pdfs:
        sys.exit(f"No sample PDFs in {INPUT_DIR}")
    OUTPUT_DIR.mkdir(parents=True, exist_ok=True)
    if args.review:
        args.review.mkdir(parents=True, exist_ok=True)

    entries = []
    total_bytes = 0
    print(f"{'granite':<16} {'tile':>5} {'px':>5} {'gain':>11} {'seam before':>12} {'after':>6} {'size':>7}")
    for pdf in pdfs:
        sample = extract_sample(pdf)
        original = decode_linear(sample)
        flattened, gain_min, gain_max = flatten_lighting(destripe(original))
        seamless = make_seamless(flattened)
        srgb8 = to_srgb8(seamless)

        before = seam_ratio(to_srgb8(original))
        after = seam_ratio(srgb8)
        tile_bytes = save_webp(srgb8, OUTPUT_DIR / f"{sample.key}.webp")
        swatch_bytes = save_webp(
            to_srgb8(swatch(seamless, sample.tile_inches)),
            OUTPUT_DIR / f"{sample.key}-swatch.webp",
        )
        total_bytes += tile_bytes + swatch_bytes
        if args.review:
            contact_sheet(to_srgb8(original), args.review / f"{sample.key}-before.jpg")
            contact_sheet(srgb8, args.review / f"{sample.key}.jpg")

        print(
            f"{sample.key:<16} {sample.tile_inches:>4g}\" {original.shape[1]:>5} "
            f"{gain_min:>5.2f}-{gain_max:<5.2f} {before:>12.2f} {after:>6.2f} "
            f"{tile_bytes / 1024:>6.0f}K"
        )
        entries.append(
            {
                "key": sample.key,
                "name": sample.name,
                "tileInches": sample.tile_inches,
                "meanHex": mean_hex(seamless),
                "_oklch": oklch(seamless.reshape(-1, 3).mean(axis=0)),
            }
        )

    entries.sort(key=picker_order)
    for entry in entries:
        del entry["_oklch"]
    manifest = {
        "tint": {"tileInches": TINT_TILE_INCHES},
        "samples": entries,
    }
    (OUTPUT_DIR / "granites.json").write_text(json.dumps(manifest, indent=2) + "\n")

    total_bytes += process_tint_base(args.review)
    print(f"\n{len(entries)} granites, {total_bytes / 1024 / 1024:.1f} MB written to {OUTPUT_DIR}")


if __name__ == "__main__":
    main()
