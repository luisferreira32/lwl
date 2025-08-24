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
	// TODO: implement checking the architecture of the host machine and restrict to amd64 linux only for now
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

	// 7. generate pseudo-assembly code
	// TODO: make this a pseudo-assembly and add optimized plugins for different architectures
	outputFile, err := os.Create(output)
	if err != nil {
		log.Fatalf("create %v: %v", output, err)
	}
	defer func() {
		if err := outputFile.Close(); err != nil {
			log.Fatalf("close %v: %v", output, err)
		}
	}()

	// hacks
	for _, f := range functions {
		outputFile.WriteString("# " + f.file + ":" + string(rune(f.line)) + "\n")
	}
}
