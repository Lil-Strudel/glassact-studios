#!/usr/bin/env bash
#
# Dumps the catalog tables (catalog_items + catalog_item_tags) from the DSN
# configured in libs/data/.env into libs/data/catalog_seed.sql.
#
# The dump is data-only and uses column-qualified INSERTs with
# "ON CONFLICT DO NOTHING", so the resulting file can be re-applied to any dev
# database without clobbering existing rows. Sequence values are included so a
# fresh restore keeps the same ids.
#
set -euo pipefail

script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
data_dir="$(dirname "$script_dir")"

# shellcheck disable=SC1091
source "$data_dir/.env"

out_file="$data_dir/catalog_seed.sql"

pg_dump "$DATABASE_DSN" \
    --data-only \
    --no-owner \
    --no-privileges \
    --column-inserts \
    --on-conflict-do-nothing \
    --table=catalog_items \
    --table=catalog_item_tags \
    >"$out_file"

echo "Catalog dumped to $out_file"
