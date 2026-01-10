# Wise CLI Specification

## Purpose

Wise CLI is a command-line client for the Wise API that enables users to prepare and execute international money transfers programmatically. It is designed to work with a personal API key and serves as an agent-friendly interface for automating payments.

## Authentication

The CLI supports three methods for providing API credentials:

1. **Environment variable**: `WISE_API_TOKEN` (highest priority)
2. **Cached token**: Stored in `~/.cache/wise-cli/token` after running `wise-cli login`
3. **Command-line flag**: `--token` for one-off usage

Tokens are stored with restricted permissions (0600) for security.

## Core Functionality

### User & Profile Management

- **`login`**: Save API token from stdin for future use
- **`me`**: Display authenticated user profile (email, name, phone, occupation)
- **`profiles`**: List all Wise profiles (personal/business) with IDs and states
- **`select-profile <id-or-name>`**: Set a default profile to avoid repeating `--profile-id`

### Recipient Management

- **`recipients`**: List saved recipient accounts with optional filters:
  - `--profile-id`: Filter by profile
  - `--currency`: Filter by currency (USD, GBP, EUR, etc.)
  - `--type`: Filter by account type (iban, swift_code, etc.)
  - `--size`: Pagination size (default: 20)

- **`new recipient`**: Create recipient accounts supporting:
  - UK Sort Code (GBP): `--sort-code`, `--account-number`
  - IBAN (EUR, etc.): `--iban`
  - US Bank (USD): `--routing-number`, `--account-number`, `--account-type`
  - Email: `--email`

### Quote Management

- **`quote`**: Get unauthenticated exchange rate quote with fees and delivery estimates
- **`new quote`**: Create authenticated quote for transfer creation

Both require `--profile-id`, `--source-currency`, `--target-currency`, and either `--source-amount` or `--target-amount`.

### Transfer Management

- **`transfers [search-term]`**: List transfers with filtering:
  - `--status`: Filter by status (incoming, outgoing, cancelled)
  - `--days`: Look back N days (default: 30)
  - `--recipient`: Filter by recipient name (substring match)
  - `--reference`: Filter by reference (substring match)

- **`new transfer`**: Create transfer from a quote:
  - `--target-account`: Recipient account ID (required)
  - `--quote-uuid`: Quote UUID from `new quote` (required)
  - `--customer-transaction-id`: Idempotency key in UUID format (required)
  - `--reference`: Payment reference/memo

### High-Level Operations

- **`send-to <recipient-name> <amount> <currency> [reference]`**: All-in-one transfer command that:
  1. Finds recipient by name (exact or substring match)
  2. Creates authenticated quote automatically
  3. Creates transfer automatically
  - `--dry-run`: Preview without creating anything
  - `--customer-transaction-id`: Custom UUID (auto-generated if not set)

### Agent Integration

- **`agents md`**: Print agent instructions as markdown
- **`agents skill`**: Generate Claude Code skill file at `.claude/skills/send-money/SKILL.md`

## Caching

The CLI implements intelligent caching in `~/.cache/wise-cli/`:

- Respects `Cache-Control` and `Expires` HTTP headers
- Default TTL: 1 hour if no headers present
- Cache keys are MD5 hashes of query parameters
- Use `--refresh` flag to bypass cache

## Data Storage

All local data is stored in `~/.cache/wise-cli/`:

| File/Directory | Purpose |
|----------------|---------|
| `token` | API token |
| `default-profile` | Default profile ID |
| `*.json` | Cached API responses |
| `transfers/` | Local transfer records indexed by customer transaction ID |

## API Endpoints Used

| Operation | Endpoint |
|-----------|----------|
| User info | `GET /v1/me` |
| Profiles | `GET /v2/profiles` |
| Recipients | `GET /v2/accounts` |
| Create recipient | `POST /v1/accounts` |
| Quote | `POST /v3/profiles/{id}/quotes` |
| Transfers | `GET /v1/transfers` |
| Create transfer | `POST /v1/transfers` |

## Design Principles

1. **Agent-friendly**: Designed for automation by AI agents with structured output
2. **Idempotent operations**: Customer transaction IDs prevent duplicate transfers
3. **Flexible search**: Substring matching for recipients, profiles, and transfer filters
4. **Safe defaults**: Dry-run mode available for validation without side effects
5. **Minimal dependencies**: Uses standard Go libraries plus Cobra for CLI framework
