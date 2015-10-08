package main

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"text/template"

	"github.com/willowtreeapps/acorn"
)

type writer interface {
	start()
	handleCommand(*invocation)
	finish()
}

type codeWriter struct {
	writer io.Writer
}

type codeCommand struct {
	*invocation
	tipe        string
	bodyPattern string
}

type mockWriter codeWriter
type mockCommand codeCommand

func newCodeWriter(w io.Writer) *codeWriter {
	return &codeWriter{w}
}

func (w *codeWriter) start()  {}
func (w *codeWriter) finish() {}

func (w *codeWriter) handleCommand(i *invocation) {
	if !i.cmd.writeOnly {
		acorn.WriteFunction(w.writer, &codeCommand{i, readtype, i.cmd.codePattern})
	}
	acorn.WriteFunction(w.writer, &codeCommand{i, writetype, i.cmd.codePattern})
}

func (c *codeCommand) Type() string {
	return c.tipe
}

func (c *codeCommand) Body() string {
	fmap := map[string]interface{}{
		"var": func() string {
			s := strings.TrimLeft(c.tipe, "(")
			return strings.Split(s, " ")[0]
		},
		"file":   func() string { return c.file },
		"params": c.params.names,
	}
	templ, err := template.New(c.bodyPattern).Funcs(fmap).Parse(c.bodyPattern)
	if err != nil {
		return fmt.Sprintf("ERROR %v", err)
	}
	var b bytes.Buffer
	templ.Execute(&b, nil)
	return b.String()
}

func newMockWriter(w io.Writer) *mockWriter {
	return &mockWriter{w}
}

func (w *mockWriter) start()  {}
func (w *mockWriter) finish() {}

func (w *mockWriter) handleCommand(i *invocation) {
	if !i.cmd.writeOnly {
		acorn.WriteFunction(w.writer, &codeCommand{i, readtype, i.cmd.mockPattern})
	}
	acorn.WriteFunction(w.writer, &codeCommand{i, writetype, i.cmd.mockPattern})
}

type interfaceWriter struct {
	codeWriter
	readCommands  []string
	writeCommands []string
}

func newInterfaceWriter(w io.Writer) *interfaceWriter {
	return &interfaceWriter{
		codeWriter:    codeWriter{w},
		readCommands:  make([]string, 0),
		writeCommands: make([]string, 0),
	}
}

func (w *interfaceWriter) handleCommand(i *invocation) {
	if i.cmd.writeOnly {
		w.writeCommands = append(w.writeCommands, i.Signature()+"\n")
	} else {
		w.readCommands = append(w.readCommands, i.Signature()+"\n")
	}
}

type namedInterface struct {
	name       string
	signatures []string
}

func (n *namedInterface) Name() string {
	return n.name
}

func (n *namedInterface) Signatures() []string {
	return n.signatures
}

func (w *interfaceWriter) finish() {
	acorn.WriteInterface(w.writer, &namedInterface{readtype, w.readCommands})
	acorn.WriteInterface(w.writer, &namedInterface{writetype, w.writeCommands})
}
