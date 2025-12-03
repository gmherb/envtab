![banner](banner.png)
`envtab` (typed `envt\t`) aims to be your goto tool for working with environment variables. Organize sets of environment variables into loadouts. A loadout is a collection of environment variables that can be exported into the shell. Loadouts are named, optionally tagged, and can include a description. `envtab` stores these loadouts in your `$HOME` directory (`~/.envtab`), by default. `envtab` loadouts can also be enabled on shell login.

![diagram](diagram.png "Take control of your environment")

- [Installation](#installation)
- [Usage](#usage)
- [Environment Variables in Values](#environment-variables-in-values)
  - [PATH](#path)
  - [Environment variables other than PATH](#environment-variables-other-than-path)
- [Encrypting Sensitive Values](#encrypting-sensitive-values)
  - [Prerequisites](#prerequisites)
    - [Sops Configuration](#sops-configuration)
  - [Value Encryption](#value-encryption)
  - [File Encryption](#file-level-encryption)
  - [Viewing Decrypted Values](#viewing-decrypted-values)
  - [Automatic Decryption](#automatic-decryption)
  - [Editing Encrypted Loadouts](#editing-encrypted-loadouts)
- [Importing Loadouts and dotenv Files](#importing-loadouts-and-dotenv-files)
- [Generating CLI documentation](#generating-cli-documentation)
- [TODO](#todo)

# Installation

## From Source

### Prerequisites

- Go 1.25 or later
- Git (for version information)

### Build and Install

```bash
git clone https://github.com/gmherb/envtab.git
cd envtab; make install
```

*Installs to `/usr/local/bin/envtab`*

```bash
make build
./envtab --version
```

# Usage

Complete documentation for all `envtab` commands:

- [`envtab add`](docs/envtab_add.md) - Add an entry to a envtab loadout
- [`envtab cat`](docs/envtab_cat.md) - Concatenate envtab loadouts to stdout
- [`envtab edit`](docs/envtab_edit.md) - Edit envtab loadout
- [`envtab export`](docs/envtab_export.md) - Export envtab loadout(s)
- [`envtab import`](docs/envtab_import.md) - Import environment variables or loadouts
- [`envtab list`](docs/envtab_list.md) - List all envtab loadouts
- [`envtab login`](docs/envtab_login.md) - Export all login loadouts
- [`envtab make`](docs/envtab_make.md) - Make loadout from a template
- [`envtab remove`](docs/envtab_remove.md) - Remove envtab loadout(s)
- [`envtab show`](docs/envtab_show.md) - Show active loadouts

See also: [`envtab.md`](docs/envtab.md) for top-level usage and flags.

# Environment Variables in Values

Sometimes you may need to utilize environment variables in the value of a loadout entry. For example, the PATH environment variable.

## PATH

The PATH environment variable has first class support and will work without utilizing eval (shown below).

NOTE: To utilize multiple entries of the same KEY such as PATH, you must utilize multiple loadouts. A single loadout cannot have duplicate keys.

### Using add

If you utilize add, the environment variable will be subjected to shell variable/parameter expansion.

```text
$ envtab add testld PATH=$PATH:/other/bin
$ envtab cat testld | grep PATH
  PATH: /usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin:/usr/local/go/bin:/other/bin
```

Use must escape to bypass expansion.

```text
$ envtab add testld PATH=\$PATH:/other/bin
$ envtab cat testld | grep PATH
  PATH: $PATH:/other/bin
```

### Using edit

Or by editing the loadout configuration directly.

```text
$ envtab edit testld
----
metadata:
  createdAt: "2025-11-21T19:21:06-05:00"
  loadedAt: "2025-11-21T19:21:06-05:00"
  updatedAt: "2025-11-21T19:25:07-05:00"
  login: false
  tags: []
  description: ""
entries:
  PATH: $PATH:/other/bin
```

```text
$ envtab export testld
export PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin:/usr/local/go/bin:/other/bin

$ $(envtab export testld)
$ envtab show
testld -------------------------------------------------------------- [ 1 / 1 ]
   PATH=$PATH:/other/bin
```

## Environment variables other than PATH

Currently, PATH is the only officially support environment variable. You can use other environment variables using eval however, do not expect `envtab show` to work properly.

### Eval

```text
$ envtab cat example
metadata:
  createdAt: "2025-11-23T22:59:13-05:00"
  loadedAt: "2025-11-23T22:59:13-05:00"
  updatedAt: "2025-11-23T23:08:32-05:00"
  login: false
  tags: []
  description: ""
entries:
  CONFIG_DIR: $HOME/conf
```

```text
# Export shows the actual variable
$ envtab export example
export CONFIG_DIR=$HOME/conf

# But when sourced it is not expanded
$ $(envtab export example)

# See variable hardcoded
$ env|grep CONFIG_DIR
CONFIG_DIR=$HOME/conf

# Use eval to expand
$ eval $(envtab export example)

# Variable expanded
$ env|grep CONFIG_DIR
CONFIG_DIR=/home/gmherb/conf

# Unfortunately, no match in `show` or `list` at this time.
$ envtab show
$ envtab list -l example
UpdatedAt  LoadedAt  Login  Active  Total  Name     Tags
23:08:32   22:59:13  false  0       1      example  []
```

# Encrypting Sensitive Values

`envtab` supports encrypting sensitive values using SOPS (Secrets OPerationS). This allows you to securely store secrets like API keys, passwords, and tokens in your loadouts.

## Prerequisites

1. Install SOPS: https://github.com/getsops/sops
2. Configure SOPS with your preferred encryption backend (AWS KMS, GCP KMS, Azure Key Vault, age, PGP, etc.)
3. Set up your `.sops.yaml` configuration file (optional, but recommended)

### Sops Configuration

Configure SOPS by creating a `.sops.yaml` file in your project root or home directory:

```yaml
creation_rules:
  - path_regex: envtab-stdin-override
    kms: 'arn:aws:kms:us-east-1:123456789012:key/12345678-1234-1234-1234-123456789012'
    pgp: >-
      FBC7B9E2A4F9289AC0C1D4843D16CEE4A27381B4

  - kms: 'arn:aws:kms:us-east-1:123456789012:key/12345678-1234-1234-1234-123456789012'
    pgp: >-
      FBC7B9E2A4F9289AC0C1D4843D16CEE4A27381B4
```

For more details, see [SOPS_INTEGRATION.md](SOPS_INTEGRATION.md).

## Value Encryption

The `-e` or `--encrypt-value` flag encrypts individual values with SOPS:

```text
$ envtab add production -e API_KEY=sk_live_1234567890abcdef
$ envtab add production -e DB_PASSWORD=super_secret_password
```

When you view the loadout, encrypted values are hidden by default:

```text
$ envtab cat production
metadata:
  createdAt: "2025-11-26T00:00:00Z"
  ...
entries:
  API_KEY: SOPS:value: ENC[AES256_GCM,data:...,iv:...,tag:...,type:str]
    sops:
      gcp_kms: [...]
      ...
  DB_PASSWORD: SOPS:value: ENC[AES256_GCM,data:...,iv:...,tag:...,type:str]
    sops:
      gcp_kms: [...]
      ...
```

## File-Level Encryption

You can also encrypt entire loadout files with SOPS using the `--encrypt-file` flag (or `-f`):

```text
$ envtab add secrets --encrypt-file API_KEY=mykey DB_PASSWORD=mypass
```

This encrypts the entire file, including metadata. The file can be edited directly with `sops`:

```text
$ sops ~/.envtab/secrets.yaml
```

NOTE: File encryption will be faster if multiple encrypted values exist in a single loadout.

## Viewing Decrypted Values

To view decrypted sensitive values, use the `-d` or `--decrypt` flag with the `show` command:

```text
$ envtab show production
production -------------------------------------------------------- [ 2 / 2 ]
   API_KEY=***encrypted***
   DB_PASSWORD=***encrypted***

$ envtab show production -d
production -------------------------------------------------------- [ 2 / 2 ]
   API_KEY=sk_live_1234567890abcdef
   DB_PASSWORD=super_secret_password
```

## Automatic Decryption

Encrypted values are automatically decrypted when exporting:

```text
$ envtab export production
export API_KEY=sk_live_1234567890abcdef
export DB_PASSWORD=super_secret_password
```

## Editing Encrypted Loadouts

When editing a loadout with encrypted values, they are automatically decrypted for editing and re-encrypted when saved:

```text
$ envtab edit production
# Values appear as plaintext in the editor
# After saving, they are automatically re-encrypted
```

# Importing Loadouts and dotenv Files

envtab imports entire loadouts from .yaml files. It also can import variables from .env files.

## Import from local files (by extension)

  ```text
  # Merge a .env into an existing/new loadout
  envtab import myloadout ./config.env

  # Replace/create a loadout from YAML
  envtab import myloadout ./prod.yaml
  ```

## Import from remote URLs (e.g., GitHub raw)

  ```text
  # Merge .env from URL into an existing/new loadout
  envtab import myloadout --url https://raw.githubusercontent.com/org/repo/branch/config.env

  # Replace/create a loadout from YAML at a URL
  envtab import myloadout --url https://raw.githubusercontent.com/org/repo/branch/loadouts/prod.yaml
  ```

## Write a loadout YAML to a file for versioning in Git

  ```text
  envtab cat myloadout --output ./envtab-loadout-repo/myloadout.yaml
  ```

You can then commit and push these YAML files to GitHub or another Git host and share them with your team.

# Generating CLI documentation

This project includes a small tool that uses Cobra's `doc` package to generate Markdown docs for all commands.

- **Generate docs into the `docs/` directory**:

  ```text
  make docs
  ```

This runs `go run ./tools/gen-docs.go` and produces per-command Markdown files and a top-level `docs/envtab.md` that reflect the current CLI.

# TODO

- SOPS:exec-env - execute a command with decrypted values inserted into the environment
  - Add support or re-implement. Reimplementation would be best as can support all envtab loadouts
- Add --raw to loginCmd. This will place actual export entries inside of a shell script to be sourced from profile script instead of calling envtab.
  - Safer, faster, but lacks encryption at rest.
  - Also supports all environment values in entries as they will be evaluated on source.
  - However, syncing login requires checking for diffs after every edit, add, and import (make should have empty values from a template unless we are supporting values in templates)
    - sync can be manual for first implementation. `envtab login status --sync`
      - but end goal is `--sync` flag on edit, add, import, and make which triggers sync process.
      - or make it automatic for simplicity
  - --raw should be utilized with either --enable or --disable. Ignored if --status or enable/disable are omitted.
  - --status should now include mode (raw|command substitution)
- Support environment variables in show; exported with eval $(envtab export loadout)
  - Can we resolve all environment variables like we do with PATH? --raw is a workaround for using other environment variables
  - Add loadout order/priority/number to support specific load order in case entries build upon environment variable expansion
- Add additional backends in addition to default (file backend).
  - File (Default)
  - Vault
- Add ability to import/export various backends (import|export subCmd)
  - Vault, S3, GCS

## Done

- Fix the color output with show --all
- Support `--key` and `--value` in showCmd to locate specific vars without using `$(echo $KEY)` or `$(env|grep $VAL)`:
  - Show env var matching `--key`
  - Show env var matching `--value`
- Support `--all` to show all envtab entries. Not just those that are active.
- Allow passing filter/pattern arg to the listCmd. (done w/ glob)
- Add support for PATH environment variable (done)
- Fix show for PATH environment variable (done)
- Fix Active/Total spacing in `ls` output when counts are double, or triple digits. (done)
- Implement `-s|--sensitive` option to the addCmd to optionally encrypt values. (done)
  - Support: GCP KMS, AWS KMS, GPG(PGP)
  - Piggy back off sops? It already supports all providers
- In edit subcommand, ensure no duplicate keys (otherwise it will be overwritten) (done)
  - edit fails when loadout does not exist (fixed)
- Add support for templates with the mk command. Set user defined for matches before utilizing included templates.
  - templates can be .env files.
  - include templates for common things: AWS|GCP|Azure, Vault, GIT|GITHUB|GITLAB, PGSQL, etc.
- Support importing templates (.env)