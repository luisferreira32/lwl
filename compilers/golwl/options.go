package main

import "flag"

type options struct {
	output         string
	verbose        int
	parseOnly      bool
	pseudoAssembly bool
}

const (
	verboseERROR = iota
	verboseWARN
	verboseINFO
	verboseDEBUG
)

func parseOptions() options {
	o := options{}
	flag.StringVar(&o.output, "o", "output", "output file name")
	flag.IntVar(&o.verbose, "v", 0, "verbose level")
	flag.BoolVar(&o.parseOnly, "zzz-parse-only", false, "[debug option] exit after parse (no assembly, assembler, nor linker)")
	flag.BoolVar(&o.pseudoAssembly, "zzz-pseudo", false, "[debug option] exit after pseudo assembly creation (no assembler, nor linker)")
	flag.Parse()

	if o.verbose > verboseDEBUG {
		o.verbose = verboseDEBUG
	}
	if o.verbose < verboseERROR {
		o.verbose = verboseERROR
	}
	return o
}
