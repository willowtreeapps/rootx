package main

import (
	"fmt"
	"strings"
)

func split2(line string) (string, string, error) {
	x := strings.Split(line, " ")
	if len(x) != 2 {
		return "", "", fmt.Errorf("Bad line: '%s'; expected 2 components, got %d", line, len(x))
	}
	return x[0], x[1], nil
}

func split3(line string) (string, string, string, error) {
	x := strings.Split(line, " ")
	if len(x) != 3 {
		return "", "", "", fmt.Errorf("Bad line: '%s'; expected 3 components, got %d", line, len(x))
	}
	return x[0], x[1], x[2], nil
}
