package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime/debug"
	"strconv"
	"strings"
	"sync"
)

const (
	// TODO: do not manually increment this
	version = "0.0.1"
)

// The magic tokens for our lexicon!
//
// goto:token
type token struct {
	v tokenType
	n string

	// metadata
}

type tokenType int

const (
	INVALID tokenType = iota

	IDENTIFIER
	LITERAL

	// We like integers, everyone does. But we don't like not to know
	// how many bits it will occupy. So we accept an integer, and assign
	// it 32 bits per default, but we will complain.
	//
	// TODO: make the compiler complain if the integer does not specify bits
	INTEGER
	INTEGER8
	INTEGER16
	INTEGER32
	INTEGER64
	// TODO: add something similar to structs

	ADD
	SUBTRACT
	MULTIPLY
	QUOCIENT
	REMAIN

	AND
	OR
	XOR
	SHIFTLEFT
	SHIFTRIGHT

	ASSIGN

	IF
	WHILE
	RETURN

	CURLYLEFT
	CURLYRIGHT
	PARENTHESISLEFT
	PARENTHESISRIGHT
	// TODO: we need brackets for arrays
)

var (
	tokenToString = map[tokenType]string{
		IDENTIFIER:       "identifier",
		LITERAL:          "literal",
		INTEGER:          "integer",
		INTEGER8:         "integer8",
		INTEGER16:        "integer16",
		INTEGER32:        "integer32",
		INTEGER64:        "integer64",
		ADD:              "+",
		SUBTRACT:         "-",
		MULTIPLY:         "*",
		QUOCIENT:         "/",
		REMAIN:           "%",
		AND:              "&",
		OR:               "|",
		XOR:              "^",
		SHIFTLEFT:        "<<",
		SHIFTRIGHT:       ">>",
		ASSIGN:           "=",
		IF:               "if",
		WHILE:            "while",
		RETURN:           "return",
		CURLYLEFT:        "{",
		CURLYRIGHT:       "}",
		PARENTHESISLEFT:  "(",
		PARENTHESISRIGHT: ")",
	}
	stringToToken = map[string]tokenType{
		"integer":   INTEGER,
		"integer8":  INTEGER8,
		"integer16": INTEGER16,
		"integer32": INTEGER32,
		"integer64": INTEGER64,
		"+":         ADD,
		"-":         SUBTRACT,
		"*":         MULTIPLY,
		"/":         QUOCIENT,
		"%":         REMAIN,
		"&":         AND,
		"|":         OR,
		"^":         XOR,
		"<<":        SHIFTLEFT,
		">>":        SHIFTRIGHT,
		"=":         ASSIGN,
		"if":        IF,
		"while":     WHILE,
		"return":    RETURN,
		"{":         CURLYLEFT,
		"}":         CURLYRIGHT,
		"(":         PARENTHESISLEFT,
		")":         PARENTHESISRIGHT,
	}
)

func (t tokenType) String() string {
	tokenStr, ok := tokenToString[t]
	if !ok {
		return fmt.Sprintf("ops, did not find string for token: %d", t)
	}
	return tokenStr
}

func tokenize(name string) token {
	tokenValue, ok := stringToToken[name]
	if ok {
		return token{v: tokenValue, n: name}
	}
	// We currently assume that we only deal in integers, and the max
	// size for an integer will be 64 bits. Then we'll worry later the
	// actual size it was assigned in the lwl program.
	//
	// TODO: optimize this (check if it has non numeric runes?)
	_, err := strconv.ParseInt(name, 10, 64)
	if err != nil {
		// TODO: re-think it
		return token{v: IDENTIFIER, n: name}
	}
	return token{v: LITERAL, n: name}
}

// The Abstract Syntax Tree needs some nodes.
//
// goto:ast
type astNode struct {
	// associated tokens with this astNode
	t []token
	n astNodeType

	parent   *astNode
	children []*astNode

	// metadata
	line int
}

type astNodeType int

const (
	root astNodeType = iota + 1
	declaration
	statement
	enclosedStatement
	returnStatement
)

var (
	declarationBegin = map[tokenType]bool{
		INTEGER:   true,
		INTEGER8:  true,
		INTEGER16: true,
		INTEGER32: true,
		INTEGER64: true,
	}
	statementBegin = map[tokenType]bool{
		IDENTIFIER: true,
	}
	enclosedStatementBegin = map[tokenType]bool{
		IF:    true,
		WHILE: true,
	}
	bitOperator = map[tokenType]bool{
		AND:        true,
		OR:         true,
		XOR:        true,
		SHIFTLEFT:  true,
		SHIFTRIGHT: true,
	}
	arithmeticOperator = map[tokenType]bool{
		ADD:      true,
		SUBTRACT: true,
		MULTIPLY: true,
		QUOCIENT: true,
		REMAIN:   true,
	}
	literalOrIdentifier = map[tokenType]bool{
		IDENTIFIER: true,
		LITERAL:    true,
	}
)

