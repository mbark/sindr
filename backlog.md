## Backlog

- [x] Allow "importing" package.json files and adding their scripts as commands
- [x] Add functionality for setting categories
- [x] Remove the watch stuff
- [x] lib:ify the `sindr` package to allow people to use and extend it through Go.
- [x] Add the newest_ts and oldest_ts functions back
- [x] When running a command, show it in some nice way.
- [x] Add an exec command that can be used to run some arbitrary programming language, similar
  to [shebang recipes](https://github.com/casey/just?tab=readme-ov-file#shebang-recipes)
- [x] Add a way to automatically source dotenv files
  like [just](https://github.com/casey/just?tab=readme-ov-file#dotenv-settings)
- [x] Move all the commands in the `sindr` "namespace" to be global functions
- [x] Generate some logo for this.
- [x] Rename everything `shmake` to `sindr`
- [x] Allow both flags and args to be added as simple strings.
- [x] Check why individual unit tests are failing when all pass
- [x] Allow configuring things like name of file and color via a config file (or global flag).
- [x] Allow the name of the global option be set via a command-line flag
- [x] Map the pflags to cli.Flags automatically
- [ ] Update the `README` with more context, comparisons to `just`, `mise`, `Makefile` etc.
- [ ] Tag a version 0.0.2 to build a binary to test `goreleaser` stuff.
- [ ] Improve the tests for `packagejson` and command to actually check the commands are added (e.g., via --help)
