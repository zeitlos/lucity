# Working in this repo

- **No code comments.** The code speaks for itself; don't add explanatory or narrative comments. (Doc comments on exported APIs where genuinely useful are fine.)
- **Never commit or push.** The maintainer commits manually, always. Don't stage, commit, or push, and don't ask to.
- **Don't start or stop the dev services.** They run with hot reload under the maintainer's control. If verifying a change needs running services, ask the maintainer to start them rather than launching them yourself.
- **Validate what you can on your own**: workspace-wide build, vet, and unit tests. For behavior that needs the running stack, drive it through the API or dashboard once the maintainer confirms services are up.
- **Absolute paths only** in commands. No `~` or `$HOME`.
