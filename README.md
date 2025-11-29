# Azure Key Vault Manager

A command-line tool for managing Azure Key Vault secrets with an interactive interface powered by fzf.

## Features

- **Interactive Vault Selection**: Browse and select Azure Key Vaults from your subscription using fzf's fuzzy finder
- **Secret Management**: List, view, and manage secrets across multiple vaults
- **Version History**: View and manage all versions of secrets
- **Multi-Selection Support**: Perform operations on multiple secrets simultaneously
- **Rich Preview**: Preview secret details directly in fzf using Azure CLI integration
- **Formatted Tables**: Clean, formatted output for vaults and secrets

## Prerequisites

- Go 1.16 or later
- [Azure CLI](https://docs.microsoft.com/en-us/cli/azure/install-azure-cli) installed and authenticated
- [fzf](https://github.com/junegunn/fzf) installed and available in PATH
- Active Azure subscription with Key Vault access

## Installation

```bash
# Clone the repository
git clone <repository-url>
cd <repository-name>

# Build the project
go build -o azkvmgr .

# Run
./azkvmgr
```

### Using Docker

A Docker Compose setup is provided for a consistent development environment:

```bash
docker-compose up -d
docker-compose exec az-go-env bash
```

## Usage

Simply run the executable:

```bash
./azkvmgr
```

The tool will guide you through an interactive workflow:

1. **Select Vault(s)**: Choose one or more Key Vaults from your subscription
2. **Choose Operation**: Select the operation you want to perform (currently supports `list`)
3. **Select Secret(s)**: Browse and select secrets with live preview
4. **Perform Actions**: Based on your selection, choose from available operations:
   - **Single secret**: remove, show password, update metadata, update password, add new version
   - **Multiple secrets**: show password, update metadata

## Project Structure

```
.
├── asf.go              # Main application logic and workflow
├── tableformatter.go   # Generic table formatting utilities
├── vault.go            # Vault-specific table formatting
├── secret.go           # Secret-specific table formatting
├── fzf.go              # fzf integration and selection logic
├── Dockerfile          # Docker environment setup
└── docker-compose.yaml # Docker Compose configuration
```

## Features in Detail

### Vault Management

- Lists all Key Vaults in your Azure subscription
- Displays: ID, Name, Tags, Resource Group, Location
- Supports multi-selection for batch operations

### Secret Management

- Lists all secrets and their versions across selected vaults
- Displays: ID, Name, Version, Enabled status
- Live preview of secret details using Azure CLI
- Version history tracking

### Interactive Interface

The tool leverages fzf to provide:

- Fuzzy search across all displayed fields
- Multi-selection with Tab key
- Live preview of secret details
- Keyboard-driven navigation

## Authentication

The tool uses Azure Default Credentials, which attempts authentication in the following order:

1. Environment variables
2. Managed Identity
3. Azure CLI credentials
4. Interactive browser login

Ensure you're logged in via Azure CLI:

```bash
az login
az account set --subscription <subscription-id>
```

## Development Roadmap

TODO:

- State machine architecture for operation stacking with backtracking support
- Enhanced operations: list with/without passwords
- Value binding for fetching and preview display
- Data streaming capabilities
- Client-server architecture using Unix sockets for external triggers
- Additional multi-select operations

## Dependencies

- `github.com/Azure/azure-sdk-for-go/sdk/azidentity` - Azure authentication
- `github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/keyvault/armkeyvault` - Key Vault management
- `github.com/Azure/azure-sdk-for-go/sdk/security/keyvault/azsecrets` - Secret operations
- `fzf` - Interactive filtering and selection

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

MIT

## Acknowledgments

- [fzf](https://github.com/junegunn/fzf) for the excellent fuzzy finder
- [Azure SDK for Go](https://github.com/Azure/azure-sdk-for-go) for Azure integration
