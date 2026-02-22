# 🎨 Brandfetch CLI — Brand assets in your terminal.

Brandfetch in your terminal. Fetch logos, colors, and fonts for any company.

## Features

- **Brand assets** - download logos in SVG/PNG/WebP format, extract colors, and get fonts
- **Quick mode** - essentials in one call, export CSS/Tailwind, optional downloads + checksums
- **Search** - find brands by name or keyword
- **Transaction matching** - resolve transaction labels to brands
- **Webhooks** - manage webhooks via GraphQL API
- **GraphQL** - run arbitrary GraphQL queries and mutations
- **Secure storage** - credentials stored in OS keychain
- **Theme variants** - fetch light or dark logo variants

## Installation

### Homebrew

```bash
brew install salmonumbrella/tap/brandfetch
```

### Go Install

```bash
go install github.com/salmonumbrella/brandfetch-cli/cmd/brandfetch@latest
```

## Quick Start

### 1. Get API Keys

Visit [brandfetch.com/developers](https://www.brandfetch.com/developers) to obtain your API keys:

- **Logo API Key / Client ID**: High quota (10,000+ requests/month), supports logo downloads and search
- **Brand API Key**: Limited quota (100 requests/month), supports full brand metadata including colors and fonts

Command requirements:
- Logo/Search only need the **Logo API Client ID**
- Brand/Colors/Fonts/Quick/Transaction/Webhooks need the **Brand API Key**

### 2. Configure Credentials

Choose one of three methods:

**Interactive:**
```bash
brandfetch auth set
```

**Stdin (CI/non-interactive):**
```bash
echo -e "your-client-id\nyour-api-key" | brandfetch auth set --stdin
```

**Environment variables:**
```bash
export BRANDFETCH_CLIENT_ID="your-client-id"
export BRANDFETCH_API_KEY="your-api-key"
```

### 3. Fetch Your First Logo

```bash
brandfetch logo stripe.com
```

## Configuration

### Environment Variables

- `BRANDFETCH_CLIENT_ID` - Logo API Client ID (high quota)
- `BRANDFETCH_API_KEY` - Brand API Key (limited quota)
- `BRANDFETCH_OUTPUT` - Output format: `text` (default) or `json`
- `BRANDFETCH_COLOR` - Color mode: `auto` (default), `always`, or `never`
- `NO_COLOR` - Set to any value to disable colors (standard convention)

## Security

### Credential Storage

Credentials are stored securely in your system's keychain:
- **macOS**: Keychain Access
- **Linux**: Secret Service (GNOME Keyring, KWallet)
- **Windows**: Credential Manager

## Rate Limiting

The Brandfetch API enforces quotas and rate limits based on your API plan. If you hit HTTP 429 or quota errors, back off and retry in your scripts. Logo/Search use the Logo API Client ID (higher quota) while Brand endpoints use the Brand API Key (lower quota).

## Commands

### Authentication

```bash
brandfetch auth set                      # Set credentials (interactive prompt)
brandfetch auth set --stdin              # Set credentials from stdin (for CI)
brandfetch auth status                   # Show credential status
brandfetch auth clear                    # Remove stored credentials
```

### Logo

```bash
brandfetch logo <identifier>                 # Logo URL (SVG, light theme)
brandfetch logo <identifier> --format png    # PNG format
brandfetch logo <identifier> --theme dark    # Dark theme variant
brandfetch logo <identifier> --type icon     # Icon variant
brandfetch logo <identifier> --width 256     # Custom width
brandfetch logo <identifier> --output json   # Logo metadata as JSON
brandfetch logo download <identifier>        # Download logo asset
brandfetch logo download <identifier> --path ./logo.svg
brandfetch logo download <identifier> --sha256 <hex>    # Verify checksum
```

### Brand

```bash
brandfetch brand <identifier>                # Get comprehensive brand data
brandfetch brand <identifier> --output json  # Full Brand API response as JSON
```

Returns brand name, description, industry, domain, social links, colors, fonts, and logo URLs.

**Note**: Requires Brand API key (limited quota)

### Search

```bash
brandfetch search <query>                # Search for brands by name
brandfetch search <query> --output json  # Search results as JSON
```

### Colors

```bash
brandfetch colors <identifier>               # Get brand color palette
brandfetch colors <identifier> --output json # Colors as JSON
```

**Note**: Requires Brand API key (limited quota)

### Fonts

```bash
brandfetch fonts <identifier>                # Get brand typography
brandfetch fonts <identifier> --output json  # Fonts as JSON
```

**Note**: Requires Brand API key (limited quota)

### Quick

```bash
brandfetch quick <identifier>                # Logos, favicon, colors, fonts
brandfetch quick <identifier> --output json  # Essentials as JSON
brandfetch quick <identifier> --css          # CSS custom properties
brandfetch quick <identifier> --tailwind     # Tailwind config
brandfetch quick <identifier> --download ./assets --sha256  # Download + checksums
brandfetch quick <identifier> --download ./assets --sha256-manifest ./checksums.sha256
brandfetch quick <identifier> --download ./assets --sha256-manifest-out ./checksums.sha256
brandfetch quick <identifier> --download ./assets --sha256-manifest-out ./checksums.sha256 --sha256-manifest-append
brandfetch quick <identifier> --download ./assets --sha256-manifest ./checksums.sha256 --sha256-manifest-verify
```

### Transaction

```bash
brandfetch transaction "STARBUCKS 1234 SEATTLE WA"      # Match transaction label
brandfetch transaction "Spotify USA" --country US       # With country hint
brandfetch transaction "Spotify USA" --output json      # Full Brand response as JSON
```

### Webhooks

```bash
brandfetch webhooks create --url https://example.com/webhooks --events brand.updated,brand.verified
brandfetch webhooks list
brandfetch webhooks list --enabled --event brand.updated
brandfetch webhooks list --url-contains example.com
brandfetch webhooks list --json-flat
brandfetch webhooks list --table
brandfetch webhooks list --table --table-truncate 24
brandfetch webhooks list --table --columns urn,url,status,events
brandfetch webhooks subscribe --webhook urn:bf:webhook:123 --subscriptions urn:bf:brand:abc,urn:bf:brand:def
brandfetch webhooks unsubscribe --webhook urn:bf:webhook:123 --subscriptions urn:bf:brand:abc
```

### GraphQL

```bash
brandfetch graphql --query "query { me { id } }"
brandfetch graphql --query-file ./query.graphql --variables '{"input": {"url": "https://example.com"}}'
cat query.graphql | brandfetch graphql --stdin
cat payload.json | brandfetch graphql --stdin-raw
```


## Output Formats

### Text

Human-readable output with colors and formatting:

```bash
$ brandfetch brand stripe.com
Stripe
Online payment processing for internet businesses

Industry: Financial Services
Domain: stripe.com

Colors:
  #635BFF (primary, dark)
  #0A2540 (accent, dark)

Fonts:
  Camphor (sans-serif, custom)
```

### JSON

Machine-readable output:

```bash
$ brandfetch brand stripe.com --output json
{
  "name": "Stripe",
  "description": "Online payment processing for internet businesses",
  "domain": "stripe.com",
  "industry": "Financial Services",
  "colors": [
    {"hex": "#635BFF", "type": "primary", "brightness": 50}
  ],
  "fonts": [
    {"name": "Camphor", "type": "sans-serif", "origin": "custom"}
  ]
}
```

Data goes to stdout, errors and progress to stderr for clean piping.

## Examples

### Download logos for multiple brands

```bash
for domain in stripe.com github.com figma.com; do
  brandfetch logo "$domain"
done
```

### Extract primary brand color

```bash
brandfetch colors stripe.com --output json | \
  jq -r '.colors[] | select(.type == "primary") | .hex'
```

### Search and get full brand info

```bash
domain=$(brandfetch search "Stripe" --output json | jq -r '.brands[0].domain')
brandfetch brand "$domain" --output json
```

### Fetch dark theme logos

```bash
brandfetch logo stripe.com --theme dark
brandfetch logo github.com --theme dark --format png
```

### Build a brand asset collection

```bash
# Download logo and extract colors for multiple companies
brands=("stripe.com" "github.com" "figma.com")

for domain in "${brands[@]}"; do
  company="${domain%.*}"
  brandfetch logo "$domain"
  brandfetch colors "$domain" --output json > "${company}-colors.json"
  echo "Processed $domain"
done
```

### JQ Filtering

```bash
# Pipe logo URL to clipboard (macOS)
brandfetch logo stripe.com --output json | jq -r '.url' | pbcopy

# Chain with other commands
brandfetch search "Stripe" --output json | jq '.brands[].domain' | \
  xargs -I {} brandfetch logo {}
```

## Global Flags

All commands support these flags:

- `--output <format>` - Output format: `text` or `json` (default: text)
- `--color <mode>` - Color mode: `auto`, `always`, or `never` (default: auto)
- `--help` - Show help for any command
- `--version` - Show version information

Logo command specific flags:
- `--format <format>` - Logo format: `svg`, `png`, or `webp` (default: svg)
- `--theme <theme>` - Logo theme: `light` or `dark` (default: light)
- `--type <type>` - Logo type: `logo`, `icon`, or `symbol` (default: logo)
- `--fallback <type>` - Fallback type: `lettermark`, `icon`, `symbol`, `brandfetch`, `404`
- `--width <px>` - Width in pixels
- `--height <px>` - Height in pixels

## Shell Completions

Generate shell completions for your preferred shell:

### Bash

```bash
# macOS (Homebrew):
brandfetch completion bash > $(brew --prefix)/etc/bash_completion.d/brandfetch

# Linux:
brandfetch completion bash > /etc/bash_completion.d/brandfetch

# Or source directly:
source <(brandfetch completion bash)
```

### Zsh

```zsh
brandfetch completion zsh > "${fpath[1]}/_brandfetch"
```

### Fish

```fish
brandfetch completion fish > ~/.config/fish/completions/brandfetch.fish
```

### PowerShell

```powershell
brandfetch completion powershell | Out-String | Invoke-Expression
```

## Development

After cloning, install git hooks:

```bash
make setup
```

This installs [lefthook](https://github.com/evilmartians/lefthook) pre-commit and pre-push hooks for linting and testing.

## License

MIT

## Links

- [Brandfetch API Documentation](https://developers.brandfetch.com/)
