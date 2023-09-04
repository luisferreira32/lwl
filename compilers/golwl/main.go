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
	lineNumber int
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

	LOGICALAND
	LOGICALOR
	EQUALS
	LESS
	GREAT
	NOT
	// Yes. You can do less and equal, but it's just not given to you by the
	// compiler. Off-by-one is my motto.
	//
	// TODO: make it easier and add LEQ and GEQ operators

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
		LOGICALAND:       "&&",
		LOGICALOR:        "||",
		EQUALS:           "==",
		LESS:             "<",
		GREAT:            ">",
		NOT:              "!",
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
		"&&":        LOGICALAND,
		"||":        LOGICALOR,
		"==":        EQUALS,
		"<":         LESS,
		">":         GREAT,
		"!":         NOT,
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
}

type astNodeType int

const (
	root astNodeType = iota + 1
	declaration
	statement
	enclosedStatement
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
)

// Usage is always a nice thing to offer to the end user.
// But we're not that nice. So no hints are really given at the moment.
//
// goto:usage
func usage(binName string, angry bool, angryReasons ...string) {
	greeting := ""
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

	// 1. Tokenize in memory

	var (
		scannerBuffer = bufio.NewScanner(fileToCompile)
		lineCounter   = 0
		// TODO: optimize based on filesize heuristics
		//
		// At the moment seemed like a reasonable number for pre-emptive token space allocation.
		tokenizedFile = make([]token, 0, 1024)
	)

	for scannerBuffer.Scan() {
		lineCounter++

		line := scannerBuffer.Text()
		words := strings.Fields(line)
		if len(words) == 0 {
			continue
		}

		for _, word := range words {
			// tokenize, fill metadata, store
			tokenizedWord := tokenize(word)
			tokenizedWord.lineNumber = lineCounter
			tokenizedFile = append(tokenizedFile, tokenizedWord)
		}
	}

	if scannerBuffer.Err() != nil {
		gracefullyHandleTheError(fmt.Errorf("read to file %s  got: %w", fileNameToCompile, err))
	}

	// 2. Build the AST

	var (
		astRoot = astNode{
			n: root,
		}
		astCurrentNode = &astRoot

		nextExpectedToken  tokenType
		parenthesisCounter = 0
	)
	astCurrentNode.parent = nil

	for _, newToken := range tokenizedFile {
		astNewNode := &astNode{
			parent: astCurrentNode,
		}

		if nextExpectedToken != INVALID && nextExpectedToken != newToken.v {
			addCompileErr(fmt.Sprintf("[%s:%d] unexpected token, expected '%s', got '%s'\n", fileNameToCompile, newToken.lineNumber, nextExpectedToken, newToken.v))
		}

		switch {
		case declarationBegin[newToken.v] && astCurrentNode.n != declaration:
			// If it's a declaration start a new node for it child to the current node
			// and expect an identifier next
			astNewNode.n = declaration
			astNewNode.t = append(astNewNode.t, newToken)
			astCurrentNode.children = append(astCurrentNode.children, astNewNode)

			astCurrentNode = astNewNode
			nextExpectedToken = IDENTIFIER

		case newToken.v == nextExpectedToken && astCurrentNode.n == declaration:
			// If we were expecting an identifier to the declaration, just add the token
			// to the relevant node token list.
			astCurrentNode.t = append(astCurrentNode.t, newToken)

		case statementBegin[newToken.v] && astCurrentNode.n != declaration:
			// Any statement can be signaled with an identifier, if it has been declared before
			// and we're not declaring it. It will be comprised of a bunch of tokens
			astNewNode.n = statement
			astNewNode.t = append(astNewNode.t, newToken)
			astCurrentNode.children = append(astCurrentNode.children, astNewNode)

			astCurrentNode = astNewNode
			// TODO: ensure the identifier has been declared

		case enclosedStatementBegin[newToken.v]:
			// If we start an enclosed statement we expect two things:
			// 1. A parenthesis to begin the enclosed statement condition and to end
			// 2. Curly braces to begin the context of the enclosed statement and to end
			astNewNode.n = enclosedStatement
			astNewNode.t = append(astNewNode.t, newToken)
			astCurrentNode.children = append(astCurrentNode.children, astNewNode)

			astCurrentNode = astNewNode
			nextExpectedToken = PARENTHESISLEFT

		case newToken.v == nextExpectedToken && astCurrentNode.n == enclosedStatement:
			// An enclosed statement openned curly braces properly
			astCurrentNode.t = append(astCurrentNode.t, newToken)
			parenthesisCounter++

		case newToken.v == PARENTHESISRIGHT && astCurrentNode.n == statement && astCurrentNode.parent.n == enclosedStatement:
			// An enclosed statement condition closed curly braces properly
			astCurrentNode = astCurrentNode.parent
			astCurrentNode.t = append(astCurrentNode.t, newToken)
			parenthesisCounter--
			// TODO: verify the condition is boolean

		case newToken.v == PARENTHESISLEFT && astCurrentNode.n == declaration:
			// This is actually a *FUNCTION* declaration
			// TODO: handle it -> separate the context into its own branch for variable declarations

		default:
			addCompileErr(fmt.Sprintf("[%s:%d] unexpected token: %s, %s\n", fileNameToCompile, newToken.lineNumber, newToken.v, newToken.n))
		}
		// TODO: account for functions
	}
	if parenthesisCounter != 0 {
		// TODO: indicate lines
		addCompileErr(fmt.Sprintf("[%s:xxx] unexpected open/closed (+/-) parenthesis: %v\n", fileNameToCompile, parenthesisCounter))
	}

	// 3. ???

	// 4. Profit

	if compileErrs := getCompileErrs(); compileErrs != "" {
		return fmt.Errorf("%w:\n%s\n", errOnCompile, compileErrs) // nolint // it's to display to the user
	}

	fmt.Printf("We are working on compiling this...\n")
	fmt.Printf("at least we got this AST:\n\n")
	walk(&astRoot, "")
	return nil
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
