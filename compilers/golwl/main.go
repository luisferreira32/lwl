package main

import (
	"flag"
	"log"
	"os"
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
	var output string
	flag.StringVar(&output, "o", "assembled.o", "output file name")

	files := os.Args[1:]
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
	// TODO: make it more obvious we expect functions to be defined in order and file name will matter for that order
	if err := parse(functions); err != nil {
		log.Fatalf("%v", err)
	}

	// generate pseudo-assembly code
	// TODO: add optimized plugins for different architectures
	instructions := passemble(functions)

	// TODO: implement checking the architecture of the host machine and restrict to amd64 linux only for now
	err = magic(instructions, output)
	if err != nil {
		log.Fatalf("%v", err)
	}
}
