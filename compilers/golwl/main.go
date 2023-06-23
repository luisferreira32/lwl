package main

import (
	"bufio"
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
type token int

const (
	INVALID token = iota

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
	FUNCTION
	RETURN

	CURLYLEFT
	CURLYRIGHT
	PARENTHESISLEFT
	PARENTHESISRIGHT
	// TODO: we need brackets for arrays
)

var (
	tokenToString = map[token]string{
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
		FUNCTION:         "function",
		RETURN:           "return",
		CURLYLEFT:        "{",
		CURLYRIGHT:       "}",
		PARENTHESISLEFT:  "(",
		PARENTHESISRIGHT: ")",
	}
	stringToToken = map[string]token{
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
		"function":  FUNCTION,
		"return":    RETURN,
		"{":         CURLYLEFT,
		"}":         CURLYRIGHT,
		"(":         PARENTHESISLEFT,
		")":         PARENTHESISRIGHT,
	}
)

func (t token) String() string {
	tokenStr, ok := tokenToString[t]
	if !ok {
		return fmt.Sprintf("ops, did not find string for token: %d", t)
	}
	return tokenStr
}

func tokenize(name string) (token, string) {
	tokenValue, ok := stringToToken[name]
	if ok {
		return tokenValue, name
	}
	// We currently assume that we only deal in integers, and the max
	// size for an integer will be 64 bits. Then we'll worry later the
	// actual size it was assigned in the lwl program.
	//
	// TODO: optimize this (check if it has non numeric runes?)
	_, err := strconv.ParseInt(name, 10, 64)
	if err != nil {
		// TODO: re-think it
		return IDENTIFIER, name
	}
	return LITERAL, name
}

// The Abstract Syntax Tree needs some nodes.
//
// goto:ast
type astNode struct {
	t token
	v string
	n astNodeType

	lineNumber int

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
	declarationBegin = map[token]bool{
		INTEGER:   true,
		INTEGER8:  true,
		INTEGER16: true,
		INTEGER32: true,
		INTEGER64: true,
	}
	statementBegin = map[token]bool{
		IDENTIFIER: true,
	}
	statementEnd           = map[token]bool{}
	enclosedStatementBegin = map[token]bool{
		IF:    true,
		WHILE: true,
	}
	enclosedStatementEnd = map[token]bool{}
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
	compileErrDB     = []string{}
	compileErrDBLock = sync.Mutex{}
)

func gracefullyHandleTheError(err error) {
	panic(err)
}
func catchTheGracefulHandler() {
	if recovered := recover(); recovered != nil {
		fmt.Printf(`There seems to be a technical problem, you can open a github 
issue at github.com/luisferreira32/lwl with the following output:

version: %s
err: %s
stack:
%s
`, version, recovered, debug.Stack())
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

// Our glorious main! It will live as long as lwl does not know how to
// compile itself.
//
// goto:main
func main() {
	defer catchTheGracefulHandler()

	if len(os.Args) < 2 || os.Args[1] == "--help" || os.Args[1] == "-h" {
		usage(os.Args[0], false)
		return
	}

	// TODO: actually have some flags for helping out and then do this check
	if len(os.Args) > 2 {
		usage(os.Args[0], true, "We only allow ONE file!")
		return
	}

	fileNameToCompile := filepath.Clean(os.Args[1])
	fileToCompile, err := os.Open(fileNameToCompile)
	if err != nil {
		gracefullyHandleTheError(fmt.Errorf("on initial file open %s got: %w", fileNameToCompile, err))
	}
	defer func() {
		_ = fileToCompile.Close()
	}()

	// 1. Tokenize & build the AST

	var (
		scannerBuffer = bufio.NewScanner(fileToCompile)
		lineCounter   = 0

		astRoot = astNode{
			v: "start of the program!",
			n: root,
		}
		astCurrentNode = &astRoot
	)
	astCurrentNode.parent = &astRoot

	for scannerBuffer.Scan() {
		lineCounter++

		line := scannerBuffer.Text()
		words := strings.Fields(line)
		if len(words) == 0 {
			continue
		}

		tokenLine := make(map[string]token, len(words))

		for _, word := range words {
			newToken, tokenName := tokenize(word)
			tokenLine[tokenName] = newToken
		}

		for tokenName, newToken := range tokenLine {
			astNewNode := &astNode{
				t:          newToken,
				v:          tokenName,
				lineNumber: lineCounter,
				parent:     astCurrentNode,
			}
			astCurrentNode.children = append(astCurrentNode.children, astNewNode)

			// TODO: this is just wrong, we can really only evaluate how the node and children will be at the end of the line
			switch {
			case declarationBegin[newToken] && astCurrentNode.n != declaration:
				astNewNode.n = declaration
				astCurrentNode = astNewNode
			case newToken == IDENTIFIER && astCurrentNode.n == declaration:
				astNewNode.n = declaration
			case statementBegin[newToken]:
				astNewNode.n = statement
			default:
				// TODO: be nicer in explaining why this should not be here
				//
				// addCompileErr(fmt.Sprintf("[%s:%d] unexpected token: %s, %s\n", fileNameToCompile, lineCounter, newToken, tokenName))
			}
		}

		// NOTE: last operation knowing we finished reading a line of tokens
		// so feel free to throw any errors at the user
		switch {
		case astCurrentNode.n == declaration:
			astCurrentNode = astCurrentNode.parent
		case astCurrentNode.n == enclosedStatement:
			addCompileErr(fmt.Sprintf("[%s:%d] unexpected linebreak, did you forget a ')'\n", fileNameToCompile, lineCounter))
		default:
		}
	}
	if scannerBuffer.Err() != nil {
		gracefullyHandleTheError(fmt.Errorf("read to file %s  got: %w", fileNameToCompile, err))
	}

	// 2. ???

	// 3. Profit

	if compileErrs := getCompileErrs(); compileErrs != "" {
		fmt.Printf("Got compilation error:\n%s\n", compileErrs)
		return
	}

	fmt.Printf("We are working on compiling this...\n")
	fmt.Printf("at least we got this AST:\n\n")
	walk(&astRoot, "")
}

// TODO: delete this
func walk(n *astNode, indent string) {
	if n == nil {
		return
	}
	fmt.Printf("%vnode: %d, %s, %s\n", indent, n.n, n.v, n.t)
	for _, v := range n.children {
		walk(v, indent+" ")
	}
}