func buildAST(fileNameToCompile string, scannerBuffer *bufio.Scanner) *astNode {
	var (
		lineNumber = 0

		astRoot = astNode{
			n: root,
		}
		astCurrentNode = &astRoot

		nextExpectedToken tokenType
	)
	astCurrentNode.parent = nil

	for scannerBuffer.Scan() {
		lineNumber++

		line := scannerBuffer.Text()
		words := strings.Fields(line)
		if len(words) == 0 {
			continue
		}

		lineTokens := make([]token, len(words))
		for i, word := range words {
			// tokenize, fill metadata, and store
			lineTokens[i] = tokenize(word)
		}

		astNewNode := &astNode{
			parent: astCurrentNode,
		}
		firstToken := lineTokens[0] // first token decides the line type

		switch {
		case firstToken.v == CURLYLEFT:
			// Parse sub-tree start.
			//
			// If we expect a CURLYLEFT token, it means the previous line was an enclosed statement or a function
			// declaration. Just reset the next expected token and continue the parsing.
			if nextExpectedToken != CURLYLEFT {
				addCompileErr(fmt.Sprintf("[%s:%d] unexpected token %s\n", fileNameToCompile, lineNumber, firstToken.n))
				continue
			}
			nextExpectedToken = INVALID

		case firstToken.v == CURLYRIGHT:
			// Parse sub-tree finish.
			//
			// It is only possible to stop a sub-tree if we're inside a sub-tree, so check if the parent is present,
			// i.e., we are not on root node, and point the current node to the parent.
			if astCurrentNode.parent == nil {
				addCompileErr(fmt.Sprintf("[%s:%d] unexpected token %s\n", fileNameToCompile, lineNumber, firstToken.n))
				continue
			}
			if nextExpectedToken != CURLYRIGHT && astCurrentNode.n != enclosedStatement {
				addCompileErr(fmt.Sprintf("[%s:%d] unexpected token %s without a function/if/while\n", fileNameToCompile, lineNumber, firstToken.n))
				continue
			}
			astCurrentNode = astCurrentNode.parent

		case declarationBegin[firstToken.v]:
			// Parse declarations.
			//
			// We expect a declaration to be either a variable declaration or a function declaration.
			// For that we need a type (variable or return) and an identifier. Given those, if it is
			// a function declaration we start a sub-tree with the arguments as declarations themselves
			// inside the sub-tree. To start the sub-tree we just pass the current node pointer to the
			// function declaration pointer and expect a curly left brace.
			astNewNode.n = declaration
			astCurrentNode.children = append(astCurrentNode.children, astNewNode)

			if len(lineTokens) < 2 || lineTokens[1].v != IDENTIFIER {
				addCompileErr(fmt.Sprintf("[%s:%d] need a name for a declaration\n", fileNameToCompile, lineNumber))
				continue
			}

			if len(lineTokens) < 3 {
				astNewNode.t = lineTokens
				continue
			}

			astCurrentNode = astNewNode
			nextExpectedToken = CURLYLEFT

			if lineTokens[2].v != PARENTHESISLEFT && lineTokens[len(lineTokens)-1].v != PARENTHESISRIGHT {
				addCompileErr(fmt.Sprintf("[%s:%d] function %v declaration must have parameters enclosed in parenthesis\n", fileNameToCompile, lineNumber, lineTokens[1].n))
				continue
			}

			for i := 0; i < (len(lineTokens)-4)/2; i++ {
				variableTypeIndex := i*2 + 3
				variableNameIndex := i*2 + 4
				if variableNameIndex > len(lineTokens)-2 {
					addCompileErr(fmt.Sprintf("[%s:%d] function %v declaration arguments have to be declared with _type_ and _name_\n", fileNameToCompile, lineNumber, lineTokens[1].n))
					break
				}
				astCurrentNode.children = append(astCurrentNode.children, &astNode{
					t:      []token{lineTokens[variableTypeIndex], lineTokens[variableNameIndex]},
					n:      declaration,
					parent: astCurrentNode,
				})
			}

		case statementBegin[firstToken.v]:
			// Parse a statement.
			//
			// When we reach a statement we need to register which kind of operations this statement is going
			// to include. The syntax check is then later done on the tree to ensure all the operators of the
			// statement were declared and are valid to use together. Expect a statement to be either an
			// arithmetic operation, a bit operation, or a function invocation.
			astNewNode.n = statement
			astCurrentNode.children = append(astCurrentNode.children, astNewNode)

			if len(lineTokens) < 3 || lineTokens[1].v != ASSIGN || !literalOrIdentifier[lineTokens[2].v] {
				addCompileErr(fmt.Sprintf("[%s:%d] statement should begin with: %s = ...\n", fileNameToCompile, lineNumber, firstToken.n))
				continue
			}
			astNewNode.t = lineTokens

			if len(lineTokens) < 5 {
				continue
			}

			switch {
			case lineTokens[3].v == PARENTHESISLEFT:
				if lineTokens[len(lineTokens)-1].v != PARENTHESISRIGHT {
					addCompileErr(fmt.Sprintf("[%s:%d] function %s invocation expected enclosing parenthesis\n", fileNameToCompile, lineNumber, lineTokens[2].n))
					continue
				}
				for _, token := range lineTokens[4:] {
					if !literalOrIdentifier[token.v] {
						addCompileErr(fmt.Sprintf("[%s:%d] function %s invocation expects constant or variable arguments, got %s\n", fileNameToCompile, lineNumber, lineTokens[2].n, token.n))
						break
					}
				}
			case arithmeticOperator[lineTokens[3].v]:
				for _, token := range lineTokens[4:] {
					if !literalOrIdentifier[token.v] && !arithmeticOperator[token.v] {
						addCompileErr(fmt.Sprintf("[%s:%d] arithmetic operation expects *only* arithmetic operators and constants/variables\n", fileNameToCompile, lineNumber))
						break
					}
				}
			case bitOperator[lineTokens[3].v]:
				for _, token := range lineTokens[4:] {
					if !literalOrIdentifier[token.v] && !bitOperator[token.v] {
						addCompileErr(fmt.Sprintf("[%s:%d] bit operation statement expects *only* bit operators and constants/variables\n", fileNameToCompile, lineNumber))
						break
					}
				}
			}

		case enclosedStatementBegin[firstToken.v]:
			// Parse an enclosed statement.
			//
			// The enclosed statement will be done based on a condition, so we expect an identifier to be that condition,
			// i.e., if the given identifier value at runtime is greater than zero, the sub-tree is processed, otherwise, it is not.
			astNewNode.n = enclosedStatement
			astCurrentNode.children = append(astCurrentNode.children, astNewNode)

			if len(lineTokens) != 4 || (lineTokens[1].v != PARENTHESISLEFT && lineTokens[2].v != IDENTIFIER && lineTokens[3].v != PARENTHESISRIGHT) {
				addCompileErr(fmt.Sprintf("[%s:%d] statement %s expects a variable within braces, i.e., %s ( identifier )\n", fileNameToCompile, lineNumber, firstToken.n, firstToken.n))
				continue
			}

			astCurrentNode = astNewNode
			nextExpectedToken = CURLYLEFT
			// TODO: ensure the identifier is not a function identifier

		case firstToken.v == RETURN:
			// Parse the return statement.
			//
			// If we return it should be expected that the next token is the closing of curly braces.
			// But we leave it up to the syntax check to prune the following children from the tree
			// after the return.
			astNewNode.n = returnStatement
			astNewNode.t = lineTokens

			astCurrentNode.children = append(astCurrentNode.children, astNewNode)
			nextExpectedToken = CURLYRIGHT

		default:
			addCompileErr(fmt.Sprintf("[%s:%d] unexpected token: %s, %s\n > %v\n", fileNameToCompile, lineNumber, firstToken.v, firstToken.n, words))
		}
	}

	if scannerBuffer.Err() != nil {
		gracefullyHandleTheError(fmt.Errorf("read to file %s  got: %w", fileNameToCompile, scannerBuffer.Err()))
	}

	return &astRoot
}

