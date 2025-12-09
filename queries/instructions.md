# Wise CLI Instructions

A Go CLI client for using the Wise API to send money.

## Setup

If you haven't logged in yet, run:
```
wise-cli login
```

This will save your personal API token for use with all commands.

To get a personal access token, see the Wise API documentation:
https://docs.wise.com/api-reference#authentication

### Verify Login

Check your login status and verify your credentials:
```
wise-cli me
```

This will display your user information and confirm you are authenticated.

## Sending Money

### Quick Send
Send money to a recipient by name:
```
wise-cli send-to "Recipient Name" 100 USD
```

With a payment reference:
```
wise-cli send-to "Recipient Name" 100 USD "Payment reference"
```

Or use the `--reference` flag:
```
wise-cli send-to "Recipient Name" 100 USD --reference "Payment reference"
```

### Prerequisites
- A Wise profile (use `wise-cli select-profile <profile-id>` to set default)
- A recipient account (create one if needed, see below)

### Dry Run
Test the transfer without creating it:
```
wise-cli send-to "Recipient Name" 100 USD --dry-run
```

## Creating Recipients

Create a recipient account before sending money to them.

### UK Recipient (Sort Code)
```
wise-cli new recipient \
  --profile-id 123 \
  --currency GBP \
  --type sort_code \
  --account-holder-name "John Doe" \
  --sort-code "20-26-26" \
  --account-number "61389010"
```

### IBAN Recipient
```
wise-cli new recipient \
  --profile-id 123 \
  --currency EUR \
  --type iban \
  --account-holder-name "Jane Doe" \
  --iban "DE89370400440532013000"
```

### US Recipient
```
wise-cli new recipient \
  --profile-id 123 \
  --currency USD \
  --type us \
  --account-holder-name "Bob Smith" \
  --routing-number "021000021" \
  --account-number "123456789" \
  --account-type CHECKING
```

### Email Recipient
```
wise-cli new recipient \
  --profile-id 123 \
  --currency USD \
  --type email \
  --account-holder-name "Alice Johnson" \
  --email "alice@example.com"
```

### List Recipients
See existing recipients:
```
wise-cli recipients
```

Filter by currency:
```
wise-cli recipients --currency GBP
```

Filter by type:
```
wise-cli recipients --type iban
```

## Listing Transfers

### List Recent Transfers
List transfers from the last 30 days:
```
wise-cli transfers
```

### Search Transfers
Search by recipient name or reference:
```
wise-cli transfers "search term"
```

### Filter by Status
```
wise-cli transfers --status incoming
wise-cli transfers --status outgoing
wise-cli transfers --status cancelled
```

### Filter by Date Range
```
wise-cli transfers --days 60
```

### Filter by Profile
```
wise-cli transfers --profile-id 123
```

## Profile Management

### List Profiles
```
wise-cli profiles
```

### Set Default Profile
```
wise-cli select-profile <profile-id>
```

Or by name:
```
wise-cli select-profile "Personal Account"
```

## Wise API Reference

https://docs.wise.com/api-reference
