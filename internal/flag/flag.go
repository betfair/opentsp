// Package flag defines command-line flags of tsp-* programs.
package flag

import (
	"flag"
	"os"
)

var (
	FilePath  string
	DebugMode bool
	TestMode  bool
)

func Parse(configPath string) {
	flag.StringVar(&FilePath, "f", configPath, "configuration file")
	flag.BoolVar(&DebugMode, "v", false, "verbose mode")
	flag.BoolVar(&TestMode, "t", false, "configuration test")
	flag.Parse()
	if flag.NArg() != 0 {
		flag.Usage()
		os.Exit(1)
	}
}
