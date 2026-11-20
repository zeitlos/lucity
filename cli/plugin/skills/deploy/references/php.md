# PHP (Railpack)

PHP apps served with FrankenPHP, with first-class Laravel support.

## Detection

Any of:
- `index.php` in the root
- `composer.json` in the root

## Version resolution (highest first)

1. `composer.json` (the `php` requirement)
2. Default `8.4`

No other version file is read (no `.php-version`, `.tool-versions` or `mise.toml`); pin the version in
`composer.json`'s `require.php`.

Only PHP 8.2+ is supported.

## Build behavior

Configures [FrankenPHP](https://frankenphp.dev/) as the app server. For Laravel, the document root is
`public`. Caches `/opt/cache/composer`. If a `package.json` is present, Node.js is installed, npm
dependencies installed, and build scripts run (useful for Laravel + Vue/React assets); dev dependencies
are pruned. See `node.md` for Node configuration.

## Start command

Started via a `start-container.sh` script (overridable by placing your own in the repo root). For
Laravel it runs migrations + seeding (disable with `RAILPACK_SKIP_MIGRATIONS=true`), creates storage
symlinks, optimizes, then starts FrankenPHP with the Caddyfile.

## PHP extensions

Installed automatically from `composer.json` `ext-*` requirements (e.g. `ext-redis`, `ext-pgsql`) and
from `RAILPACK_PHP_EXTENSIONS` (comma-separated).

## Custom configuration

Override defaults by placing files in the repo root: `/Caddyfile` (Caddy server config), `/php.ini` (PHP
config), `/start-container.sh` (startup process).

## Config variables

| Variable | Effect | Example |
| :-- | :-- | :-- |
| `RAILPACK_PHP_ROOT_DIR` | Document root | `/app/public` |
| `RAILPACK_PHP_EXTENSIONS` | Extra PHP extensions (comma-separated) | `gd,imagick,redis` |
| `RAILPACK_SKIP_MIGRATIONS` | Skip Laravel migrations at start | `true` |

## Common failure modes

- **Missing PHP extension at runtime** → add it to `composer.json` `require` (`ext-*`) or `RAILPACK_PHP_EXTENSIONS`.
- **Laravel serves the wrong directory** → set `RAILPACK_PHP_ROOT_DIR` to `/app/public`.
- **Migrations run on a DB not yet reachable at boot** → set `RAILPACK_SKIP_MIGRATIONS=true` and run migrations separately.
- **Frontend assets not built** → ensure the `build` script in `package.json` produces assets; see `node.md`.
- **`composer install` fails on a path repository** (a `"type": "path"` package in the repo) → Railpack copies only the `composer.json` / `composer.lock` files (and `artisan`) before installing, not the package sources, so the path package is incomplete. Set `RAILPACK_INSTALL_CMD` to `composer install --no-dev --optimize-autoloader --no-interaction`, which runs with all files present.

Docs: https://railpack.com/languages/php
