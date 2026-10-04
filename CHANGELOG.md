# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Changed

- `bw login` collects its token from the standard OAuth token endpoint, `POST /api/v1/auth/token` (device-code grant), which replaces `/api/v1/auth/device/token`. It reads the RFC 6749 `error` code rather than the HTTP status, and honours `slow_down`.

### Fixed

- `bw superadmin system` and `bw superadmin orgs get` printed a literal `%v` instead of the counts.

## [0.1.0] - 2026-03-06

### Added

- `bw login` — authenticate via browser device authorization flow
- `bw logout` — clear stored credentials
- `bw me` — show current user and organizations
- `bw version` — show CLI version
- `bw boards list / get` — list and inspect boards and committees
- `bw meetings list / get / create` — manage meetings
- `bw agenda add` — add agenda items to a meeting
- `bw people list` — list active people in an organization
- `bw tasks list / create` — manage action items
- `bw docs list` — browse documents
- `bw messages list / get / send` — read and send board messages
- `bw my notifications / schedule / rsvps / tasks / messages / consents / declarations / surveys` — personal dashboard across all organizations
- `bw superadmin orgs / users / system` — superadmin operations (hidden unless the authenticated user is a superadmin)
- `--json` flag on all commands for machine-readable output
- `--org` flag and stored default org for organization-scoped commands
- `--url` flag and `BW_API_URL` environment variable for targeting non-production servers

[Unreleased]: https://github.com/boardwiseco/boardwise-cli/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/boardwiseco/boardwise-cli/releases/tag/v0.1.0
