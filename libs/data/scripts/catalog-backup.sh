#!/usr/bin/env bash
#
# Price groups, glass colors and grouts ride along because catalog manifests
# reference them by integer id.
#
# Prod Postgres only listens on the instance's loopback. Rather than require
# the SSM port-forwarding plugin, the dump runs on the box via send-command and
# comes back through the backups bucket, which the instance role can already
# write under postgres/.
#
set -euo pipefail

profile="${AWS_PROFILE:-glassact-studios}"
region="${AWS_REGION:-us-west-2}"
files_bucket="${FILES_BUCKET:-glassact-nkm52j}"

script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
data_dir="$(dirname "$script_dir")"

aws_() { aws --profile "$profile" --region "$region" "$@"; }

stamp="$(date -u +%Y%m%dT%H%M%SZ)"
backup_dir="$data_dir/backups/$(date -u +%Y-%m-%d)"
mkdir -p "$backup_dir/files"

instance_id="$(aws_ ec2 describe-instances \
    --filters "Name=tag:Name,Values=glassact-api" "Name=instance-state-name,Values=running" \
    --query 'Reservations[0].Instances[0].InstanceId' --output text)"
backups_bucket="$(aws_ s3api list-buckets \
    --query "Buckets[?starts_with(Name, 'glassact-backups-')].Name | [0]" --output text)"
dump_key="postgres/catalog-$stamp.sql"

echo "Dumping catalog on $instance_id to s3://$backups_bucket/$dump_key"

remote_cmd="set -o pipefail; docker exec glassact-postgres-1 pg_dump -U glassact -d glassact \
--data-only --no-owner --no-privileges --column-inserts \
-t price_groups -t glass_colors -t grouts -t catalog_items -t catalog_item_tags \
| aws s3 cp - s3://$backups_bucket/$dump_key"

params="$(python3 -c 'import json,sys; print(json.dumps({"commands": ["bash -c " + json.dumps(sys.argv[1])]}))' "$remote_cmd")"

command_id="$(aws_ ssm send-command \
    --instance-ids "$instance_id" \
    --document-name AWS-RunShellScript \
    --parameters "$params" \
    --query Command.CommandId --output text)"

while true; do
    status="$(aws_ ssm get-command-invocation --instance-id "$instance_id" --command-id "$command_id" \
        --query Status --output text 2>/dev/null || echo Pending)"
    case "$status" in
    Success) break ;;
    Pending | InProgress | Delayed) sleep 2 ;;
    *)
        aws_ ssm get-command-invocation --instance-id "$instance_id" --command-id "$command_id" \
            --query StandardErrorContent --output text >&2
        echo "Remote dump failed with status $status" >&2
        exit 1
        ;;
    esac
done

# pg_dump 18 brackets the file in \restrict / \unrestrict, which older psql
# clients reject and which carry a per-dump random key we don't want in git.
aws_ s3 cp "s3://$backups_bucket/$dump_key" - |
    grep -v -E '^\\(un)?restrict ' >"$backup_dir/catalog.sql"

cp "$backup_dir/catalog.sql" "$data_dir/catalog_seed.sql"

grep -o "'/file/catalog-items/[^']*'" "$backup_dir/catalog.sql" | tr -d "'" | sort -u |
    sed 's|^/||' >"$backup_dir/file-keys.txt"

key_count="$(wc -l <"$backup_dir/file-keys.txt")"
echo "Downloading $key_count catalog SVGs"

export profile region files_bucket backup_dir
xargs -P 8 -I{} sh -c \
    'aws --profile "$profile" --region "$region" s3 cp --only-show-errors "s3://$files_bucket/$1" "$backup_dir/files/$1"' _ {} \
    <"$backup_dir/file-keys.txt"

missing=0
while read -r key; do
    if [[ ! -s "$backup_dir/files/$key" ]]; then
        echo "Missing: $key" >&2
        missing=$((missing + 1))
    fi
done <"$backup_dir/file-keys.txt"
if ((missing > 0)); then
    echo "$missing referenced files could not be downloaded" >&2
    exit 1
fi

(cd "$backup_dir" && find files -type f -print0 | sort -z | xargs -0 sha256sum catalog.sql >MANIFEST.sha256)

echo "Backup written to $backup_dir"
echo "Seed refreshed at $data_dir/catalog_seed.sql"
