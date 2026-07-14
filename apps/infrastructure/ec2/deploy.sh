#!/usr/bin/env bash
# Run on the instance via SSM Run Command:
#   aws s3 cp s3://<deploy-bucket>/deploy/deploy.sh - | bash -s -- "$API_IMAGE" "$MIGRATE_IMAGE" "$DEPLOY_BUCKET"
set -euo pipefail

API_IMAGE="$1"
MIGRATE_IMAGE="$2"
DEPLOY_BUCKET="$3"

mkdir -p /opt/glassact
aws s3 cp "s3://${DEPLOY_BUCKET}/deploy/docker-compose.yml" /opt/glassact/docker-compose.yml
aws s3 cp "s3://${DEPLOY_BUCKET}/deploy/backup.sh" /opt/glassact/backup.sh
chmod +x /opt/glassact/backup.sh

# Assemble the api container's env file from SSM Parameter Store.
aws ssm get-parameters-by-path --path /glassact/api --with-decryption --recursive \
  --query 'Parameters[*].[Name,Value]' --output text \
  | awk -F'\t' '{ n=$1; sub(".*/", "", n); print n"="$2 }' > /opt/glassact/api.env.new
mv /opt/glassact/api.env.new /opt/glassact/api.env
chmod 600 /opt/glassact/api.env

# docker compose needs POSTGRES_PASSWORD (and AWS_REGION, for the ECR login below)
# in its own shell environment, not just the api container's env_file.
set -a
# shellcheck disable=SC1091
source /opt/glassact/api.env
set +a

REGISTRY="${API_IMAGE%%/*}"
aws ecr get-login-password --region "${AWS_REGION}" | docker login --username AWS --password-stdin "${REGISTRY}"

cd /opt/glassact

API_IMAGE="${API_IMAGE}" docker compose up -d postgres

echo "Waiting for postgres to be healthy..."
for _ in $(seq 1 60); do
  if docker compose exec -T postgres pg_isready -U glassact -d glassact >/dev/null 2>&1; then
    break
  fi
  sleep 2
done

docker pull "${MIGRATE_IMAGE}"
docker run --rm --network glassact_internal --env-file /opt/glassact/api.env \
  "${MIGRATE_IMAGE}" -path /migrations -database "${DATABASE_DSN}" up

docker pull "${API_IMAGE}"
API_IMAGE="${API_IMAGE}" docker compose up -d api

docker image prune -f
