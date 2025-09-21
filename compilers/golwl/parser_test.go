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
		name        string
		functions   []function
		wantErr     error
		wantLog     string
		wantOpTrees map[string]*operation
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
			wantOpTrees: map[string]*operation{
				mainFuncName: {
					op: token{v: "f", t: tvariable},
					v: map[token]token{
						{v: "x", t: tvariable}: {v: "1", t: tconstant},
					},
				},
				"f": {
					op: token{v: "x", t: tvariable},
				},
			},
		},
		{
			name: "complex valid function and main",
			functions: []function{
				{
					name: "f",
					file: "test.lwl",
					line: 1,
					tkns: []token{
						{t: tvariable, v: "f"},
						{t: tlparenth, v: "("},
						{t: tvariable, v: "x"},
						{t: tcomma, v: ","},
						{t: tvariable, v: "y"},
						{t: trparenth, v: ")"},
						{t: teq, v: "="},
						{t: tvariable, v: "x"},
						{t: tadd, v: "+"},
						{t: tconstant, v: "2"},
						{t: tmul, v: "*"},
						{t: tvariable, v: "x"},
						{t: tmul, v: "*"},
						{t: tconstant, v: "2"},
						{t: tadd, v: "+"},
						{t: tvariable, v: "y"},
						{t: tdiv, v: "/"},
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
						{t: tcomma, v: ","},
						{t: tconstant, v: "1"},
						{t: trparenth, v: ")"},
					},
					main: true,
				},
			},
			wantOpTrees: map[string]*operation{
				mainFuncName: {
					op: token{v: "f", t: tvariable},
					v: map[token]token{
						{v: "x", t: tvariable}: {v: "1", t: tconstant},
						{v: "y", t: tvariable}: {v: "1", t: tconstant},
					},
				},
				"f": {
					op: token{v: "+", t: tadd},
					p: [2]*operation{
						{
							op: token{v: "+", t: tadd},
							p: [2]*operation{
								{op: token{t: tvariable, v: "x"}},
								{
									op: token{t: tmul, v: "*"},
									p: [2]*operation{
										{
											op: token{t: tmul, v: "*"},
											p: [2]*operation{
												{op: token{t: tconstant, v: "2"}},
												{op: token{t: tvariable, v: "x"}},
											},
										},
										{op: token{t: tconstant, v: "2"}},
									},
								},
							},
						},
						{
							op: token{v: "/", t: tdiv},
							p: [2]*operation{
								{op: token{t: tvariable, v: "y"}},
								{op: token{t: tvariable, v: "x"}},
							},
						},
					},
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
			opTrees, err := parse(tc.functions)
			if !errors.Is(err, tc.wantErr) {
				t.Errorf("parse() error = %v, wantErr %v", err, tc.wantErr)
			}
			if !strings.Contains(b.String(), tc.wantLog) {
				t.Errorf("log output = %v, want to contain %v", b.String(), tc.wantLog)
			}
			for fname := range tc.wantOpTrees {
				opTree, ok1 := opTrees[fname]
				wantOpTree, ok2 := tc.wantOpTrees[fname]
				if !ok1 || !ok2 {
					t.Fatalf("did not find expected op tree for function: %s", fname)
				}
				cmpOpTree(t, fname, opTree, wantOpTree, 0)
			}
		})
	}
}

func cmpOpTree(t *testing.T, fname string, op *operation, wantOp *operation, l int) {
	if op == nil && wantOp == nil {
		return
	}
	if op == nil || wantOp == nil {
		t.Errorf("%s(%d) unexpected op, want/got:\n%v\n%v\n", fname, l, wantOp, op)
		return
	}
	if op.op != wantOp.op {
		t.Errorf("%s(%d) unexpected op, want/got:\n%v\n%v\n", fname, l, wantOp, op)
		return
	}
	cmpOpTree(t, fname, op.p[0], wantOp.p[0], l+1)
	cmpOpTree(t, fname, op.p[1], wantOp.p[1], l+1)
}
