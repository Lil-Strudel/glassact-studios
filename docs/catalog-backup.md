# Catalog backup and restore

The production catalog is the source of truth for the dev seed. A snapshot is
the catalog rows, the reference data their manifests point at by id (price
groups, glass colors, grouts), and every SVG those rows reference.

## Take a snapshot

Needs an SSO session for the `glassact-studios` AWS profile
(`aws sso login --profile glassact-studios`).

```bash
pnpm db:catalog/backup
```

This dumps on the `glassact-api` instance via SSM `send-command`, hands the
dump back through the backups bucket (`postgres/catalog-<timestamp>.sql`), and
writes:

```
libs/data/backups/<YYYY-MM-DD>/
  catalog.sql        # data-only dump, also copied to libs/data/catalog_seed.sql
  file-keys.txt      # every S3 key catalog.sql references
  files/file/...     # those SVGs, laid out by S3 key
  MANIFEST.sha256
```

`backups/` is gitignored; commit the refreshed `catalog_seed.sql`.

## Restore the database

On a freshly migrated database:

```bash
pnpm db:seed                                      # catalog_seed.sql + dev users
pnpm db:catalog/restore -- libs/data/backups/<date>   # or a specific snapshot
```

Both refuse to load over existing catalog rows. `--replace` truncates the
catalog tables with `CASCADE`, which also wipes local projects and inlays:

```bash
pnpm db:catalog/restore -- --replace
```

## Restore the images

Dev reads from the shared `glassact-nkm52j` bucket, so the SVGs normally
resolve without restoring anything. To repopulate a bucket from a snapshot
(disaster recovery, or a new bucket):

```bash
pnpm db:catalog/restore-files -- libs/data/backups/<date> <bucket>
```
