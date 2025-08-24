package main

import (
	"errors"
	"fmt"
	"log"
)

var (
	errParse         = errors.New("parse error")
	errNoMain        = errors.New("no main function defined")
	errMultipleMains = errors.New("multiple main functions defined")
)

// NOTE: for now parse only checks syntax parsing
// there's no abstract syntax tree or semantic analysis due to the simplicity of the language
// TODO: accept parenthesis syntax in expressions for grouping order
func parse(functions []function) error {
	functionRegistry := make(map[string]struct{})
	mainFunctions := make([]function, 0, 1)
	for i := range functions {
		// TODO: make this possible to run in parallel and safer than this
		f := &functions[i] // get the pointer to be able to append to errs
		if _, exists := functionRegistry[f.name]; exists && !f.main {
			f.errs = append(f.errs, errors.New("function "+f.name+" already defined"))
			continue
		}
		functionRegistry[f.name] = struct{}{}

		// check only one or zero eq are defined
		eqCount := 0
		for _, t := range f.tkns {
			if t.t == teq {
				eqCount++
			}
		}
		if eqCount > 1 {
			f.errs = append(f.errs, errors.New("function "+f.name+" has multiple '='"))
			continue
		}
		if eqCount == 0 && !f.main { // should not be possible due to tokenizer logic
			f.errs = append(f.errs, errors.New("function "+f.name+" has no '='"))
			continue
		}
		if f.main {
			mainFunctions = append(mainFunctions, *f)
		}

		if f.main && f.tkns[0].t != tvariable && f.tkns[0].t != tconstant {
			f.errs = append(f.errs, errors.New("main function must start with a variable or constant"))
			continue
		}

		// for each token, check if it is a valid declared variable on the function or an existing function call
		declaredVariables := make(map[string]struct{})
		functionHeaderEnded := false
		i := 0
		for {
			prevToken := f.tkns[i]
			i++
			if i >= len(f.tkns) {
				break
			}
			t := f.tkns[i]

			// function declaration is going to b X(vars)=, unless it is the main function
			if t.t == tvariable && !f.main && !functionHeaderEnded {
				declaredVariables[t.v] = struct{}{}
				continue
			}
			if t.t == teq {
				functionHeaderEnded = true
				continue
			}
			if t.t == tlparenth && prevToken.t != tvariable {
				f.errs = append(f.errs, errors.New("unexpected '(' after "+prevToken.v))
				continue
			}
			if t.t == trparenth && prevToken.t != tconstant && prevToken.t != tvariable {
				f.errs = append(f.errs, errors.New("unexpected ')' after "+prevToken.v))
				continue
			}
			if t.t == tcomma && prevToken.t != tconstant && prevToken.t != tvariable {
				f.errs = append(f.errs, errors.New("unexpected ',' after "+prevToken.v))
				continue
			}
			if t.t == tvariable {
				_, isDeclared := declaredVariables[t.v]
				_, isFunction := functionRegistry[t.v]
				if !isDeclared && !isFunction {
					f.errs = append(f.errs, errors.New("undefined variable "+t.v))
					continue
				}
			}

			if t.isOp() && prevToken.t != tconstant && prevToken.t != tvariable && prevToken.t != trparenth {
				f.errs = append(f.errs, errors.New("unexpected operator after "+prevToken.v))
				continue
			}

			if t.t == tconstant && prevToken.t != tcomma && !prevToken.isOp() && prevToken.t != tlparenth && prevToken.t != teq {
				f.errs = append(f.errs, errors.New("unexpected constant after "+prevToken.v))
				continue
			}

			if t.t == tvariable && prevToken.t != tcomma && !prevToken.isOp() && prevToken.t != tlparenth && prevToken.t != teq {
				f.errs = append(f.errs, errors.New("unexpected variable after "+prevToken.v))
				continue
			}
		}
	}

	if len(mainFunctions) == 0 {
		return errNoMain
	}

	if len(mainFunctions) > 1 {
		definedMains := make([]string, 0, len(mainFunctions))
		for _, f := range mainFunctions {
			definedMains = append(definedMains, fmt.Sprintf("%v:%v", f.file, f.line))
		}
		return fmt.Errorf("%w: %v", errMultipleMains, definedMains)
	}

	// at the end of the parsing, collect all errors and return them
	foundErrors := 0
	for _, f := range functions {
		if len(f.errs) > 0 {
			for _, err := range f.errs {
				log.Printf("%v:%v: %v", f.file, f.line, err)
			}
			foundErrors += len(f.errs)
		}
	}
	if foundErrors > 0 {
		return fmt.Errorf("%w %v found errors", errParse, foundErrors)
	}
	return nil
}