// The best way to handle errors, ours, or theirs.
//
// goto:error_handling
var (
	errMisusage       = fmt.Errorf("Misusage")
	errGrossMissusage = fmt.Errorf("Gross misusage")
	errOnCompile      = fmt.Errorf("Compiler error")

	compileErrDB     = []string{}
	compileErrDBLock = sync.Mutex{}
)

func gracefullyHandleTheError(err error) {
	panic(err)
}

func issuePrint(info interface{}, stack []byte) {
	fmt.Printf(`There seems to be a technical problem, you can open a github 
issue at github.com/luisferreira32/lwl with the following output:

version: %s
err: %s
stack:
%s
`, version, info, stack)
}

func catchTheGracefulHandler() {
	if recovered := recover(); recovered != nil {
		issuePrint(recovered, debug.Stack())
		os.Exit(1)
	}
}

func addCompileErr(compileErr ...string) {
	compileErrDBLock.Lock()
	defer compileErrDBLock.Unlock()
	compileErrDB = append(compileErrDB, compileErr...)
}

func getCompileErrs() string {
	compileErrDBLock.Lock()
	defer compileErrDBLock.Unlock()
	r := ""
	for _, v := range compileErrDB {
		r += v
	}
	return r
}

// Syntax checking comes from the AST.
//
// goto:syntax
func checkChildSyntax(fileName string, n *astNode, declaredVariables, declaredFunctions map[string]bool) {
	if n == nil {
		return
	}

	switch {
	case n.n == declaration:
		if declaredVariables[n.t[1].n] || declaredFunctions[n.t[1].n] {
			addCompileErr(fmt.Sprintf("[%s:%d] re-declaration of %s\n", fileName, n.line, n.t[1].n))
		}
		// TODO other declaration checks
	case n.n == statement:
		// TODO statement checks
	case n.n == enclosedStatement:
		// TODO enclosedStatement checks
	case n.n == returnStatement:
		// TODO returnStatement checks
	case n.n == root:
		// root just goes to children
	}

	for _, child := range n.children {
		if n.n == declaration { // function declaration changes variable scope
			checkChildSyntax(fileName, child, make(map[string]bool), make(map[string]bool))
		} else {
			checkChildSyntax(fileName, child, declaredVariables, declaredFunctions)
		}
	}
}

