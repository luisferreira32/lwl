package main

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

type instruction struct {
	opcode opset
	args   []string
}

func passemble(functions []function) []instruction {
	instructions := []instruction{}
	for _, f := range functions {
		// prologue
		name := f.name
		if f.main {
			name = "_start"
		}
		instructions = append(instructions,
			instruction{opcode: funcstart, args: []string{name}},
		)
		if !f.main {
			instructions = append(instructions,
				instruction{opcode: pushop, args: []string{rbp}},
				instruction{opcode: movop, args: []string{rbp, rsp}},
			)
		}
		// TODO: figure out MOD operator
		// TODO: figure out function calls
		// TODO: figure out most things...

		// body

		// TODO: this part must be done in the AST parsing, but this is a dummy test implementation to see something end-to-end
		// since this only supports adding constants, we know it will always be const + const + const + ...

		i := 0
		for {
			if i >= len(f.tkns) {
				break
			}
			t := f.tkns[i]
			// test: only know how to handle adding N numbers right now
			if t.t != tconstant && t.t != tadd {
				continue
			}

			// a + b // TODO: this should really be done at AST level
			if t.t == tconstant {
				instructions = append(instructions, instruction{opcode: movop, args: []string{t.v, rax}})
				i++
				if i >= len(f.tkns) { // TODO: should never happen, but need AST handling for that
					break
				}
				if f.tkns[i].t != tadd {
					continue
				}
				i++
				if i >= len(f.tkns) { // TODO: should never happen, but need AST handling for that
					break
				}
				t = f.tkns[i]
				if t.t != tconstant {
					continue
				}
				instructions = append(instructions,
					instruction{opcode: movop, args: []string{t.v, rbx}},
					instruction{opcode: addop, args: []string{rax, rbx}},
				)
			}

			// (had previously ADD in rax) + c
			if t.t == tadd {
				i++
				if i >= len(f.tkns) { // TODO: should never happen, but need AST handling for that
					break
				}
				t = f.tkns[i]
				if t.t != tconstant {
					continue
				}
				instructions = append(instructions,
					instruction{opcode: movop, args: []string{t.v, rbx}},
					instruction{opcode: addop, args: []string{rax, rbx}},
				)
			}

			i++
		}
		for _, t := range f.tkns {
			if t.t == tconstant {
				instructions = append(instructions,
					instruction{opcode: movop, args: []string{t.v, rax}},
				)
			}
			if t.t == tadd {
				instructions = append(instructions,
					instruction{opcode: popop, args: []string{rbx}},
					instruction{opcode: popop, args: []string{rax}},
					instruction{opcode: addop, args: []string{rax, rbx}},
					instruction{opcode: pushop, args: []string{rax}},
				)
			}
		}

		// epilogue
		if f.main {
			instructions = append(instructions,
				instruction{opcode: movop, args: []string{rax, rdi}},
				instruction{opcode: movop, args: []string{"60", rax}},
				instruction{opcode: syscallop, args: []string{}},
			)
		} else {
			instructions = append(instructions,
				instruction{opcode: popop, args: []string{rbp}},
				instruction{opcode: retop, args: []string{}},
			)

		}
	}
	return instructions
}
