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
envtab login --enable
```

To remove envtab from login shells.

```text
envtab login --disable
```

## TODO

- Fix Active/Total spacing in `ls` output when counts are double, or triple digits.
- Allow passing filter/pattern arg to the listCmd.
- Create templates for most commonly used tools.
  - AWS, Vault, etc
  - Check for predefined if no user defined templates match.
- Implement `-s|--sensitive` option to the addCmd to optionally encrypt values.
  - Support: AES, AWS KMS, GPG(PGP)
- Add additional backends.
  - File (Default)
  - Vault
- Add ability to import/export various backends
