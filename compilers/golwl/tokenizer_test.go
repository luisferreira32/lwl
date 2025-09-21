package main

import (
	"errors"
	"os"
	"path/filepath"
	"slices"
	"testing"
)

func Test_parseFiles(t *testing.T) {
	tests := []struct {
		name      string
		files     map[string]string // filename -> content
		wantFuncs []function
		wantErr   error
	}{
		{
			name: "simple function declaration and call",
			files: map[string]string{
				"test.lwl": "f(x,y)=x+y\nf(1,2)\n",
			},
			wantFuncs: []function{
				{
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
						{t: tvariable, v: "y"},
					},
					name: "f",
					main: false,
				},
				{
					line: 2,
					tkns: []token{
						{t: tvariable, v: "f"},
						{t: tlparenth, v: "("},
						{t: tconstant, v: "1"},
						{t: tcomma, v: ","},
						{t: tconstant, v: "2"},
						{t: trparenth, v: ")"},
					},
					main: true,
				},
			},
		},
		{
			name: "empty file",
			files: map[string]string{
				"empty.lwl": "",
			},
			wantFuncs: []function{},
		},
		{
			name: "simple addition in main",
			files: map[string]string{
				"addition.lwl": "1+3+1\n",
			},
			wantFuncs: []function{
				{
					line: 1,
					file: "addition.lwl",
					tkns: []token{
						{t: tconstant, v: "1"},
						{t: tadd, v: "+"},
						{t: tconstant, v: "3"},
						{t: tadd, v: "+"},
						{t: tconstant, v: "1"},
					},
					main: true,
				},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Create temporary directory
			tempDir := t.TempDir()

			// Create test files
			var filePaths []string
			for filename, content := range tc.files {
				filePath := filepath.Join(tempDir, filename)
				if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
					t.Fatalf("failed to write test file %s: %v", filename, err)
				}
				filePaths = append(filePaths, filePath)
			}

			// Run parseFiles
			got, err := tokenize(filePaths)

			if !errors.Is(err, tc.wantErr) {
				t.Errorf("parseFiles() error = %v, wantErr %v", err, tc.wantErr)
			}

			if len(tc.wantFuncs) != len(got) {
				t.Fatalf("parseFiles() got %d functions, want %d", len(got), len(tc.wantFuncs))
			}

			for i := range tc.wantFuncs {
				if tc.wantFuncs[i].name != got[i].name {
					t.Errorf("function %d name = %v, want %v", i, got[i].name, tc.wantFuncs[i].name)
				}
				if tc.wantFuncs[i].line != got[i].line {
					t.Errorf("function %d line = %v, want %v", i, got[i].line, tc.wantFuncs[i].line)
				}
				if !slices.EqualFunc(
					tc.wantFuncs[i].tkns,
					got[i].tkns,
					func(a, b token) bool { return a.t == b.t && a.v == b.v },
				) {
					t.Errorf("function %d tokens = %v, want %v", i, got[i].tkns, tc.wantFuncs[i].tkns)
				}
			}
		})
	}
}
