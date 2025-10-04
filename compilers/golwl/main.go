package main

import (
	"flag"
	"log"
)

var (
	// version: will be set during compilation onto the binary
	// DO NOT TOUCH IT! I WILL FIND YOU! AND I WILL HURT YOU!
	version string
)

func main() {
	// TODO: logger with levels
	log.Printf("version: %v\n", version)

	// parse input
	opts := parseOptions()

	files := flag.Args()
	if len(files) == 0 {
		log.Fatalf("no input files provided")
	}

	// parse files
	functions, err := tokenize(files)
	if err != nil {
		log.Fatalf("%v", err)
	}

	// handle syntax
	// TODO: gracefully handle syntax and semantic errors since they accumulate per function / line
	// TODO: make it more obvious we expect functions to be defined in order and file provided order name will matter
	ops, err := parse(functions, opts.verbose)
	if err != nil {
		log.Fatalf("%v", err)
	}

	if opts.parseOnly {
		return
	}

	// generate pseudo-assembly code
	// TODO: add optimized plugins for different architectures
	instructions := passemble(ops, opts.verbose)

	if opts.pseudoAssembly {
		return
	}

	// TODO: implement checking the architecture of the host machine and restrict to amd64 linux only for now
	// TODO: allow utilization of other assemblers/linkers besides assumed GNU tools
	err = magic(instructions, opts.output)
	if err != nil {
		log.Fatalf("%v", err)
	}
}
