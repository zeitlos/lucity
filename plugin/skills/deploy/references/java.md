# Java (Railpack)

Java apps (including Spring Boot) built with Gradle or Maven.

## Detection

Any of:
- `gradlew` (Gradle wrapper) in the root — create it with `gradle wrapper` if missing
- `pom.{xml,atom,clj,groovy,rb,scala,yaml,yml}` in the root

## Version resolution (highest first)

1. `RAILPACK_JDK_VERSION` env var
2. Java 8 if the project uses Gradle ≤ 5
3. Default `21`

## Build behavior

Builds with the detected tool (Gradle via the wrapper, or Maven). Caches build artifacts: Gradle
`~/.gradle` (key `gradle`), Maven `.m2/repository` (key `maven`).

## Start command

Runs the built application (the Spring Boot / executable jar produced by the build). If the entry point
is ambiguous or you need custom flags, set the start command via `configure_service` or a `Procfile`.

## Config variables

| Variable | Effect | Example |
| :-- | :-- | :-- |
| `RAILPACK_JDK_VERSION` | JDK version | `17` |
| `RAILPACK_GRADLE_VERSION` | Gradle version | `8.5` |

## Common failure modes

- **No Gradle wrapper** → run `gradle wrapper` and commit `gradlew` + `gradle/wrapper/` (code change; propose it).
- **Wrong JDK version** → set `RAILPACK_JDK_VERSION`.
- **Gradle version mismatch** → set `RAILPACK_GRADLE_VERSION`.
- **App doesn't bind the injected port** → configure the server port from `PORT` (e.g. `server.port=${PORT}` for Spring Boot).

Docs: https://railpack.com/languages/java
