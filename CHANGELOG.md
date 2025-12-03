# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.1.11-alpha] - 2025-12-03

### Changed

- Refactored configuration handling:
  - Consolidated Viper configuration logic into `cmd/root.go` and `internal/config/config.go`
  - Simplified configuration initialization and path resolution
  - Improved code organization and maintainability

### Fixed

- Removed debug logging statement from `envtab show` command

## [0.1.10-alpha] - 2025-12-02

### Fixed

- Fixed `envtab show --all` command logic:
  - Corrected order of active entry checking to only occur when no key or value filter is applied
  - Fixed condition evaluation order to properly handle `--all` flag with filtering options

## [0.1.9-alpha] - 2025-12-02

### Changed

- Improved `envtab show --all` command display:
  - Added properly colored entry counts for better visibility when using `--all` / `-a` flag
  - Refactored color handling for improved consistency and maintainability

## [0.1.8-alpha] - 2025-12-02

### Added

- Enhanced `envtab show` command with filtering options:
  - Added `--key` / `-k` flag to show entries matching a specific key
  - Added `--value` / `-v` flag to show entries matching a specific value (supports both raw and SOPS-encrypted values)
  - Added `--all` / `-a` flag to show all entries in loadouts regardless of active status
  - Flags are mutually exclusive for clear behavior

### Changed

- Refactored SOPS and environment packages:
  - Consolidated SOPS display value logic into `SOPSDisplayValue` function in `pkg/sops`
  - Simplified environment comparison methods to use centralized SOPS display handling
  - Removed `DecryptFunc` type in favor of direct SOPS package integration
  - Improved code organization and maintainability

## [0.1.7-alpha] - 2025-12-02

### Added

- Test coverage for PATH expansion with SOPS-encrypted values:
  - Added `TestExportWithSOPSEncryptedPATH` to verify SOPS-encrypted PATH values are decrypted before PATH expansion (fixes from 0.1.4-alpha)
- Test coverage for empty value handling:
  - Added `TestExportWithEmptyValues` to verify proper handling of empty PATH entries, empty PATH segments, and empty non-PATH entries

## [0.1.6-alpha] - 2025-12-02

### Fixed

- Fixed PATH resolving in `loadout.Export()`:
  - `$PATH` variable expansion now correctly replaces `$PATH` with the actual current PATH value instead of an empty string
  - Fixed duplicate PATH export statement (PATH was being exported twice)
  - Fixed PATH handling to skip empty path entries when processing PATH values

## [0.1.5-alpha] - 2025-12-02

### Added

- SOPS stdin support for value encryption/decryption:
  - `SOPSEncryptValue` and `SOPSDecryptValue` now use stdin instead of temporary files
  - Added `--filename-override` flag support for stdin operations
  - Added `ENVTAB_SOPS_PATH_REGEX` environment variable to customize filename override (defaults to `envtab-stdin-override`)
- SOPS verbose mode support:
  - Added `SOPS_VERBOSE` environment variable to enable `--verbose` flag for all sops commands
- Updated SOPS configuration documentation:
  - Added path_regex patterns for `envtab-stdin-override` and catchall in configuration examples
  - Documented stdin filename override requirements

### Changed

- Improved SOPS value encryption/decryption efficiency by eliminating temporary file creation
- Updated SOPS integration to use stdin for all value-level operations

### Fixed

- Fixed syntax error in `buildSOPSArgs` function
- Removed sensitive value logging from debug statements

## [0.1.4-alpha] - 2025-12-02

### Fixed

- Fixed bug where temporary files were left behind after using `envtab edit` command
- Fixed `--remove-tags` flag to properly parse and remove tags from loadouts

### Changed

- Refactored export logic to evaluate PATH expansion for encrypted values (decrypts SOPS-encrypted PATH values before processing PATH expansion)
- Updated documentation examples for `--add-tags` and `--remove-tags` flags to use consistent comma-separated format

## [0.1.3-alpha] - 2025-12-02

### Fixed

- Fixed shell wildcard escaping in `envtab list` command examples in both code and documentation
- Updated README documentation with planning notes regarding --raw flag

### Security

## [0.1.2-alpha] - 2025-12-02

### Added

- `envtab edit` command now supports `--remove-entry` flag to remove entries from loadouts
- Comprehensive test suite for `envtab edit` command covering all flag combinations

## [0.1.1-alpha] - 2025-12-01

### Changed

- `envtab edit` command now uses `--add-tags` and `--remove-tags` flags instead of `--tags` for more granular tag management

## [0.1.0-alpha] - 2025-12-01

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