func parseAST(fileName string, astRoot *astNode) {
	astCurrentNode := astRoot
	// TODO: sort this out in terms of scope
	var (
		declaredVariables map[string]bool
		declaredFunctions map[string]bool
	)
	checkChildSyntax(fileName, astCurrentNode, declaredVariables, declaredFunctions)
}

// Our glorious compiler! It will live as long as lwl does not know how to
// compile itself.
//
// goto:compile
func compile() error {
	defer catchTheGracefulHandler()

	if len(os.Args) < 2 || os.Args[1] == "--help" || os.Args[1] == "-h" {
		return errMisusage
	}

	// TODO: actually have some flags for helping out and then do this check
	if len(os.Args) > 2 {
		return fmt.Errorf("%w: We only allow ONE file, %v given!", errGrossMissusage, len(os.Args)-1) // nolint // it's to display to the user
	}

	fileNameToCompile := filepath.Clean(os.Args[1])
	fileToCompile, err := os.Open(fileNameToCompile)
	if err != nil {
		gracefullyHandleTheError(fmt.Errorf("on initial file open %s got: %w", fileNameToCompile, err))
	}
	defer func() {
		_ = fileToCompile.Close()
	}()

	// Tokenize + build AST
	astRoot := buildAST(fileNameToCompile, bufio.NewScanner(fileToCompile))

	// Parse syntax
	parseAST(fileNameToCompile, astRoot)

	// ???

	// Profit

	if compileErrs := getCompileErrs(); compileErrs != "" {
		return fmt.Errorf("%w:\n%s\n", errOnCompile, compileErrs) // nolint // it's to display to the user
	}

	fmt.Printf("We are working on compiling this...\n")
	fmt.Printf("at least we got this AST:\n\n")
	walk(astRoot, "")
	return nil
}

// Usage is always a nice thing to offer to the end user.
// But we're not that nice. So no hints are really given at the moment.
//
// goto:usage
func usage(binName string, angry bool, angryReasons ...string) {
	var greeting string
	if angry {
		greeting = "You have failed the lwl compiler! How could you?\n"
		for _, reason := range angryReasons {
			greeting += " * " + reason + "\n"
		}
		greeting += "Might as well explain it to you."
	} else {
		greeting = "Welcome to the lwl compiler!"
	}
	fmt.Printf(`%s The usage is pretty simple:
	%s file.lwl

It will write the executable as a "file.magic". You can also define 
a couple flags, but I won't spoil the fun, read the source code!
`, greeting, binName)
}

// Our glorious main! It will live as long as lwl does not know how to
// compile itself.
//
// goto:main
func main() {
	err := compile()
	switch {
	case err == nil:
		os.Exit(0)
	case errors.Is(err, errMisusage):
		usage(os.Args[0], false)
		os.Exit(1)
	case errors.Is(err, errGrossMissusage):
		usage(os.Args[0], true, err.Error())
		os.Exit(1)
	case errors.Is(err, errOnCompile):
		fmt.Printf("%v", err.Error())
		os.Exit(1)
	default:
		issuePrint(err.Error(), []byte("unexpected"))
		os.Exit(1)
	}
}

// TODO: delete this
func walk(n *astNode, indent string) {
	if n == nil {
		return
	}
	fmt.Printf("%vnode: %d, %v\n", indent, n.n, n.t)
	for _, v := range n.children {
		walk(v, indent+" ")
	}
}
