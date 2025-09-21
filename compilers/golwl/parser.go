package main

import (
	"errors"
	"fmt"
	"log"
	"slices"
)

var (
	errParse         = errors.New("parse error")
	errNoMain        = errors.New("no main function defined")
	errMultipleMains = errors.New("multiple main functions defined")
)

// not an AST, but close: a binary tree of operations
type operation struct {
	p  [2]*operation   // if nil, op should be variable or constant
	op token           // only if isOp() == true, a variable or constant
	v  map[token]token // if op is a function variable, map from variable to concrete value (variable or constant)
}

type fdeclaration struct {
	name string
	v    []string // variable names of the function
}

// TODO: accept parenthesis syntax in expressions for grouping order
func parse(functions []function) (map[string]*operation, error) {
	functionRegistry := make(map[string][]token) // function name to variable name list
	mainFunctions := make([]function, 0, 1)
	for i := range functions {
		// TODO: make this possible to run in parallel and safer than this
		f := &functions[i] // get the pointer to be able to append to errs
		if _, exists := functionRegistry[f.name]; exists && !f.main {
			f.errs = append(f.errs, errors.New("function "+f.name+" already defined"))
			continue
		}
		functionRegistry[f.name] = make([]token, 0)

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

		// TODO IDEA: parse the tokens with a window of 3 to check the rules
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
				functionRegistry[f.name] = append(functionRegistry[f.name], t)
				continue
			}
			if t.t == teq {
				functionHeaderEnded = true
				continue
			}
			// only expect "(" in function calls or declarations, e.g., g(x)=x+f(1,2)
			if t.t == tlparenth && prevToken.t != tvariable {
				f.errs = append(f.errs, errors.New("unexpected '(' after "+prevToken.v))
				continue
			}

			// FIXME: can only close parenthesis if they're open first
			// parenthesis should only be closed ")" if first opened "(" and there's a variable/constant within the call
			if t.t == trparenth && prevToken.t != tconstant && prevToken.t != tvariable {
				f.errs = append(f.errs, errors.New("unexpected ')' after "+prevToken.v))
				continue
			}
			// FIXME: can only happen within a function call or declaration
			if t.t == tcomma && prevToken.t != tconstant && prevToken.t != tvariable {
				f.errs = append(f.errs, errors.New("unexpected ',' after "+prevToken.v))
				continue
			}
			// a variable must have been declared in the function scope or in global scope (only functions are global for now)
			if t.t == tvariable {
				_, isDeclared := declaredVariables[t.v]
				_, isFunction := functionRegistry[t.v]
				if !isDeclared && !isFunction {
					f.errs = append(f.errs, errors.New("undefined variable "+t.v))
					continue
				}
			}

			// FIXME: cannot have an OP without a const/variable after
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
		return nil, errNoMain
	}

	if len(mainFunctions) > 1 {
		definedMains := make([]string, 0, len(mainFunctions))
		for _, f := range mainFunctions {
			definedMains = append(definedMains, fmt.Sprintf("%v:%v", f.file, f.line))
		}
		return nil, fmt.Errorf("%w: %v", errMultipleMains, definedMains)
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
		return nil, fmt.Errorf("%w %v found errors", errParse, foundErrors)
	}

	ops := make(map[string]*operation, len(functions))
	for _, f := range functions {
		offSet := slices.IndexFunc(f.tkns, func(t token) bool { return t.t == teq })
		offSet++ // account for main (-1) or skip "=" sign

		var (
			opRoot *operation
			op     *operation
			op1    *operation
			op2    *operation
		)
		op2 = &operation{op: f.tkns[offSet]}
		opRoot = op2
		for i := offSet; i+2 < len(f.tkns); i += 2 {
			op1 = op2
			i = parseFunctionVariableMapping(op1, f, functionRegistry, i)
			if i+2 >= len(f.tkns) { // edge case: just one var/constant
				opRoot = op1
				break
			}
			op = &operation{
				op: f.tkns[i+1],
			}
			op2 = &operation{
				op: f.tkns[i+2],
			}
			i = parseFunctionVariableMapping(op2, f, functionRegistry, i)

			op.p[0] = op1
			op.p[1] = op2

			if takesPrecedence(op) && opRoot != nil { // opRoot == nil is first operation check
				op.p[0] = opRoot.p[1]
				opRoot.p[1] = op
			} else {
				op.p[0] = opRoot
				opRoot = op
			}
		}
		fname := f.name
		if f.main {
			fname = mainFuncName
		}
		ops[fname] = opRoot
	}

	// TODO: only print on super duper verbose mode
	printOperations(ops)
	return ops, nil
}

func takesPrecedence(op *operation) bool {
	return op.op.t == tmul || op.op.t == tdiv
}

// NOTE: this is no-op if not within the function registry
func parseFunctionVariableMapping(op *operation, f function, functionRegistry map[string][]token, i int) int {
	// assume validation checked size, parenthesis, commas, etc
	if declaredVariables, isFunction := functionRegistry[op.op.v]; isFunction {
		op.v = make(map[token]token)
		for _, v := range declaredVariables {
			i++ // skip lparenthesis, or comma
			op.v[v] = f.tkns[i]
			i++ // skip rparenthesis, or go to comma
		}
	}

	return i
}

// TODO: pretty print this properly
// TODO: don't do recusive to avoid stack overflow
func printOp(op *operation, level int) {
	if op == nil {
		return
	}
	fmt.Printf("%d> %s \n", level, op.op)
	printOp(op.p[0], level+1)
	printOp(op.p[1], level+1)
}

func printOperations(ops map[string]*operation) {
	for name, tree := range ops {
		fmt.Printf("function: %s\n", name)
		printOp(tree, 0)
	}
}
