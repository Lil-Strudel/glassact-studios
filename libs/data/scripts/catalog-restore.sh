#!/usr/bin/env bash
#
# Restores the catalog tables from libs/data/catalog_seed.sql into the DSN
# configured in libs/data/.env.
#
# The seed file inserts with "ON CONFLICT DO NOTHING", so this is non-destructive:
# existing rows are left untouched and only missing catalog items/tags are added.
# Requires the referenced price_groups to already exist (they ship in seed.sql).
#
set -euo pipefail

script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
data_dir="$(dirname "$script_dir")"

# shellcheck disable=SC1091
source "$data_dir/.env"

seed_file="$data_dir/catalog_seed.sql"

if [[ ! -f "$seed_file" ]]; then
    echo "No catalog seed found at $seed_file — run 'pnpm db:catalog/dump' first." >&2
    exit 1
fi

psql "$DATABASE_DSN" --set=ON_ERROR_STOP=1 -f "$seed_file"

echo "Catalog restored from $seed_file"
