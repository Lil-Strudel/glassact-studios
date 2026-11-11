#!/usr/bin/env bash
#
# Usage: catalog-restore.sh [--replace] [backup-dir]
#
# Loads the catalog snapshot (catalog_seed.sql, or <backup-dir>/catalog.sql)
# into the DSN from libs/data/.env. Refuses to run over existing catalog rows
# unless --replace is given, because the snapshot carries fixed ids.
#
set -euo pipefail

script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
data_dir="$(dirname "$script_dir")"

if [[ -z "${DATABASE_DSN:-}" ]]; then
    # shellcheck disable=SC1091
    source "$data_dir/.env"
fi

replace=false
if [[ "${1:-}" == "--replace" ]]; then
    replace=true
    shift
fi

if [[ -n "${1:-}" ]]; then
    seed_file="$1/catalog.sql"
else
    seed_file="$data_dir/catalog_seed.sql"
fi

if [[ ! -f "$seed_file" ]]; then
    echo "No catalog snapshot at $seed_file — run 'pnpm db:catalog/backup' first." >&2
    exit 1
fi

catalog_tables="price_groups, glass_colors, grouts, catalog_items, catalog_item_tags"

existing="$(psql "$DATABASE_DSN" -Atc "SELECT
    (SELECT count(*) FROM price_groups) + (SELECT count(*) FROM glass_colors) +
    (SELECT count(*) FROM grouts) + (SELECT count(*) FROM catalog_items)")"

truncate_sql=""
if ((existing > 0)); then
    if [[ "$replace" != true ]]; then
        echo "The catalog tables already have rows. Re-run with --replace to wipe and reload them." >&2
        exit 1
    fi
    echo "--replace truncates $catalog_tables with CASCADE."
    echo "That also deletes every local project, inlay, proof and invoice that references them."
    read -r -p "Type 'replace' to continue: " answer
    if [[ "$answer" != "replace" ]]; then
        echo "Aborted." >&2
        exit 1
    fi
    truncate_sql="TRUNCATE $catalog_tables RESTART IDENTITY CASCADE;"
fi

# Prod's price_groups sequence was never advanced past its explicitly-id'd
# rows, so the dump's setval would leave the next insert colliding with id 1.
{
    echo "BEGIN;"
    echo "$truncate_sql"
    cat "$seed_file"
    for table in price_groups glass_colors grouts catalog_items catalog_item_tags; do
        echo "SELECT setval(pg_get_serial_sequence('public.$table', 'id'), GREATEST((SELECT max(id) FROM public.$table), 1));"
    done
    echo "COMMIT;"
} | psql "$DATABASE_DSN" --quiet --set=ON_ERROR_STOP=1 -o /dev/null

echo "Catalog restored from $seed_file"
