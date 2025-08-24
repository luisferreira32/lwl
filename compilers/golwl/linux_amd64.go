package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// TODO: actual compiler optimizations, etc.
func toAs(i instruction) (string, error) {
	switch i.opcode {
	case syscallop:
		return "    SYSCALL", nil
	case addop:
		if len(i.args) != 2 {
			return "", fmt.Errorf("invalid number of args for ADD, expected 2, got: %v", i.args)
		}
		args1 := i.args[0]
		args2 := i.args[1]
		// TODO: handle add from const, memory, etc.
		if !isRegister(args1) || !isRegister(args2) {
			return "", fmt.Errorf("invalid args for ADD, expected registers, got: %v", i.args)
		}
		args1 = "%" + args1
		args2 = "%" + args2

		return fmt.Sprintf("    ADD %s, %s", args1, args2), nil
	case movop:
		// TODO: handle memory, etc.
		if len(i.args) != 2 {
			return "", fmt.Errorf("invalid number of args for MOV, expected 2, got: %v", i.args)
		}
		args1 := i.args[0]
		switch {
		case isConstant(args1):
			args1 = "$" + args1
		case isRegister(args1):
			args1 = "%" + args1
		}

		args2 := i.args[1]
		switch {
		case isConstant(args2):
			args2 = "$" + args2
		case isRegister(args2):
			args2 = "%" + args2
		}

		// TODO: make this assumption move obvious, but we do AT&T syntax src, dst
		return fmt.Sprintf("    MOV %s, %s", args1, args2), nil
	case funcstart:
		if len(i.args) != 1 {
			return "", fmt.Errorf("invalid number of args for %v, expected 1, got: %v", funcstart, i.args)
		}
		name := i.args[0]
		return fmt.Sprintf("%s:", name), nil
	default:
	}
	return "", errors.New("unhandled op") // TODO: actually handle other stuff and output meaningful errros
}

func magic(instructions []instruction, outfileName string) error {
	// transform instructions into GAS assembly
	// run it through the assembler then linker
	// write the output to the file
	asCode := strings.Builder{}
	asCode.WriteString(".section .text\n")
	asCode.WriteString(".global _start\n")
	for _, inst := range instructions {
		asLine, err := toAs(inst)
		if err != nil {
			return err
		}
		asCode.WriteString(asLine + "\n")
	}
	err := os.WriteFile(outfileName+".tmp.S", []byte(asCode.String()), 0o600)
	if err != nil {
		return err
	}

	// TODO: handle different platforms and check pre-conditions (like having as or ld installed)

	o, err := exec.Command("as", "-o", outfileName+".tmp.o", outfileName+".tmp.S").CombinedOutput()
	if err != nil {
		return fmt.Errorf("as failed: %v: %v", err, string(o))
	}
	o, err = exec.Command("ld", "-o", outfileName, outfileName+".tmp.o").CombinedOutput()
	if err != nil {
		return fmt.Errorf("as failed: %v: %v", err, string(o))
	}
	return nil
}
