# envtab

![diagram](diagram.png "Take control of your environment")

`envtab` aims to be your goto tool for working with environment variables. Organize sets of environment variables into loadouts. A loadout is a collection of environment variables that can be exported into the shell. Loadouts are named, optionally tagged, and can include a description. `envtab` stores these loadouts in your `$HOME` directory (`~/.envtab`), by default. Note: Additional backends TBD..

`envtab` loadouts can also be enabled on shell login.

## Usage

```
$ ./envtab
Take control of your environment.

Usage:
  envtab [command]

Available Commands:
  add         Add an entry to a envtab loadout
  cat         Print an envtab loadout
  completion  Generate the autocompletion script for the specified shell
  delete      Delete envtab loadout(s)
  edit        Edit envtab loadout
  export      Export envtab loadout
  help        Help about any command
  list        List all envtab loadouts
  login       Export all login loadouts
  show        Show active loadouts

Flags:
  -h, --help      help for envtab
  -v, --version   version for envtab

Use "envtab [command] --help" for more information about a command.
```

To export a loadout into your current shell.

```
$ $(./envtab export my-essentials)
```

To show the current state of login (enabled|disabled).

```
$ ./envtab login --status
```

To enable envtab to export all login loadouts.

```
$ ./envtab login --enable
```

To remove envtab from login shells.

```
$ ./envtab login --disable
```

## TODO
- Implement `-s|--sensitive` option to the addCmd to optionally encrypt values.
  - Support: AES, AWS KMS, GPG(PGP)
- Add ability to create/use templates.
  - Create templates for most commonly used tools.
    - AWS, Vault, etc
- Add additional backends.
  - File (Default)
  - Vault
- Add ability to import/export various backends
