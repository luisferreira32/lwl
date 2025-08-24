package main

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

type tokenType int

const (
	undefined tokenType = iota
	tvariable
	tconstant
	tlparenth
	trparenth
	tcomma
	add
	sub
	mul
	div
	mod
	eq
)

type token struct {
	t tokenType
	v string
}

func (t token) String() string {
	return t.v
}

func (t token) isOp() bool {
	return t.t == add || t.t == sub || t.t == mul || t.t == div || t.t == mod
}

func tokenFromRune(r rune) (token, error) {
	t := token{}
	t.v = string(r)
	switch r {
	case '=':
		t.t = eq
	case '+':
		t.t = add
	case '-':
		t.t = sub
	case '*':
		t.t = mul
	case '/':
		t.t = div
	case '%':
		t.t = mod
	case '(':
		t.t = tlparenth
	case ')':
		t.t = trparenth
	case ',':
		t.t = tcomma
	default:
		if r >= 'a' && r <= 'z' {
			t.t = tvariable
		} else if r >= '0' && r <= '9' {
			t.t = tconstant
		}
	}
	if t.t == undefined {
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
			p := line[0]
			pt, err := tokenFromRune(rune(p))
			if err != nil {
				f.errs = append(f.errs, err)
				continue
			}
			if !f.main && pt.t == tvariable {
				f.name = pt.v
			}
			f.tkns = append(f.tkns, pt)

			j := 0
			for {
				j++
				if j >= len(line) {
					break
				}

				// skip spaces
				if line[j] == ' ' {
					continue
				}

				// any constant might have multiple digits, so we need to parse them all
				if line[j] >= '0' && line[j] <= '9' && pt.t == tconstant {
					t := f.tkns[len(f.tkns)-1]
					t.v += string(line[j])
					f.tkns[len(f.tkns)-1] = t
					continue
				}

				t, err := tokenFromRune(rune(line[j]))
				if err != nil {
					f.errs = append(f.errs, err)
					continue
				}
				f.tkns = append(f.tkns, t)
			}
			functions = append(functions, f)
		}
	}

	return functions, nil
}
