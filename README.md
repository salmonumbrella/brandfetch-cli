# 🎨 Brandfetch CLI — Brand assets in your terminal.

Fetch logos, colors, and fonts for any company.

## Features

- **Brand assets** - download logos in SVG/PNG format, extract colors, and get fonts
- **Colors & fonts** - extract brand color palettes and typography information
- **Search** - find brands by name or keyword
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
brandfetch logo <domain>                 # Download logo (SVG, light theme)
brandfetch logo <domain> --format png    # Download PNG format
brandfetch logo <domain> --theme dark    # Get dark theme variant
brandfetch logo <domain> --output json   # Get logo metadata as JSON
```

### Brand

```bash
brandfetch brand <domain>                # Get comprehensive brand data
brandfetch brand <domain> --output json  # Get brand data as JSON
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
brandfetch colors <domain>               # Get brand color palette
brandfetch colors <domain> --output json # Colors as JSON
```

**Note**: Requires Brand API key (limited quota)

### Fonts

```bash
brandfetch fonts <domain>                # Get brand typography
brandfetch fonts <domain> --output json  # Fonts as JSON
```

**Note**: Requires Brand API key (limited quota)

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
    {"hex": "#635BFF", "type": "primary", "brightness": "dark"}
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
- `--format <format>` - Logo format: `svg` or `png` (default: svg)
- `--theme <theme>` - Logo theme: `light` or `dark` (default: light)

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
