# Light Weight Language

A light weight language created to learn how to do a compiler.

## Rules to Follow Coherently
1. We want it simple.
2. Imports are for the weak, you should be able to write your entire OS in **ONE** file, all with **ORIGINAL** implementations.
3. Return value is **MANDATORY**, but there can be only **ONE**.
4. There are **NO** floats! We only deal **EXACT** numbers.

## Ever changing syntax

The objective of the LWL is to be light weight and that comes with simplicity. The syntax is proposed as follows, with the RFCs in mind:
* Newlines are loved: any assign, operation, condition, curly brace, declaration, return statement, etc., should live in it's line.
* Spaces are the rule: every word, operation, or token, must have a space (or more) to separate it from others.
* Identifiers are not picky, any sequence of bytes can be an identifier.
* Comments are for the weak, the code documents itself!
* Pointers... will be a second thought.
* If we have an if condition we can have an if not, what else?
* We don't lie to ourselves, bools are, in fact, just 8-bit integers.

## Contribution

Feel free to open issues, pull requests, and/or propose changes in the language. The RFCs (rules to follow coherently) should be... followed.

## Roadmap

- [ ] MVP (AST + syntax check + assembly)
- [ ] Add concept of structs and arrays
- [ ] Enable compiler plugins
- [ ] Demo compiler plugin string to 8 bit integer array
- [ ] Demo compiler plugin to decide integer type based on architecture to compile