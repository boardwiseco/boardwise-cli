# bw — Boardwise CLI

Command line interface for [Boardwise](https://boardwise.co) — manage your boards, meetings, tasks, and more from the terminal.

## Installation

### Download a release (recommended)

Download the latest binary for your platform from the [releases page](https://github.com/boardwiseco/boardwise-cli/releases).

**macOS (Apple Silicon)**
```bash
curl -L https://github.com/boardwiseco/boardwise-cli/releases/latest/download/bw_latest_darwin_arm64.tar.gz | tar xz
sudo mv bw /usr/local/bin/
```

**macOS (Intel)**
```bash
curl -L https://github.com/boardwiseco/boardwise-cli/releases/latest/download/bw_latest_darwin_amd64.tar.gz | tar xz
sudo mv bw /usr/local/bin/
```

**Linux (amd64)**
```bash
curl -L https://github.com/boardwiseco/boardwise-cli/releases/latest/download/bw_latest_linux_amd64.tar.gz | tar xz
sudo mv bw /usr/local/bin/
```

### Build from source

Requires Go 1.21+.

```bash
git clone https://github.com/boardwiseco/boardwise-cli.git
cd boardwise-cli
go build -o bw .
```

## Quick Start

```bash
# Log in — opens your browser to authorize the CLI
bw login

# See who you're logged in as
bw me

# List your upcoming meetings across all organizations
bw my schedule

# List boards in an organization
bw --org 974826 boards list
```

The `--org` flag accepts the numeric organization slug shown in the Boardwise URL (e.g. `app.boardwise.co/974826/`). If you only belong to one organization, it is set automatically after login.

## Commands

### Authentication

| Command | Description |
|---------|-------------|
| `bw login` | Authenticate via browser device authorization |
| `bw logout` | Clear stored credentials |
| `bw me` | Show current user and organizations |
| `bw version` | Show CLI version |

### Boards & Committees

| Command | Description |
|---------|-------------|
| `bw boards list` | List boards and committees |
| `bw boards get <slug>` | Get details and member list |

### Meetings

| Command | Description |
|---------|-------------|
| `bw meetings list` | List meetings (`--status upcoming\|past\|all`) |
| `bw meetings get <id>` | Get agenda, attendees, and documents |
| `bw meetings create` | Create a meeting (`--title`, `--starts-at`, `--ends-at` required) |
| `bw agenda add <meeting-id>` | Add an agenda item (`--title`, `--duration` required) |

### People & Tasks

| Command | Description |
|---------|-------------|
| `bw people list` | List active people in the organization |
| `bw tasks list` | List action items (`--status pending\|completed\|all`) |
| `bw tasks create` | Create an action item (`--title` required) |

### Documents & Messages

| Command | Description |
|---------|-------------|
| `bw docs list` | List documents |
| `bw messages list <group-slug>` | List messages for a board or committee |
| `bw messages get <id>` | Get a message |
| `bw messages send <group-slug>` | Send a message (`--subject`, `--body` required) |

### Personal Dashboard

These commands work across all your organizations.

| Command | Description |
|---------|-------------|
| `bw my schedule` | Upcoming and past meetings |
| `bw my rsvps` | Meetings with pending RSVPs |
| `bw my tasks` | Pending tasks assigned to you |
| `bw my messages` | Unread messages |
| `bw my notifications` | Unread notifications |
| `bw my notifications mark-read` | Mark all notifications as read |
| `bw my consents` | Consent packets awaiting your signature |
| `bw my declarations` | Declarations awaiting your response |
| `bw my surveys` | Surveys awaiting your response |

## Global Flags

| Flag | Description |
|------|-------------|
| `-o, --org <slug>` | Organization slug (saved after login if you only have one org) |
| `--json` | Output raw JSON — useful for scripting |
| `--url <url>` | API base URL (default: `https://app.boardwise.co`) |

The `--url` flag and `BW_API_URL` environment variable are useful for targeting a local development server:

```bash
BW_API_URL=http://boardwise.test bw login
```

## Shell Completion

```bash
# Bash
bw completion bash > /usr/local/etc/bash_completion.d/bw

# Zsh
bw completion zsh > "${fpath[1]}/_bw"

# Fish
bw completion fish > ~/.config/fish/completions/bw.fish
```

## Releasing

Releases are built automatically by GitHub Actions when a tag is pushed:

```bash
git tag v0.2.0
git push origin v0.2.0
```

This produces signed archives and a checksums file for macOS (arm64/amd64), Linux (arm64/amd64), and Windows (amd64) on the [releases page](https://github.com/boardwiseco/boardwise-cli/releases).
