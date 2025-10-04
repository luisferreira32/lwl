package main

import (
	"strconv"
)

// This is a pseudo-assembler for the LWL language.
// It will create a generic pseudo-assembly code that can be later be thrown in different architectures.

type opset string

const (
	funcstart opset = "FUNC_START"
	retop     opset = "RET"
	movop     opset = "MOV"
	addop     opset = "ADD"
	subop     opset = "SUB"
	mulop     opset = "MUL"
	divop     opset = "DIV"
	pushop    opset = "PUSH"
	popop     opset = "POP"
	callop    opset = "CALL"
	syscallop opset = "SYSCALL"
)

const (
	rax = "RAX"
	rbx = "RBX"
	rcx = "RCX"
	rdx = "RDX"
	rsi = "RSI"
	rdi = "RDI"
	rbp = "RBP"
	rsp = "RSP"
)

func isRegister(s string) bool {
	switch s {
	case rax, rbx, rcx, rdx, rsi, rdi, rbp, rsp:
		return true
	}
	return false
}

func isConstant(s string) bool {
	_, err := strconv.Atoi(s)
	return err == nil
}

type instruction struct {
	opcode opset
	args   []string
}

func passemble(ops map[fname]*operation, verboseLevel int) []instruction {
	instructions := []instruction{}
	for name := range ops {
		// prologue
		instructions = append(instructions,
			instruction{opcode: funcstart, args: []string{string(name)}},
			instruction{opcode: pushop, args: []string{rbp}},
			instruction{opcode: movop, args: []string{rbp, rsp}},
		)

		// TODO: figure out MOD operator
		// TODO: figure out function calls
		// TODO: figure out most things...

		// body

		// epilogue
		instructions = append(instructions,
			instruction{opcode: popop, args: []string{rbp}},
			instruction{opcode: retop, args: []string{}},
		)
	}

	// now the hacks: call main from _start and ensure ret value goes to exit code
	instructions = append(instructions,
		instruction{opcode: funcstart, args: []string{"_start"}},
		instruction{opcode: movop, args: []string{rax, rdi}},
		instruction{opcode: movop, args: []string{"60", rax}},
		instruction{opcode: syscallop, args: []string{}},
	)

	if verboseLevel >= verboseDEBUG {

	}
	return instructions
}
