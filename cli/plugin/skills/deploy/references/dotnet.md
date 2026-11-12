# .NET (Railpack)

.NET applications. Provider value: `dotnet`.

## Detection

A `*.csproj` file in the root directory.

## Version resolution (highest first)

1. `TargetFramework` from the first `*.csproj`
2. `version` from `global.json` in the root
3. `RAILPACK_DOTNET_VERSION` env var
4. Default `6.0.428`

## Build behavior

Installs the .NET SDK + runtime, restores with `dotnet restore`, and builds with
`dotnet publish --no-restore -c Release -o out`. `libicu-dev` is installed for internationalization.

## Start command

Runs the publish output: `./out/{project_name}`.

## Port binding

Railpack sets `ASPNETCORE_URLS=http://0.0.0.0:${PORT:-3000}` in the start command, so ASP.NET Core apps
bind the injected `PORT` on all interfaces automatically. If your app reads a URL from elsewhere, make it
honor `ASPNETCORE_URLS` or `PORT`.

## Config variables

| Variable | Effect | Example |
| :-- | :-- | :-- |
| `RAILPACK_DOTNET_VERSION` | .NET version | `6.0.428` |

## Common failure modes

- **Wrong target framework** → set `TargetFramework` in the `.csproj` or pin `RAILPACK_DOTNET_VERSION`.
- **Multiple projects, wrong one published** → keep the intended `.csproj` at the root, or use a `global.json` / custom build command.
- **App ignores `ASPNETCORE_URLS`** → remove any hardcoded `UseUrls(...)`/`applicationUrl` that overrides it (code fix; propose it).

Docs: https://railpack.com/languages/dotnet
