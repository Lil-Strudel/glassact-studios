# Connecting to the prod database locally

Postgres runs in a Docker container on the `glassact-api` EC2 instance, bound
only to `127.0.0.1:5432` on that box (see `apps/infrastructure/ec2/docker-compose.yml`).
It is never reachable over the network directly — the security group has no
rule for port 5432 at all. The only way in is an SSM port-forwarding session,
which tunnels through to that loopback port.

## Prerequisites

- AWS CLI configured with credentials that can `ssm:StartSession` on the instance.
- The Session Manager plugin installed locally (needed for `start-session`,
  not just `send-command`):
  https://docs.aws.amazon.com/systems-manager/latest/userguide/session-manager-working-with-install-plugin.html
- Everything is in `us-west-2` (passed explicitly on each command below).

## 1. Open the tunnel

```bash
cd apps/infrastructure
aws ssm start-session \
  --region us-west-2 \
  --target "$(terraform output -raw api_instance_id)" \
  --document-name AWS-StartPortForwardingSession \
  --parameters '{"portNumber":["5432"],"localPortNumber":["5432"]}'
```

Leave this running in its own terminal. It forwards `localhost:5432` on your
machine through SSM to `127.0.0.1:5432` on the instance (the postgres
container). Ctrl+C closes it.

## 2. Get the password

The Postgres password lives in SSM Parameter Store, not in Terraform state:

```bash
aws ssm get-parameter --region us-west-2 --name /glassact/api/POSTGRES_PASSWORD \
  --with-decryption --query Parameter.Value --output text
```

## 3. Connect

**psql** (recommended for ad hoc queries — env vars avoid any DSN encoding issues):

```bash
PGPASSWORD='<password from step 2>' psql -h localhost -p 5432 -U glassact -d glassact
```

**golang-migrate** (needs a `postgres://` URI, so the password must be
percent-encoded if it contains `/`, `+`, `=`, `@`, `:`, or other reserved
characters — which base64-generated passwords often do):

```bash
ENCODED_PASSWORD=$(python3 -c "import urllib.parse,sys; print(urllib.parse.quote(sys.argv[1], safe=''))" '<password from step 2>')
DATABASE_DSN="postgresql://glassact:${ENCODED_PASSWORD}@localhost:5432/glassact?sslmode=disable"

migrate -path libs/data/migrations -database "$DATABASE_DSN" up
```

## Notes

- This is for manual/ad hoc access (debugging, one-off queries, running
  migrations). Schema migrations are never applied automatically by CI/CD —
  `apps/infrastructure/ec2/deploy.sh` only starts `postgres` and `api`. See
  `docs/migrations.md` for the manual migration workflow that uses this tunnel.
- Nightly backups land in the `glassact-backups-*` S3 bucket
  (`terraform output backups_bucket_name`), under the `postgres/` prefix,
  30-day retention.
