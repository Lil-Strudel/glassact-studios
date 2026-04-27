import json
import sys
import textwrap
from pathlib import Path

import cairosvg
import ollama

MODEL = "qwen3-vl:8b"

PROMPT_TEMPLATE = textwrap.dedent("""
    You are a catalog metadata specialist for a stained glass inlay company.
    You are analyzing catalog item {catalog_code} in category {category}.
    Analyze this stained glass design image and respond with ONLY a JSON object — no markdown, no explanation, no code fences.

    These designs are used exclusively as stained glass inlays for memorial gravestones — not home decor or interior design.
    All designs are clipart-style stained glass artwork, so the art style is always minimalistic by default.

    The JSON must have exactly these fields:
    - "name": a short, descriptive catalog item name (2-4 words). Do NOT use the word "inlay".
    - "description": 1-2 sentences describing only what the scene depicts — the subject, composition, and any notable elements.
      DO NOT mention stained glass, the art style, clean lines, minimalism, or color quality — these are assumed.
      DO NOT mention use context, atmosphere, or interior suitability — these are memorial products.
      You MAY mention style or tone only if the design is clearly tailored to a specific demographic (e.g. children, military veterans).
    - "tags": an array of search-index-friendly strings covering subject matter, motif, occasion, and demographic where relevant.
      DO NOT include tags for "stained glass", "minimalist", "clipart", or any home/interior use context — these apply to every item.

    Example format:
    {{"name": "Hummingbird in Flight", "description": "A hummingbird hovering beside a cluster of tropical flowers.", "tags": ["bird", "hummingbird", "nature", "floral", "tropical", "wildlife"]}}
""").strip()


def process_svg(svg_path: Path) -> bool:
    catalog_code = svg_path.stem
    category = svg_path.parent.name
    json_path = svg_path.with_suffix(".json")

    if json_path.exists():
        print(f"  Skipping {svg_path.name} (already processed)")
        return True

    print(f"  Converting {svg_path.name} to PNG...")
    try:
        png_bytes = cairosvg.svg2png(url=str(svg_path))
    except Exception as e:
        print(f"  Error: Failed to convert {svg_path.name} — {e}", file=sys.stderr)
        return False

    prompt = PROMPT_TEMPLATE.format(catalog_code=catalog_code, category=category)

    print(f"  Querying {MODEL}...")
    try:
        response = ollama.chat(
            model=MODEL,
            messages=[
                {
                    "role": "user",
                    "content": prompt,
                    "images": [png_bytes],
                }
            ],
        )
    except Exception as e:
        print(
            f"  Error: Could not reach Ollama for {svg_path.name} — {e}",
            file=sys.stderr,
        )
        return False

    raw = response.message.content.strip()

    try:
        result = json.loads(raw)
    except json.JSONDecodeError:
        print(
            f"  Error: Model response was not valid JSON for {svg_path.name}:",
            file=sys.stderr,
        )
        print(f"  {raw}", file=sys.stderr)
        return False

    if not isinstance(result.get("name"), str):
        print(
            f"  Error: Response missing string field 'name' for {svg_path.name}.",
            file=sys.stderr,
        )
        return False
    if not isinstance(result.get("description"), str):
        print(
            f"  Error: Response missing string field 'description' for {svg_path.name}.",
            file=sys.stderr,
        )
        return False
    if not isinstance(result.get("tags"), list) or not all(
        isinstance(t, str) for t in result["tags"]
    ):
        print(
            f"  Error: Response missing 'tags' array of strings for {svg_path.name}.",
            file=sys.stderr,
        )
        return False

    output = {
        "catalog_code": catalog_code,
        "category": category,
        "name": result["name"],
        "description": result["description"],
        "tags": result["tags"],
    }

    json_path.write_text(json.dumps(output, indent=2))
    print(f"  Written to {json_path.name}")
    return True


def main():
    input_folder = sys.argv[1] if len(sys.argv) > 1 else "./input"
    folder = Path(input_folder)

    if not folder.exists() or not folder.is_dir():
        print(f"Error: '{input_folder}' is not a valid directory.", file=sys.stderr)
        sys.exit(1)

    svgs = sorted(folder.rglob("*.svg"))
    total = len(svgs)

    if total == 0:
        print(f"No SVG files found in '{input_folder}'.")
        sys.exit(0)

    print(f"Found {total} SVG files in '{input_folder}'.")
    errors = 0

    for i, svg_path in enumerate(svgs, 1):
        print(f"[{i}/{total}] {svg_path.relative_to(folder)}")
        if not process_svg(svg_path):
            errors += 1

    print(f"\nDone. {total - errors}/{total} succeeded, {errors} errors.")
    if errors:
        sys.exit(1)


if __name__ == "__main__":
    main()
