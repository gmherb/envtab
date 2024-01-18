# envtab

![diagram](diagram.png "Take control of your environment")

`envtab` aims to be your goto tool for working with environment variables. Organize, tag, list, search, export, and enable (on login) environment loadouts.

## Usage

```
$ ./envtab
envtab is a tool used for organizing environment variables into sets
called loadouts. A loadout is a collection of environment variables that can be
exported into the current shell. Each loadout has a name and can have a
description along with tags which can be used to filter/search.

Usage:
  envtab [command]

Available Commands:
  add         Add an envtab entry to a loadout
  cat         Print an envtab loadout
  completion  Generate the autocompletion script for the specified shell
  delete      Delete loadout(s)
  edit        Edit envtab loadout
  export      Export a loadout
  help        Help about any command
  list        List all envtab loadouts
  show        Show active loadouts

Flags:
  -h, --help   help for envtab

Use "envtab [command] --help" for more information about a command.
```

## TODO

- Add ability to create/use templates.
  - Create templates for most commonly used tools.
    - AWS, Vault, etc
- Implement `-s|--sensitive` option to the addCmd to optionally encrypt values.
  - Support: AWS KMS, GPG(PGP)
- Add ability to populate .bashrc or similar with loadouts with `login:true`.
  - Show conflicts when two or more loadouts with the same environment variable is set.
  - Potentially have weights/priority to deal with conflicts?
- Add additional backends.
- Add ability to import/export to various backends (e.g. vault)
