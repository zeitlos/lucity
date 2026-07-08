# Ruby (Railpack)

Ruby apps with Rails, Sprockets/Propshaft asset pipelines, and Bootsnap support.

## Detection

A `Gemfile` in the root directory. Rails is detected by `config/application.rb`.

## Version resolution (highest first)

1. `RAILPACK_RUBY_VERSION` env var
2. `.ruby-version`
3. `Gemfile`
4. Mise version files (`.tool-versions`, `mise.toml`)
5. Default `3.4.6`

Ruby builds from source by default (Mise default). Opt into precompiled binaries for faster builds with
`mise.toml`: `[tools] ruby = "3"` and `[settings] ruby.compile = false` (not available for all versions).

## Build behavior

Installs Ruby + system deps, installs gems, configures for production. Bundler version comes from
`BUNDLED WITH` in `Gemfile.lock`. Rails asset pipelines run `bundle exec rake assets:precompile`
(Sprockets or Propshaft); API-only apps skip it. Bootsnap is precompiled when present. jemalloc is
preloaded (`LD_PRELOAD`); Ruby 3.2+ gets YJIT (`rustc`/`cargo` installed). If a `package.json` exists or
`execjs` is used, Node.js is installed and JS assets are built (see `node.md`).

## Start command derivation

1. Framework-specific command (Rails)
2. `config/environment.rb`
3. `config.ru`
4. `Rakefile`

What breaks it: none of these present, or a non-standard server → set the start command via
`configure_service` or a `Procfile`.

## Config variables

| Variable | Effect | Example |
| :-- | :-- | :-- |
| `RAILPACK_RUBY_VERSION` | Ruby version | `3.4.2` |

Runtime defaults set: `BUNDLE_GEMFILE=/app/Gemfile`, `GEM_HOME`/`GEM_PATH=/usr/local/bundle`,
`MALLOC_ARENA_MAX=2`, `LD_PRELOAD=libjemalloc.so.2`.

## System dependencies (auto)

PostgreSQL (`libpq-dev`), MySQL (`default-libmysqlclient-dev`), Magick (`libmagickwand-dev`), Vips
(`libvips-dev`), Charlock Holmes (`libicu-dev`, `libxml2-dev`, `libxslt-dev`).

## Common failure modes

- **Wrong Ruby version / source build too slow** → pin `RAILPACK_RUBY_VERSION`, optionally enable precompiled binaries in `mise.toml`.
- **Asset precompile fails** → confirm the pipeline gem and that `RAILS_ENV=production` assets build; check runtime logs.
- **No start command derived** → set it via `configure_service` or add a `Procfile` (`web: bundle exec rails server -b 0.0.0.0 -p $PORT`).
- **Rails doesn't bind the injected port** → bind `-b 0.0.0.0 -p $PORT` in the start command.

Docs: https://railpack.com/languages/ruby
