# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.2.0] - 2026-10-05

### Added

- `bw login --read-only` asks for a token that can view but not change anything.
- `bw login --org <slug>` asks for a token limited to one organization and makes it the default.
- `bw tasks list` shows each action item's assignees.

### Changed

- `bw login` collects its token from the standard OAuth token endpoint, `POST /api/v1/auth/token` (device-code grant), which replaces `/api/v1/auth/device/token`. It reads the RFC 6749 `error` code rather than the HTTP status, and honours `slow_down`.
- `bw logout` revokes the token on the server (`DELETE /api/v1/auth`) before clearing local credentials. If the server can't be reached, it still clears them and warns that the token may still be valid, pointing to the API tokens page.
- Errors show the server's explanation, such as "We couldn't find that. (not_found)", instead of the raw response body, and list the fields a validation error names.
- A request refused as too many (429) is retried once after the server's `Retry-After` (at most 60 seconds).
- `bw my consents` and `bw my declarations` show the web page where you sign or complete each one, which the API can't do.

### Removed

- `bw superadmin` commands. The API refuses access tokens there by design, so they could only fail. An old config file with a `superadmin` setting still loads.
- `bw agenda add --position`. The API always adds an item at the end of the agenda, so the flag had no effect.

### Fixed

- `bw tasks create --assign` created the action item with no assignees. Create commands now send their fields inside the resource key the API expects (`action_item`, `meeting`, `agenda_item`, `message`).
- `bw my consents`, `bw my declarations` and `bw my surveys` printed blank titles.
- List commands showed only the first page (50 records). They now read every page, 100 at a time, and `--json` prints every item in the API's own shape.

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

[Unreleased]: https://github.com/boardwiseco/boardwise-cli/compare/v0.2.0...HEAD
[0.2.0]: https://github.com/boardwiseco/boardwise-cli/compare/v0.1.0...v0.2.0
[0.1.0]: https://github.com/boardwiseco/boardwise-cli/releases/tag/v0.1.0
