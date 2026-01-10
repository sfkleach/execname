# Initial development

## Goal

To set-up the repository for the main task.

## Initial Task

Implement the skeleton structure for the execman console-application in Golang.
Use a full workspace structure with `cmd` and `pkg` folders. It should use cobra
for command line parsing. The subcommands are:

- version
- install
- info
- list
- check
- update
- remove

Only implement the version subcommand, the other subcommands should be simple
TBD stubs and will be elaborated in the next task. The usual help option 
should be added.

Exceptionally, `pathman` without a subcommand will simply print the usage
summary but if the long option `--version` is supplied then it works the
same way as any other application.
