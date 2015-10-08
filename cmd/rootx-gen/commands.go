package main

import (
	"fmt"
	"strings"
)

type command struct {
	name          string
	writeOnly     bool
	readInterface *string
	returnType    string
	codePattern   string
	mockPattern   string
}

var commands []command
var commandMap map[string]command

func initializeCommands() {
	instance := "instance"
	instances := "instances"

	commands = []command{
		{
			"exists",
			false,           // writes
			nil,             // read interface
			"(bool, error)", // return type
			"return rootx.Exists({{ var }}, \"{{ file }}\", {{ params }})",
			"return {{ var }}.Bool, {{ var }}.Error",
		},
		{
			"selectOne",
			false,     // writes
			&instance, // read interface
			"error",   // return type
			"return rootx.SelectOne({{ var }}, \"{{ file }}\", instance, {{ params }})",
			"instance = {{ var }}.Thing\nreturn {{ var }}.Error",
		},
		{
			"selectAll",
			false,      // writes
			&instances, // read interface
			"error",    // return type
			"return rootx.SelectAll({{ var }}, \"{{ file }}\", instances, {{ params }})",
			"instances = {{ var }}.Slice\nreturn {{ var }}.Error",
		},
		{
			"insert",
			true,             // writes
			nil,              // read interface
			"(int64, error)", // return type
			"return rootx.Insert({{ var }}, \"{{ file }}\", {{ params }})",
			"return {{ var }}.Int64, {{ var }}.Error",
		},
		{
			"updateOne",
			true,    // writes
			nil,     // read interface
			"error", // return type
			"return rootx.UpdateOne({{ var }}, \"{{ file }}\", {{ params }})",
			"return {{ var }}.Error",
		},
		{
			"deleteOne",
			true,    // writes
			nil,     // read interface
			"error", // return type
			"return rootx.DeleteOne({{ var }}, \"{{ file }}\", {{ params }})",
			"return {{ var }}.Error",
		},
		{
			"exec",
			false,   // writes
			nil,     // read interface
			"error", // return type
			"return rootx.Exec({{ var }}, \"{{ file }}\", {{ params }})",
			"return {{ var }}.Error",
		},
	}

	if psql {
		alterCommands(map[string]string{
			"insert": "return rootx.InsertPsql({{ var }}, \"{{ file }}\", {{ params }})",
		})
	}

	commandMap = make(map[string]command)
	for _, c := range commands {
		commandMap[c.name] = c
	}
}

type param struct {
	name string
	tipe string
}
type params []param

type invocation struct {
	cmd command

	file   string
	name   string
	params params
}

func parseCommand(file string, raw []string) (*invocation, error) {
	l, raw := raw[0], raw[1:]
	cmdName, funcName, err := split2(l)
	if err != nil {
		return nil, err
	}

	cmd, ok := commandMap[cmdName]
	if !ok {
		return nil, fmt.Errorf("Command %s is not defined", cmdName)
	}

	params, err := parseParams(raw)
	if err != nil {
		return nil, err
	}

	return &invocation{cmd, file, funcName, params}, nil
}

func parseParams(raw []string) (params, error) {
	var params []param
	for i, l := range raw {
		n, p, t, err := split3(l)
		if err != nil {
			return nil, err
		}
		if strings.TrimSpace(n) != fmt.Sprintf("$%d:", i+1) {
			return nil, fmt.Errorf("Parameter line is bad: '%s'", l)
		}
		p = strings.TrimSpace(p)
		t = strings.TrimSpace(t)
		params = append(params, param{p, t})
	}
	return params, nil
}

func (c *invocation) Signature() string {
	s := fmt.Sprintf("%s(", c.name)
	if c.cmd.readInterface != nil {
		s = s + *c.cmd.readInterface + " interface{}, "
	}
	s = s + c.params.signature()
	s = s + ") " + c.cmd.returnType
	return s
}

func (p params) names() string {
	var names []string
	for _, param := range p {
		names = append(names, param.name)
	}
	return strings.Join(names, ", ")
}

func (p params) signature() string {
	var params []string
	for _, param := range p {
		params = append(params, param.name+" "+param.tipe)
	}
	return strings.Join(params, ", ")
}

func alterCommands(alterations map[string]string) {
	for i, c := range commands {
		for k, v := range alterations {
			if c.name == k {
				c.codePattern = v
				commands[i] = c
			}
		}
	}
}
