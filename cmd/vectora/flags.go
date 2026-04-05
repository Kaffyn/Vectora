package main

import (
	"flag"
)

// Flags holds command-line flags
type Flags struct {
	TestIPC bool
	Tests   bool
	Verbose bool
	Help    bool
}

// ParseFlags parses command-line arguments
func ParseFlags() *Flags {
	f := &Flags{}
	flag.BoolVar(&f.TestIPC, "test-ipc", false, "Test IPC connection")
	flag.BoolVar(&f.Tests, "tests", false, "Run system tests")
	flag.BoolVar(&f.Verbose, "verbose", false, "Verbose output")
	flag.BoolVar(&f.Help, "help", false, "Show help")
	flag.Parse()
	return f
}
