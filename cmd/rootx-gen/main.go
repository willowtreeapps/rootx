package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/willowtreeapps/acorn"
)

var (
	pkg    string
	dir    string
	output string
	mode   string

	readtype  string
	writetype string

	formatCommand string
	dryRun        bool
	psql          bool

	parser = acorn.NewParser("--", "!")
)

const (
	code     = "code"
	mock     = "mock"
	interfac = "interface"
)

func init() {
	flag.StringVar(&pkg, "pkg", "", "Go package to use")
	flag.StringVar(&dir, "dir", "", "Directory containing sql files")
	flag.StringVar(&output, "o", "", "Output file")
	flag.StringVar(&mode, "mode", "code", "Mode for generator: code | mock | interface")

	flag.StringVar(&readtype, "readType", "", "Type of instance to use for read methods")
	flag.StringVar(&writetype, "writeType", "", "Type of instance to use for write methods")

	flag.StringVar(&formatCommand, "formatter", "goimports", "Command to use to format source code (gofmt, goimports)")
	flag.BoolVar(&dryRun, "dryRun", false, "Output to STDOUT instead of writing files")

	flag.BoolVar(&psql, "psql", true, "Whether to use Postgres insert strategy, using \"RETURNING id\"")
}

func main() {
	flag.Parse()
	initializeCommands()

	if pkg == "" {
		log.Fatal("Package must be specified with -pkg name")
	}

	if dir == "" {
		log.Fatal("Directory must be specified with -dir path/to/dir")
	}

	if mode != code && mode != mock && mode != interfac {
		log.Fatal("Mode must be one of code | mock | interface")
	}

	var filename *string
	if !dryRun {
		if output == "" {
			log.Fatal("Output file must be specified with -o path/to/file")
		}
		filename = &output
	}

	var formatter *string
	if formatCommand != "" {
		formatter = &formatCommand
	}

	var err error
	var invocations []*invocation

	trimPath := func(s string) string {
		return strings.TrimPrefix(s, fmt.Sprintf("%s/", dir))
	}

	err = filepath.Walk(dir, func(path string, f os.FileInfo, err error) error {
		if !strings.HasSuffix(path, ".sql") {
			return nil
		}

		return parser.ParseFile(path, func(raw []string) {
			invocation, err := parseCommand(trimPath(path), raw)
			if err != nil {
				log.Printf("Could not parse command found in %v, %v", path, raw)
				log.Fatal(err)
			}
			invocations = append(invocations, invocation)
		})
	})

	if err != nil {
		log.Fatal(err)
	}

	acorn.Output(filename, formatter, generate(invocations))
}

func generate(invocations []*invocation) func(io.Writer) {
	return func(out io.Writer) {
		var w writer
		if mode == code {
			w = newCodeWriter(out)
		} else if mode == interfac {
			w = newInterfaceWriter(out)
		} else {
			w = newMockWriter(out)
		}

		out.Write([]byte(fmt.Sprintf("package %s\n\n", pkg)))
		w.start()
		for _, i := range invocations {
			w.handleCommand(i)
		}
		w.finish()
	}
}
