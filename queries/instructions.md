# Wise CLI Instructions

A Go CLI client for using the Wise API to send money.

## Setup

If you haven't logged in yet, run:
```
wise login
```

This will save your personal API token for use with all commands.

To get a personal access token, see the Wise API documentation:
https://docs.wise.com/api-reference#authentication

### Verify Login

Check your login status and verify your credentials:
```
wise me
```

This will display your user information and confirm you are authenticated.

## Sending Money

### Quick Send
Send money to a recipient by name:
```
wise send-to "Recipient Name" 100 USD
```

With a payment reference:
```
wise send-to "Recipient Name" 100 USD "Payment reference"
```

Or use the `--reference` flag:
```
wise send-to "Recipient Name" 100 USD --reference "Payment reference"
```

### Prerequisites
- A Wise profile (use `wise select-profile <profile-id>` to set default)
- A recipient account (create one if needed, see below)

### Dry Run
Test the transfer without creating it:
```
wise send-to "Recipient Name" 100 USD --dry-run
```

## Creating Recipients

Create a recipient account before sending money to them.

### UK Recipient (Sort Code)
```
wise new recipient \
  --profile-id 123 \
  --currency GBP \
  --type sort_code \
  --account-holder-name "John Doe" \
  --sort-code "20-26-26" \
  --account-number "61389010"
```

### IBAN Recipient
```
wise new recipient \
  --profile-id 123 \
  --currency EUR \
  --type iban \
  --account-holder-name "Jane Doe" \
  --iban "DE89370400440532013000"
```

### US Recipient
```
wise new recipient \
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
wise new recipient \
  --profile-id 123 \
  --currency USD \
  --type email \
  --account-holder-name "Alice Johnson" \
  --email "alice@example.com"
```

### List Recipients
See existing recipients:
```
wise recipients
```

Filter by currency:
```
wise recipients --currency GBP
```

Filter by type:
```
wise recipients --type iban
```

## Listing Transfers

### List Recent Transfers
List transfers from the last 30 days:
```
wise transfers
```

### Search Transfers
Search by recipient name or reference:
```
wise transfers "search term"
```

### Filter by Status
```
wise transfers --status incoming
wise transfers --status outgoing
wise transfers --status cancelled
```

### Filter by Date Range
```
wise transfers --days 60
```

### Filter by Profile
```
wise transfers --profile-id 123
```

## Profile Management

### List Profiles
```
wise profiles
```

### Set Default Profile
```
wise select-profile <profile-id>
```

Or by name:
```
wise select-profile "Personal Account"
```

## Wise API Reference

https://docs.wise.com/api-reference
