# envtab

![diagram](diagram.png "Take control of your environment")

`envtab` aims to be your goto tool for working with environment variables. Organize sets of environment variables into loadouts. A loadout is a collection of environment variables that can be exported into the shell. Loadouts are named, optionally tagged, and can include a description. `envtab` stores these loadouts in your `$HOME` directory (`~/.envtab`), by default. Note: Additional backends TBD..

`envtab` loadouts can also be enabled on shell login.

## Usage

```shell
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
  rm          Remove envtab loadout(s)
  show        Show active loadouts

Flags:
  -h, --help      help for envtab
  -v, --version   version for envtab

Use "envtab [command] --help" for more information about a command.
```

To export a loadout into your current shell.

```shell
$ $(envtab export aws-prd)
$ envtab show
aws-dev ---------------------------------------------------------------------------------- [ 1 / 3 ]
   AWS_DEFAULT_REGION=us-west-2

aws-prd ---------------------------------------------------------------------------------- [ 3 / 3 ]
   AWS_SECRET_ACCESS_KEY=wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY
   AWS_ACCESS_KEY_ID=AKIAIOSFODNN7EXAMPLE
   AWS_DEFAULT_REGION=us-west-2

```

*Note: The same key pair value can be set in different loadouts. envtab shows each matching loadout.*

To show the current state of login (enabled|disabled).

```shell
$ envtab login --status
enabled
```

To enable envtab to export all login loadouts.

```shell
envtab login --enable
```

To remove envtab from login shells.

```shell
envtab login --disable
```

## TODO

- Implement `new` subcmd which creates a loadout.
  - Optionally, specify a template with preset Keys.
- Create templates for most commonly used tools.
  - AWS, Vault, etc
- Implement `-s|--sensitive` option to the addCmd to optionally encrypt values.
  - Support: AES, AWS KMS, GPG(PGP)
- Add additional backends.
  - File (Default)
  - Vault
- Add ability to import/export various backends
