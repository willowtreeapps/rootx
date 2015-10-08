### WillowTree is Hiring!

Want to write Go for mobile applications? Want to write anything else for mobile
applications? [Check out our openings!](http://willowtreeapps.com/careers/)

# RootX

Helper functions and code generation tools for https://github.com/jmoiron/sqlx.

Blog post coming soon!

## Installation

```
go get github.com/willowtreeapps/rootx/...
```

## Command Usage

```
$ rootx-gen -h
Usage of rootx-gen:
  -dir string
    	Directory containing sql files
  -dryRun
    	Output to STDOUT instead of writing files
  -formatter string
    	Command to use to format source code (gofmt, goimports) (default "goimports")
  -mode string
    	Mode for generator: code | mock | interface (default "code")
  -o string
    	Output file
  -pkg string
    	Go package to use
  -psql
    	Whether to use Postgres insert strategy, using "RETURNING id" (default true)
  -readType string
    	Type of instance to use for read methods
  -writeType string
    	Type of instance to use for write methods
```
