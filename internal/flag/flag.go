// Package flag defines command-line flags of tsp-* programs.
package flag

import (
	"flag"
	"fmt"
	"os"

	"opentsp.org/internal/version"
)

var (
	FilePath    string
	DebugMode   bool
	TestMode    bool
	VersionMode bool
)

func Parse(configPath string) {
	flag.StringVar(&FilePath, "f", configPath, "configuration file")
	flag.BoolVar(&DebugMode, "v", false, "verbose mode")
	flag.BoolVar(&VersionMode, "version", false, "echo version and exit")
	flag.BoolVar(&TestMode, "t", false, "configuration test")
	flag.Parse()
	if flag.NArg() != 0 {
		flag.Usage()
		os.Exit(1)
	}
	if VersionMode {
		fmt.Println(version.ToString())
		os.Exit(0)
	}
}
