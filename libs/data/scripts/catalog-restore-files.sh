#!/usr/bin/env bash
#
# Usage: catalog-restore-files.sh <backup-dir> <bucket>
#
# Uploads a backup's catalog SVGs to <bucket>, keeping their file/ keys so the
# svg_url values in catalog.sql resolve. The bucket is deliberately required:
# dev shares the prod bucket, so there is no safe default.
#
set -euo pipefail

if [[ $# -ne 2 ]]; then
    echo "Usage: $0 <backup-dir> <bucket>" >&2
    exit 1
fi

backup_dir="$1"
bucket="$2"

(cd "$backup_dir" && sha256sum -c --quiet MANIFEST.sha256)

aws --profile "${AWS_PROFILE:-glassact-studios}" --region "${AWS_REGION:-us-west-2}" \
    s3 sync --size-only "$backup_dir/files/file" "s3://$bucket/file"

echo "Restored $(find "$backup_dir/files" -type f | wc -l) files to s3://$bucket"
