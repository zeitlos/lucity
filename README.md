<p>
  <picture>
    <source media="(prefers-color-scheme: dark)" srcset="assets/logo-dark.svg">
    <img src="assets/logo.svg" alt="Lucity" width="230">
  </picture>
</p>

<p>
  The Deployment Platform that is a joy to use, built to grow, and always yours to keep.
</p>

<p>
  Lucity is an open-source Vercel/Railway/Heroku alternative deliberately designed to avoid lock-in. Push to deploy, one-click PostgreSQL, Redis and object storage.
</p>

<p>
  <a href="https://github.com/zeitlos/lucity/actions/workflows/release.yml">
    <img src="https://github.com/zeitlos/lucity/actions/workflows/release.yml/badge.svg?branch=main" alt="CI">
  </a>
  <a href="https://github.com/zeitlos/lucity/blob/main/LICENSE">
    <img src="https://img.shields.io/github/license/zeitlos/lucity?color=blue" alt="License">
  </a>
  <a href="https://go.dev/">
    <img src="https://img.shields.io/badge/Go-1.27-00ADD8?logo=go&logoColor=white" alt="Go 1.27">
  </a>
</p>

<p align="center">
  <img src="docs/public/img/docs/projects/canvas.png" alt="A project on the Lucity canvas: a GitHub service, a PostgreSQL database and an object storage bucket" width="100%">
</p>

<table>
  <tr>
    <td width="50%" valign="top">
      <img src="docs/public/img/docs/quickstart/first-build.png" alt="Live build logs of a Next.js app">
      <p align="center"><b>Push to deploy.</b> Builds straight from your repo, no Dockerfile needed. <a href="https://lucity.cloud/docs/quickstart">Quickstart</a></p>
    </td>
    <td width="50%" valign="top">
      <img src="docs/public/img/docs/quickstart/variable-reference.png" alt="Referencing database and bucket credentials from a service variable">
      <p align="center"><b>Wire services together.</b> Reference database and bucket credentials instead of copying them. <a href="https://lucity.cloud/docs/variables">Variables</a></p>
    </td>
  </tr>
  <tr>
    <td width="50%" valign="top">
      <img src="docs/public/img/docs/postgres/explorer-query.png" alt="Running a SQL query against a managed PostgreSQL database">
      <p align="center"><b>Managed PostgreSQL.</b> Browse tables and run queries right in the dashboard. <a href="https://lucity.cloud/docs/postgres">PostgreSQL</a></p>
    </td>
    <td width="50%" valign="top">
      <img src="docs/public/img/docs/postgres/backups.png" alt="Point-in-time restore for a PostgreSQL database">
      <p align="center"><b>Point-in-time restore.</b> Every write is archived, so you can rewind to any moment in the last 30 days.</p>
    </td>
  </tr>
  <tr>
    <td width="50%" valign="top">
      <img src="docs/public/img/docs/metrics/database-metrics.png" alt="CPU, memory, storage and connection metrics for a database">
      <p align="center"><b>Metrics and logs built in.</b> CPU, memory, storage and connections at a glance. <a href="https://lucity.cloud/docs/metrics">Metrics</a></p>
    </td>
    <td width="50%" valign="top">
      <img src="docs/public/img/docs/eject/eject-dialog.png" alt="The eject dialog exporting a project as a Helm chart">
      <p align="center"><b>Eject anytime.</b> Download your project as a plain Helm chart and run it without Lucity. <a href="https://lucity.cloud/docs/eject">Eject</a></p>
    </td>
  </tr>
</table>

## Documentation

Full documentation at **[lucity.cloud](https://lucity.cloud)**.

## Lucity Cloud

Don't want to run your own infrastructure? **[Lucity Cloud](https://lucity.cloud/)** is the managed version of Lucity.

## Built by zeitlos

Lucity is made by [zeitlos.software](https://zeitlos.software), a Swiss software company.

## Contributing

This project is still relatively early stage, but if you want to contribute, see [CONTRIBUTING.md](CONTRIBUTING.md) for development setup, architecture overview, and how to get started.

## License

[AGPL-3.0](LICENSE)
