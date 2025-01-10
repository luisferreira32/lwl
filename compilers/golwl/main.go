package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
)

var (
	// version: will be set during compilation onto the binary
	// DO NOT TOUCH IT! I WILL FIND YOU! AND I WILL HURT YOU!
	version string
)

type tokenType int

const (
	undefined tokenType = iota
	variable
	constant
	lparenth
	rparenth
	add
	sub
	mul
	div
	mod
	co
	eq
)

type token struct {
	t tokenType
	v string
}

func (t token) String() string {
	return t.v
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
		t.t = lparenth
	case ')':
		t.t = rparenth
	case ',':
		t.t = co
	default:
		if r >= 'a' && r <= 'z' {
			t.t = variable
		} else if r >= '0' && r <= '9' {
			t.t = constant
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

func parseFunctions(files []string) ([]function, error) {
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
				if line[j] >= '0' && line[j] <= '9' && pt.t == constant {
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
func main() {
	// TODO: logger with levels
	// TODO: implement checking the architecture of the host machine and restrict to amd64 linux only for now
	log.Printf("version: %v\n", version)

	// 1. parse input
	var output string
	flag.StringVar(&output, "o", "assembled.o", "output file name")

	files := os.Args[1:]
	if len(files) == 0 {
		log.Fatalf("no input files provided")
	}

	// 2. parse files
	functions, err := parseFunctions(files)
	if err != nil {
		log.Fatalf("%v", err)
	}

	// 3. handle syntax
	// TODO: gracefully handle syntax and semantic errors since they accumulate per function / line

	// 7. generate pseudo-assembly code
	// TODO: make this a pseudo-assembly and add optimized plugins for different architectures
	outputFile, err := os.Create(output)
	if err != nil {
		log.Fatalf("create %v: %v", output, err)
	}
	defer func() {
		if err := outputFile.Close(); err != nil {
			log.Fatalf("close %v: %v", output, err)
		}
	}()

	// hacks
	for _, f := range functions {
		outputFile.WriteString("# " + f.file + ":" + string(rune(f.line)) + "\n")
	}
}
