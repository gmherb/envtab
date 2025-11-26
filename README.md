# envtab (`envt\t`)

![diagram](diagram.png "Take control of your environment")

`envtab` aims to be your goto tool for working with environment variables. Organize sets of environment variables into loadouts. A loadout is a collection of environment variables that can be exported into the shell. Loadouts are named, optionally tagged, and can include a description. `envtab` stores these loadouts in your `$HOME` directory (`~/.envtab`), by default.

`envtab` loadouts can also be enabled on shell login.

## Usage

```text
Take control of your environment.

Usage:
  envtab [command]

Available Commands:
  add         Add an entry to a envtab loadout
  cat         Concatenate envtab loadouts to stdout
  completion  Generate the autocompletion script for the specified shell
  edit        Edit envtab loadout
  export      Export envtab loadout
  help        Help about any command
  login       Export all login loadouts
  ls          List all envtab loadouts
  mk          Make loadout from a template
  rm          Remove envtab loadout(s)
  show        Show active loadouts

Flags:
  -h, --help      help for envtab
  -v, --version   version for envtab

Use "envtab [command] --help" for more information about a command.
```

To export a loadout into your current shell.

```text
$ $(envtab export aws-prd)
$ envtab show
aws-dev ------------------------------------------------------------- [ 1 / 3 ]
   AWS_DEFAULT_REGION=us-west-2

aws-prd ------------------------------------------------------------- [ 3 / 3 ]
   AWS_ACCESS_KEY_ID=AKIAIOSFODNN7EXAMPLE
   AWS_DEFAULT_REGION=us-west-2
   AWS_SECRET_ACCESS_KEY=wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY
```

*Note: The same key pair value can be set in different loadouts. envtab shows each matching loadout.*

To show the current state of login (enabled|disabled).

```text
$ envtab login --status
enabled
```

To enable envtab to export all login loadouts.

```text
$ envtab login --enable
```

To remove envtab from login shells.

```text
$ envtab login --disable
```

### Environment Variables in Values

Sometimes you may need to utilize environment variables in the value of a loadout entry. For example, the PATH environment variable.

#### PATH

The PATH environment variable has first class support and will work without utilizing eval (shown below).

NOTE: To utilize multiple entries of the same KEY such as PATH, you must utilize multiple loadouts. A single loadout cannot have duplicate keys.

##### add

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

##### edit

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
   PATH=$PATH:/some/bin
```

#### Environment variables other than PATH

Currently, PATH is the only officially support environment variable. You can use other envrionment variables using eval however, do not expect `envtab show` to work properly.

##### Eval

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
$ envtab ls -l example
UpdatedAt  LoadedAt  Login  Active  Total  Name     Tags
23:08:32   22:59:13  false  0       1      example  []
```

## Encrypting Sensitive Values

`envtab` supports encrypting sensitive values using SOPS (Secrets OPerationS). This allows you to securely store secrets like API keys, passwords, and tokens in your loadouts.

### Prerequisites

1. Install SOPS: https://github.com/getsops/sops
2. Configure SOPS with your preferred encryption backend (AWS KMS, GCP KMS, Azure Key Vault, age, PGP, etc.)
3. Set up your `.sops.yaml` configuration file (optional, but recommended)

### Using Value Encryption

The `-v` or `--encrypt-value` flag encrypts individual values with SOPS:

```text
$ envtab add production -v API_KEY=sk_live_1234567890abcdef
$ envtab add production -v DB_PASSWORD=super_secret_password
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

### Viewing Decrypted Values

To view decrypted sensitive values, use the `-s` flag with the `show` command:

```text
$ envtab show production
production -------------------------------------------------------- [ 2 / 2 ]
   API_KEY=***encrypted***
   DB_PASSWORD=***encrypted***

$ envtab show production -s
production -------------------------------------------------------- [ 2 / 2 ]
   API_KEY=sk_live_1234567890abcdef
   DB_PASSWORD=super_secret_password
```

### File-Level Encryption

You can also encrypt entire loadout files with SOPS using the `--encrypt-file` flag (or `-f`):

```text
$ envtab add secrets --encrypt-file API_KEY=mykey DB_PASSWORD=mypass
```

This encrypts the entire file, including metadata. The file can be edited directly with `sops`:

```text
$ sops ~/.envtab/secrets.yaml
```

### Automatic Decryption

Encrypted values are automatically decrypted when exporting:

```text
$ envtab export production
export API_KEY=sk_live_1234567890abcdef
export DB_PASSWORD=super_secret_password
```

### Editing Encrypted Loadouts

When editing a loadout with encrypted values, they are automatically decrypted for editing and re-encrypted when saved:

```text
$ envtab edit production
# Values appear as plaintext in the editor
# After saving, they are automatically re-encrypted
```

### Configuration

Configure SOPS by creating a `.sops.yaml` file in your project root or home directory:

```yaml
creation_rules:
  - path_regex: .*\.yaml$
    kms: 'arn:aws:kms:us-east-1:123456789012:key/12345678-1234-1234-1234-123456789012'
    pgp: >-
      FBC7B9E2A4F9289AC0C1D4843D16CEE4A27381B4
```

For more details, see [SOPS_INTEGRATION.md](SOPS_INTEGRATION.md).

## TODO

- Support environment variables in show; exported with eval $(envtab export loadout)
  - Can we resolve all environment variables like we do with PATH?
- Add loadout priority/number to support specific load order in case entries build upon environment variable expansion.
- Add support for templates with the mk command. Set user defined for matches before utilizing included templates.
  - templates can be .env files.
  - include templates for common things: AWS|GCP|Azure, Vault, GIT|GITHUB|GITLAB, PGSQL, etc.
- Add additional backends in addition to default (file backend).
  - File (Default)
  - Vault
- Add ability to import/export various backends (import|export subCmd)
  - First, support .env which will be used for templates.
  - Vault, S3, GCS

### Done

- Allow passing filter/pattern arg to the listCmd. (done w/ glob)
- Add support for PATH environemnt variable (done)
- Fix show for PATH environment variable (done)
- Fix Active/Total spacing in `ls` output when counts are double, or triple digits. (done)
- Implement `-s|--sensitive` option to the addCmd to optionally encrypt values. (done)
  - Support: GCP KMS, AWS KMS, GPG(PGP)
  - Piggy back off sops? It already supports all providers
- In edit subcommand, ensure no duplicate keys (otherwise it will be overwritten) (done)
  - edit fails when loadout does not exist (fixed)
