package main

import "flag"

type options struct {
	output    string
	verbose   int
	parseOnly bool
}

func parseOptions() options {
	o := options{}
	flag.StringVar(&o.output, "o", "output", "output file name")
	flag.IntVar(&o.verbose, "v", 0, "verbose level")
	flag.BoolVar(&o.parseOnly, "zzz-parse-only", false, "[debug option] exit after parse (no assembly, assembler, nor linker)")
	flag.Parse()
	return o
}
