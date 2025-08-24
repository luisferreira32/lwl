package main

import (
	"bytes"
	"errors"
	"log"
	"strings"
	"testing"
)

func Test_parse(t *testing.T) {
	tests := []struct {
		name      string
		functions []function
		wantErr   error
		wantLog   string
	}{
		{
			name: "valid function and main",
			functions: []function{
				{
					name: "f",
					file: "test.lwl",
					line: 1,
					tkns: []token{
						{t: tvariable, v: "f"},
						{t: tlparenth, v: "("},
						{t: tvariable, v: "x"},
						{t: trparenth, v: ")"},
						{t: teq, v: "="},
						{t: tvariable, v: "x"},
					},
				},
				{
					name: "",
					file: "test.lwl",
					line: 2,
					tkns: []token{
						{t: tvariable, v: "f"},
						{t: tlparenth, v: "("},
						{t: tconstant, v: "1"},
						{t: trparenth, v: ")"},
					},
					main: true,
				},
			},
		},
		{
			name: "no main function",
			functions: []function{
				{
					name: "f",
					file: "test.lwl",
					line: 1,
					tkns: []token{
						{t: tvariable, v: "f"},
						{t: tlparenth, v: "("},
						{t: tvariable, v: "x"},
						{t: trparenth, v: ")"},
						{t: teq, v: "="},
						{t: tvariable, v: "x"},
					},
				},
			},
			wantErr: errNoMain,
		},
		{
			name: "multiple main functions",
			functions: []function{
				{
					name: "",
					file: "test1.lwl",
					line: 1,
					tkns: []token{
						{t: tvariable, v: "f"},
						{t: tlparenth, v: "("},
						{t: tconstant, v: "1"},
						{t: trparenth, v: ")"},
					},
					main: true,
				},
				{
					name: "",
					file: "test2.lwl",
					line: 1,
					tkns: []token{
						{t: tvariable, v: "g"},
						{t: tlparenth, v: "("},
						{t: tconstant, v: "2"},
						{t: trparenth, v: ")"},
					},
					main: true,
				},
			},
			wantErr: errMultipleMains,
		},
		{
			name: "duplicate function definition",
			functions: []function{
				{
					name: "f",
					file: "test.lwl",
					line: 1,
					tkns: []token{
						{t: tvariable, v: "f"},
						{t: tlparenth, v: "("},
						{t: tvariable, v: "x"},
						{t: trparenth, v: ")"},
						{t: teq, v: "="},
						{t: tvariable, v: "x"},
					},
				},
				{
					name: "f", // Same name as previous function
					file: "test.lwl",
					line: 2,
					tkns: []token{
						{t: tvariable, v: "f"},
						{t: tlparenth, v: "("},
						{t: tvariable, v: "y"},
						{t: trparenth, v: ")"},
						{t: teq, v: "="},
						{t: tvariable, v: "y"},
					},
				},
				{
					name: "",
					file: "test.lwl",
					line: 3,
					tkns: []token{
						{t: tvariable, v: "f"},
						{t: tlparenth, v: "("},
						{t: tconstant, v: "1"},
						{t: trparenth, v: ")"},
					},
					main: true,
				},
			},
			wantErr: errParse,
			wantLog: "already defined",
		},
		{
			name: "multiple equals in function",
			functions: []function{
				{
					name: "f",
					file: "test.lwl",
					line: 1,
					tkns: []token{
						{t: tvariable, v: "f"},
						{t: tlparenth, v: "("},
						{t: tvariable, v: "x"},
						{t: trparenth, v: ")"},
						{t: teq, v: "="},
						{t: tvariable, v: "x"},
						{t: teq, v: "="}, // Second equal sign
						{t: tconstant, v: "1"},
					},
				},
				{
					name: "",
					file: "test.lwl",
					line: 2,
					tkns: []token{
						{t: tvariable, v: "f"},
						{t: tlparenth, v: "("},
						{t: tconstant, v: "1"},
						{t: trparenth, v: ")"},
					},
					main: true,
				},
			},
			wantErr: errParse,
			wantLog: "multiple '='",
		},
		{
			name: "undefined variable",
			functions: []function{
				{
					name: "f",
					file: "test.lwl",
					line: 1,
					tkns: []token{
						{t: tvariable, v: "f"},
						{t: tlparenth, v: "("},
						{t: tvariable, v: "x"},
						{t: trparenth, v: ")"},
						{t: teq, v: "="},
						{t: tvariable, v: "y"}, // y is not defined
					},
				},
				{
					name: "",
					file: "test.lwl",
					line: 2,
					tkns: []token{
						{t: tvariable, v: "f"},
						{t: tlparenth, v: "("},
						{t: tconstant, v: "1"},
						{t: trparenth, v: ")"},
					},
					main: true,
				},
			},
			wantErr: errParse,
			wantLog: "undefined variable",
		},
		{
			name: "unexpected operator",
			functions: []function{
				{
					name: "f",
					file: "test.lwl",
					line: 1,
					tkns: []token{
						{t: tvariable, v: "f"},
						{t: tlparenth, v: "("},
						{t: tvariable, v: "x"},
						{t: trparenth, v: ")"},
						{t: teq, v: "="},
						{t: tadd, v: "+"}, // Operator with no preceding value
						{t: tvariable, v: "x"},
					},
				},
				{
					name: "",
					file: "test.lwl",
					line: 2,
					tkns: []token{
						{t: tvariable, v: "f"},
						{t: tlparenth, v: "("},
						{t: tconstant, v: "1"},
						{t: trparenth, v: ")"},
					},
					main: true,
				},
			},
			wantErr: errParse,
			wantLog: "unexpected operator",
		},
		{
			name: "main function not starting with variable or constant",
			functions: []function{
				{
					name: "",
					file: "test.lwl",
					line: 1,
					tkns: []token{
						{t: tadd, v: "+"}, // Should start with variable or constant
						{t: tconstant, v: "1"},
					},
					main: true,
				},
			},
			wantErr: errParse,
			wantLog: "must start with a variable or constant",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			b := bytes.Buffer{}
			originalOutput := log.Writer()
			defer log.SetOutput(originalOutput)
			log.SetOutput(&b) // TODO: make the logger parallel safe in unit tests
			err := parse(tc.functions)
			if !errors.Is(err, tc.wantErr) {
				t.Errorf("parse() error = %v, wantErr %v", err, tc.wantErr)
			}
			if !strings.Contains(b.String(), tc.wantLog) {
				t.Errorf("log output = %v, want to contain %v", b.String(), tc.wantLog)
			}
		})
	}
}
