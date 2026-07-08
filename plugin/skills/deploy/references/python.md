# Python (Railpack)

Python apps across pip, poetry, pdm, uv, and pipenv, with framework-aware start commands.

## Detection

Any of:
- One of `main.py`, `app.py`, `start.py`, `bot.py`, `hello.py`, `server.py` in the root
- `requirements.txt`
- `pyproject.toml`
- `Pipfile`

## Version resolution (highest first)

1. `RAILPACK_PYTHON_VERSION` env var
2. Mise version files (`.python-version`, `.tool-versions`, `mise.toml`)
3. `runtime.txt`
4. `Pipfile`
5. Default `3.13.2`

Only non-EOL Python versions (3.10+) are officially supported.

## Precompiled vs. compiled

Python installs from precompiled Mise binaries by default. If a precompiled binary is unavailable for
the requested version, the build **fails on purpose** (compiling from source often breaks packages that
ship prebuilt wheels). To opt into source compilation set the deploy variable `MISE_PYTHON_COMPILE=1` or
`[settings.python] compile = true` in `mise.toml`. Prefer pinning a version that has a precompiled binary.

## Package managers

- **pip** — `requirements.txt`
- **poetry** — `pyproject.toml` + `poetry.lock`
- **pdm** — `pyproject.toml` + `pdm.lock`
- **uv** — `pyproject.toml` + `uv.lock` (part of the `python` provider, not a separate provider)
- **pipenv** — `Pipfile`

## Start command derivation

1. Framework-specific command (below)
2. First main file found, checked in order: `main.py`, `app.py`, `start.py`, `bot.py`, `hello.py`, `server.py`

Framework start commands (all bind `0.0.0.0` and `$PORT`, default 8000):
- **Flask** (with `gunicorn`): `gunicorn --bind 0.0.0.0:${PORT:-8000} main:app`
- **FastAPI** / **FastHTML** (with `uvicorn`): `uvicorn main:app --host 0.0.0.0 --port ${PORT:-8000}`
- **Django** (`manage.py` or `django` dep): `python manage.py migrate && gunicorn {appName}:application`, where `appName` comes from `RAILPACK_DJANGO_APP_NAME`, else the `WSGI_APPLICATION` setting.

What breaks it: a non-standard entry module (app object not named `app`, or file not one of the six
detected names) → set `RAILPACK_START_CMD` / `configure_service`, or add a `Procfile`.

## Config variables

| Variable | Effect | Example |
| :-- | :-- | :-- |
| `RAILPACK_PYTHON_VERSION` | Python version | `3.11` |
| `RAILPACK_DJANGO_APP_NAME` | Django WSGI module | `myapp.wsgi` |

Runtime defaults set: `PYTHONUNBUFFERED=1`, `PYTHONDONTWRITEBYTECODE=1`, `PIP_DISABLE_PIP_VERSION_CHECK=1`,
and more.

## System dependencies (auto)

Installs libs for common packages: pycairo (`libcairo2`), pdf2image (`poppler-utils`), pydub/pymovie
(`ffmpeg`, Qt), Playwright (headless Chromium). Databases: PostgreSQL (`libpq-dev` build / `libpq5`
runtime), MySQL (`default-libmysqlclient-dev` / `default-mysql-client`).

## Common failure modes

- **Build fails: no precompiled binary for the Python version** → pin `RAILPACK_PYTHON_VERSION` to a version with a prebuilt binary, or opt into `MISE_PYTHON_COMPILE=1`.
- **Wrong start command (app object not `app`)** → set the start command via `configure_service` or a `Procfile`.
- **Django app name misdetected** → set `RAILPACK_DJANGO_APP_NAME`.
- **Missing system library at build (e.g. a C extension)** → add `RAILPACK_BUILD_APT_PACKAGES`.

Docs: https://railpack.com/languages/python
