package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime/debug"
	"strconv"
	"unicode/utf8"
)

const (
	// TODO: do not manually increment this
	version = "0.0.1"
	// TODO: do more than a wild guess on how it should be done
	readBufferSize = 1024
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
		return tokenValue, ""
	}
	// We currently assume that we only deal in integers, and the max
	// size for an integer will be 64 bits. Then we'll worry later the
	// actual size it was assigned in the lwl program.
	//
	// TODO: optimize this (check if it has non numeric runes?)
	_, err := strconv.ParseInt(name, 10, 64)
	if err != nil {
		// TODO: re-think it
		return LITERAL, name
	}
	return IDENTIFIER, name
}

// The Abstract Syntax Tree needs some nodes.
//
// goto:ast
type astNode struct {
	t token
	v string
	n astNodeType

	parent   *astNode
	children []*astNode
}

type astNodeType int

const (
	root astNodeType = iota + 1
	declaration
	statement
)

var (
	declarationBegin = map[token]struct{}{
		INTEGER:   {},
		INTEGER8:  {},
		INTEGER16: {},
		INTEGER32: {},
		INTEGER64: {},
		FUNCTION:  {},
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

	// 1. Tokenize

	var (
		readBuffer     = make([]byte, readBufferSize)
		offSet         int
		tokenName      string
		tokenList      []token
		tokenValueList []string
	)

	for {
		readBytes, err := fileToCompile.Read(readBuffer)
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			gracefullyHandleTheError(fmt.Errorf("read to buffer of size %d got: %w", readBufferSize, err))
		}

		offSet = 0
		for {
			if offSet >= readBytes {
				// TODO: there ought to be a more graceful way to handle this
				// or should we just enforce that there needs to ALWAYS be a newline at
				// then end of the file?
				if tokenName != "" {
					newToken, newLiteral := tokenize(tokenName)
					tokenList = append(tokenList, newToken)
					tokenValueList = append(tokenValueList, newLiteral)
				}
				tokenName = ""
				break
			}
			currentRune, offSetIncrement := utf8.DecodeRune(readBuffer[offSet:])
			if currentRune == utf8.RuneError {
				gracefullyHandleTheError(fmt.Errorf("decoding a rune from buffer %s got: RuneError", readBuffer))
			}
			offSet += offSetIncrement

			if currentRune == '\n' || currentRune == ' ' || currentRune == '	' {
				if tokenName != "" {
					newToken, newLiteral := tokenize(tokenName)
					tokenList = append(tokenList, newToken)
					tokenValueList = append(tokenValueList, newLiteral)
				}
				tokenName = ""
				continue
			}

			tokenName += string(currentRune)
		}
	}

	// 2. AST
	astRoot := astNode{
		v: "start of the program!",
		n: root,
	}
	astCurrentNode := &astRoot

	for i, t := range tokenList {
		_, ok := declarationBegin[t]
		if ok {
			astNewNode := &astNode{
				t:      t,
				v:      tokenValueList[i],
				n:      declaration,
				parent: astCurrentNode,
			}
			astCurrentNode.children = append(astCurrentNode.children, astNewNode)
			astCurrentNode = astNewNode
			continue
		}

		astNewNode := &astNode{
			t:      t,
			v:      tokenValueList[i],
			n:      statement,
			parent: astCurrentNode,
		}
		astCurrentNode.children = append(astCurrentNode.children, astNewNode)

		// TODO: actually do more than declaration
		//
		// gracefullyHandleTheError(fmt.Errorf("unknown token %v, %s, %s", t, t, tokenValueList[i]))
	}

	// 3. ...

	// 4. Profit

	fmt.Printf("We are working on compiling this... currently we got the list of tokens:\n")
	for i := 0; i < len(tokenList); i++ {
		fmt.Printf("token, name: %d - %s, %s\n", tokenList[i], tokenList[i], tokenValueList[i])
	}
	fmt.Printf("And this AST:\n")
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
