# Running migrations

Schema migrations use [golang-migrate](https://github.com/golang-migrate/migrate),
with versioned SQL pairs living in `libs/data/migrations/`
(`000001_init.up.sql` / `.down.sql`, etc).

## Local dev

Root `package.json` exposes `pnpm db:migrate/*`, which delegate to
`libs/data`'s scripts (sourcing `libs/data/.env` for `DATABASE_DSN`):

```bash
pnpm db:migrate/new <name>   # scaffold a new up/down pair
pnpm db:migrate/up           # apply all pending migrations
pnpm db:migrate/down         # roll back one migration
pnpm db:migrate/goto <ver>   # jump to a specific version
pnpm db:migrate/force <ver>  # mark the DB as being at <ver> without running anything (for fixing a dirty state)
```

After any migration, regenerate the Jet query-builder code so `libs/data/pkg/gen`
matches the new schema:

```bash
pnpm db:gen
```

These all need the `migrate` CLI installed locally (see `docs/api.md`) and a
running dev Postgres (`pnpm dev` / `pnpm dev:api` starts one via `mprocs`).

## Production (EC2)

**Migrations do NOT run automatically.** CI/CD never touches the production
schema — `.github/workflows/deploy-api.yml` only builds and deploys the API
container; `apps/infrastructure/ec2/deploy.sh` starts `postgres` and `api` and
does not invoke `migrate`. Applying a migration to prod is always a deliberate,
manual step (see below), done whenever you're ready — not automatically when
a migration file merges to `main`.

`.github/workflows/deploy-api.yml` still builds and pushes a `glassact-migrate`
image (`apps/api/Dockerfile.migrate`, the official `migrate/migrate` image with
`libs/data/migrations` copied in), and `docker-compose.yml` on the instance
still has a `migrate` service defined. Neither runs on its own — they're there
so you can, if you want, run the exact pinned version via
`docker compose run --rm migrate` on the instance instead of the local-CLI
approach below. Either way, running it is something you choose to do, not
something a deploy does for you.

## Running migrations manually against prod

Open the SSM tunnel described in `docs/prod-database-access.md`, then run
`migrate` from your laptop against it, using the same `migrate/migrate:v4.19.1`
version pinned in `Dockerfile.migrate`:

```bash
# 1. Open the tunnel (separate terminal, leave running):
aws ssm start-session \
  --target "$(cd apps/infrastructure && terraform output -raw api_instance_id)" \
  --document-name AWS-StartPortForwardingSession \
  --parameters '{"portNumber":["5432"],"localPortNumber":["5432"]}'

# 2. Get the password and build a percent-encoded DSN (migrate needs a
#    postgres:// URI, so special characters in the password must be encoded -
#    see docs/prod-database-access.md for why):
PASSWORD=$(aws ssm get-parameter --name /glassact/api/POSTGRES_PASSWORD \
  --with-decryption --query Parameter.Value --output text)
ENCODED_PASSWORD=$(python3 -c "import urllib.parse,sys; print(urllib.parse.quote(sys.argv[1], safe=''))" "$PASSWORD")
DATABASE_DSN="postgresql://glassact:${ENCODED_PASSWORD}@localhost:5432/glassact?sslmode=disable"

# 3. Run whichever migrate subcommand you need:
migrate -path libs/data/migrations -database "$DATABASE_DSN" up
migrate -path libs/data/migrations -database "$DATABASE_DSN" down 1
migrate -path libs/data/migrations -database "$DATABASE_DSN" force <version>
```

Use `force` carefully — it doesn't run any SQL, it just overwrites the
recorded schema version, for recovering from a migration that partially
applied and left the DB "dirty".

## Notes

- `migrate/new` always creates a matching `.down.sql` — write it. There's no
  automated rollback path in production beyond running `down` manually as above.
- The migrate CLI version is pinned in two places that should stay in sync:
  `go.mod` (`github.com/golang-migrate/migrate/v4`, used by the app's Go code
  indirectly via `libs/data`) and `apps/api/Dockerfile.migrate`
  (`migrate/migrate:v4.19.1`, the version the CI-built `glassact-migrate`
  image uses, for whenever you run migrations manually against prod).
