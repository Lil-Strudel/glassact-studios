#!/usr/bin/env bash
set -euo pipefail

: "${BACKUP_BUCKET:?BACKUP_BUCKET must be set}"

TIMESTAMP=$(date -u +%Y%m%dT%H%M%SZ)
TMP_FILE="/tmp/glassact-${TIMESTAMP}.dump.gz"

docker exec glassact-postgres-1 pg_dump -U glassact -d glassact -F c | gzip > "${TMP_FILE}"

aws s3 cp "${TMP_FILE}" "s3://${BACKUP_BUCKET}/postgres/${TIMESTAMP}.dump.gz"
rm -f "${TMP_FILE}"
