package main

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
)

const (
	mainFuncName = "main"
)

type tokenType int

const (
	tundefined tokenType = iota
	tvariable
	tconstant
	tlparenth
	trparenth
	tcomma
	tadd
	tsub
	tmul
	tdiv
	tmod
	teq
)

type token struct {
	t tokenType
	v string
	c int
}

func (t token) String() string {
	return t.v
}

func (t token) isOp() bool {
	return t.t == tadd || t.t == tsub || t.t == tmul || t.t == tdiv || t.t == tmod
}

func scanToken(r rune, c int) (token, error) {
	t := token{c: c}
	t.v = string(r)
	switch r {
	case '=':
		t.t = teq
	case '+':
		t.t = tadd
	case '-':
		t.t = tsub
	case '*':
		t.t = tmul
	case '/':
		t.t = tdiv
	case '%':
		t.t = tmod
	case '(':
		t.t = tlparenth
	case ')':
		t.t = trparenth
	case ',':
		t.t = tcomma
	default:
		if r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' {
			t.t = tvariable
		} else if r >= '0' && r <= '9' {
			t.t = tconstant
		}
	}
	if t.t == tundefined {
		// TODO: better explain our ways of writing
		return t, errors.New("invalid token: " + t.v)
	}
	return t, nil
}

type function struct {
	name string
	file string
	line int
	tkns []token
	main bool
	errs []error
}

func tokenize(files []string) ([]function, error) {
	functions := make([]function, 0)
	// - yes, given file order matters. no, we won't inform the user about it.
	for _, file := range files {
		contents, err := os.ReadFile(file)
		if err != nil {
			return nil, fmt.Errorf("read %v: %v", file, err)
		}
		// TODO: handle different line endings in different OS
		lines := strings.Split(string(contents), "\n")

		for i, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}
			// - every line is a function declaration, if there is no "=" it is the main function declaration
			// - we are chads and only support single character variable names
			f := function{
				file: file,
				line: i + 1,
				main: !strings.Contains(line, "="),
			}
			pt := token{}
			for j := 0; j < len(line); j++ {
				// skip spaces
				// TODO: proper whitespace list
				if line[j] == ' ' || line[j] == '\t' {
					continue
				}

				t, err := scanToken(rune(line[j]), j)
				if err != nil {
					f.errs = append(f.errs, err)
					continue
				}
				// any constant might have multiple digits, so include them in the previous token
				if t.t == tconstant && pt.t == tconstant {
					f.tkns[len(f.tkns)-1].v += t.v
					continue
				}
				f.tkns = append(f.tkns, t)
				pt = t
			}

			// add function names to simplify further logic
			if f.main {
				f.name = mainFuncName
			} else if !f.main && len(f.tkns) > 0 && f.tkns[0].t == tvariable {
				f.name = f.tkns[0].v
			} else {
				f.errs = append(f.errs, errors.New("useless line found: "+strconv.Itoa(f.line)))
			}
			functions = append(functions, f)
		}
	}

	return functions, nil
}
