# ASF (Azure Secret Finder)

A command-line tool for interacting with Azure Key Vault secrets through an intuitive fuzzy-finding interface powered by fzf.

## Overview

ASF provides a streamlined workflow for discovering and managing secrets across Azure Key Vaults. It combines the power of the Azure SDK with fzf's fuzzy search capabilities to create an efficient, interactive experience for working with sensitive data.

## Features

- **Multi-Vault Support**: Browse and select from all Key Vaults in your Azure subscription
- **Interactive Secret Selection**: Use fzf's fuzzy search to quickly find secrets across vaults
- **Version Management**: List and explore all versions of a secret
- **Password Retrieval**: Securely fetch and display secret values
- **Tag-Based Filtering**: Filter vaults and secrets by tags for better organization
- **Multiple Operations**:
  - List secret versions
  - Get secret passwords
  - List versions and retrieve passwords in one flow
  - Edit metadata (planned)
  - Delete secrets (planned)

## Prerequisites

- Go 1.19 or later
- Azure CLI installed and configured (`az login`)
- Active Azure subscription with Key Vault access
- Appropriate Azure RBAC permissions (Key Vault Secrets Officer or equivalent)

## Installation

### Option 1: Install with Go

```bash
go install github.com/braunbearded/asf/cmd/asf@latest
```

### Option 2: Build from Source

```bash
git clone https://github.com/braunbearded/asf.git
cd asf
go build -o asf ./cmd/asf
```

### Option 3: Docker (Recommended for Development)

The project includes a Docker setup with all dependencies pre-configured:

```bash
# Start the container
docker-compose up -d

# Enter the container
docker-compose exec asf bash

# Inside the container, build and run
go build -o asf ./cmd/asf
./asf
```

The Docker environment includes:

- Go 1.23.0
- Azure CLI
- fzf
- All necessary build tools

Your Azure CLI credentials are persisted in `./docker-files/az-cli` and mounted into the container.

## Usage

Simply run the tool:

```bash
asf
```

### Workflow

1. **Select Vaults**: Choose one or more Key Vaults from your subscription using fzf
2. **Select Secrets**: Browse and select secrets from the chosen vaults
3. **Choose Operation**: Pick an action to perform on the selected secrets
4. **View Results**: See the output and optionally perform additional operations

### Available Operations

- **get-passwords**: Retrieve the current value of selected secrets
- **list-versions**: Display all versions of selected secrets
- **list-version-get-password**: List all versions and retrieve their passwords
- **edit-meta**: Edit metadata for selected secrets (coming soon)
- **delete-secret**: Remove a secret and all its versions (coming soon)

### Navigation

- Use arrow keys or fuzzy search to filter items
- Press `Tab` to select multiple items
- Press `Enter` to confirm selection
- Press `Esc` or `Ctrl+C` to exit

## Configuration

ASF uses your default Azure CLI authentication and subscription.

### First-time Setup

```bash
# Login to Azure
az login

# Set your default subscription (if needed)
az account set --subscription <subscription-id>
```

### Docker Setup

When using Docker, you need to authenticate Azure CLI inside the container:

```bash
# Enter the container
docker-compose exec asf bash

# Login to Azure (only needed once)
az login

# Your credentials will be persisted in ./docker-files/az-cli
```

To use a different subscription:

```bash
az account set --subscription <subscription-id>
```

## Display Format

Secrets are displayed with the following information:

- Secret name
- Password value (masked as `******` until retrieved)
- Version
- Vault name
- Tags
- Created timestamp
- Enabled status

## Security Considerations

- Passwords are only fetched when explicitly requested
- Secret values are displayed in the terminal - ensure you're working in a secure environment
- ASF respects Azure RBAC permissions - you can only access secrets you have rights to

## Dependencies

- [Azure SDK for Go](https://github.com/Azure/azure-sdk-for-go)
- [fzf](https://github.com/junegunn/fzf) (embedded)
- [Azure Identity](https://github.com/Azure/azure-sdk-for-go/sdk/azidentity)

## Development

### Local Development

```bash
# Build
go build -o asf ./cmd/asf

# Run locally
go run ./cmd/asf
```

### Docker Development Environment

The project includes a complete Docker development environment:

```bash
# Build and start the container
docker-compose up -d

# Enter the development environment
docker-compose exec asf bash

# Inside the container, you have access to:
# - Go 1.23.0
# - Azure CLI
# - fzf
# - All source code mounted at /workspace

# Build and test
go build -o asf ./cmd/asf
go test ./...

# Stop the container
docker-compose down
```

The Docker setup uses:

- Debian Bookworm base image
- Non-root user (UID 1000, GID 1000)
- Persistent Azure CLI credentials in `./docker-files/az-cli`
- Volume mount for live code editing

## Project Structure

```
.
├── cmd/asf/              # Main entry point
├── internal/             # Internal packages
│   ├── vault.go          # Vault operations
│   ├── secret.go         # Secret operations
│   ├── selector.go       # fzf integration
│   ├── operation.go      # Operation definitions
│   └── subscription.go   # Azure subscription handling
├── docker-files/         # Docker-related files
│   └── az-cli/          # Persistent Azure CLI credentials
├── asf.go                # Core application logic
├── Dockerfile            # Docker image definition
├── docker-compose.yaml   # Docker Compose configuration
└── go.mod                # Go module definition
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

GPL-3.0

## Author

braunbearded

## Acknowledgments

- Built with [fzf](https://github.com/junegunn/fzf) by Junegunn Choi
- Uses the [Azure SDK for Go](https://github.com/Azure/azure-sdk-for-go)
