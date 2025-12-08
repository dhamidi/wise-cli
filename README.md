# wise-cli

A command-line interface for the Wise API to send international payments.

## Installation

```bash
go install github.com/dhamidi/wise-cli/cmd/wise-cli@latest
```

Or build from source:

```bash
git clone https://github.com/dhamidi/wise-cli.git
cd wise-cli
go build -o wise-cli ./cmd/wise-cli
```

## Getting Started

### 1. Get a Wise API Token

Create a personal API token at <https://wise.com/settings/api-tokens>

### 2. Login

```bash
wise-cli login
```

This saves your token locally for future commands.

Alternatively, set the `WISE_API_TOKEN` environment variable or use `--token` with each command.

### 3. Select a Profile

List your profiles:

```bash
wise-cli profiles
```

Select the profile to use for transfers:

```bash
wise-cli select-profile <profile-id-or-name>
```

### 4. Send Money

Send money to an existing recipient:

```bash
wise-cli send-to "John Doe" 100 EUR
```

Add a payment reference:

```bash
wise-cli send-to "John Doe" 100 EUR "Invoice #123"
```

Preview a transfer without creating it:

```bash
wise-cli send-to "John Doe" 100 EUR --dry-run
```

## Commands

| Command | Description |
|---------|-------------|
| `login` | Save your Wise API token |
| `profiles` | List your Wise profiles |
| `select-profile` | Set the default profile for transfers |
| `recipients` | List your saved recipients |
| `send-to` | Send money to a recipient |
| `quote` | Get an exchange rate quote |
| `new quote` | Create a quote for a transfer |
| `new transfer` | Create a transfer from a quote |
| `new recipient` | Create a new recipient |

## Examples

List recipients for a specific currency:

```bash
wise-cli recipients --currency EUR
```

Get a quote for 1000 EUR to GBP:

```bash
wise-cli quote --profile-id 12345 --source-currency EUR --target-currency GBP --source-amount 1000
```

## Configuration

The CLI stores configuration in `~/.cache/wise-cli/`:

- `token` - Your API token
- `default_profile` - Your selected profile ID
- Response cache for improved performance

Use `--refresh` with any command to bypass the cache.

## Best Used With an Agent

This CLI is designed to be used by AI coding agents like [Amp](https://ampcode.com). Give your agent access to your terminal and let it handle international payments for you.

## FAQ

**Is this secure?**

Against what threat model?

**Is this not more effort than clicking buttons in the Wise App?**

It's a tool for your agent, not for you.

**Was this written by AI?**

100%, with human oversight.

**Why?**

I got tired of paying my bills manually.
