# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.1.0-alpha] - 2025-11-30

### Added

- Core loadout management commands:
  - `envtab add` - Add entries to loadouts
  - `envtab edit` - Edit loadout configuration files
  - `envtab remove` - Remove loadouts
  - `envtab list` - List all loadouts with filtering support
  - `envtab show` - Show active loadouts in current environment
  - `envtab cat` - Display loadout contents
  - `envtab export` - Export loadouts as shell-compatible export statements
- Loadout import functionality:
  - Import from local `.env` and `.yaml` files. Both create loadout if missing; when existing:
    - .env is merged to existing loadouts
    - .yaml will replace if existing
  - Import from remote URLs (e.g., GitHub raw files)
- Loadout export functionality:
  - Export to stdout for shell sourcing
  - To dump the raw loadout (YAML with encryption if used), use cat without -d|--decrypt flag.
- Template system with `envtab make` command:
  - Pre-built templates for cloud providers (AWS, GCP, Azure, OpenStack)
  - Database templates (PostgreSQL, MySQL, MongoDB, Elasticsearch)
  - Message queue templates (Kafka, RabbitMQ)
  - Cache templates (Redis, Memcached)
  - Container templates (Docker, Kubernetes)
  - Secrets management templates (Vault, Consul)
  - Infrastructure tools (Terraform, Terragrunt, Helm, Ansible, Packer, Vagrant)
  - Language templates (Python, Go, Rust, C)
  - VCS templates (Git, GitHub, GitLab)
  - Network templates (Proxy, WireGuard)
  - Utility templates (SOPS, yq, jq, jo, etcd, k6, Jira CLI)
  - Support for custom user templates in `~/.envtab/templates/`
- SOPS encryption integration:
  - Value-level encryption with `--encrypt-value` / `-e` flag
  - File-level encryption with `--encrypt-file` / `-f` flag
  - Automatic decryption on export
  - Decrypted view with `--decrypt` / `-d` flag in show command
  - Support for AWS KMS, GCP KMS, Azure Key Vault, age, PGP backends
- PATH environment variable support:
  - First-class support for PATH expansion
  - Automatic resolution in export and show commands
- Login loadouts with `envtab login`:
  - Mark loadouts for automatic export on shell login
  - Export all login-enabled loadouts
- Loadout metadata:
  - Tags for organizing loadouts
  - Descriptions for loadouts
  - Creation, update, and load timestamps
- Configuration system:
  - YAML-based configuration file support
  - Environment variable configuration
  - Configurable log levels
  - Verbose mode with `--verbose` / `-v` flag
- CLI documentation generation:
  - `make docs` target to generate Markdown documentation
  - Per-command documentation in `docs/` directory
- Makefile targets:
  - `make build` - Build the binary
  - `make install` - Install to system
  - `make test` - Run tests with coverage
  - `make docs` - Generate CLI documentation
  - `make version` - Display version information
- Build-time version management:
  - Version information from git tags
  - Commit hash and build date included in version string
  - `make version` target to display version information

### Changed

- N/A (Initial release)

### Deprecated

- N/A (Initial release)

### Removed

- N/A (Initial release)

### Fixed

- N/A (Initial release)

### Security

- SOPS integration for secure storage of sensitive environment variables
- Support for multiple encryption backends (KMS, age, PGP)
