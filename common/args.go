package common

import (
	"flag"
	"fmt"
)

var (
	Build BuildInfo // Build information
	Args  Arguments // Parsed arguments
)

type BuildInfo struct {
	Name    string // Build name
	Version string // Build version
	Commit  string // Build commit
	Date    string // Build date
	Summary string // Build summary
}

type Arguments struct {
	Config       string // Argument for config file path
	PrintDisplay bool   // Print the number of the current display
	Lock         string // Argument for lock file path
	Sock         string // Argument for sock file path
	Log          string // Argument for log file path
	VVV          bool   // Argument for very very verbose mode
	VV           bool   // Argument for very verbose mode
	V            bool   // Argument for verbose mode
}

func InitArgs(name, version, commit, date string) {

	// Build information
	Build = BuildInfo{
		Name:    name,
		Version: version,
		Commit:  Truncate(commit, 7),
		Date:    date,
	}
	Build.Summary = fmt.Sprintf("%s v%s-%s, built on %s", Build.Name, Build.Version, Build.Commit, Build.Date)

	// Command line arguments
	flag.StringVar(&Args.Config, "config", ConfigFilePath(Build.Name), "config file path")
	flag.BoolVar(&Args.PrintDisplay, "print-display", false, "number of current display")
	flag.StringVar(&Args.Lock, "lock", fmt.Sprintf("/tmp/%s.lock", Build.Name), "lock file path")
	flag.StringVar(&Args.Sock, "sock", fmt.Sprintf("/tmp/%s.sock", Build.Name), "sock file path")
	flag.StringVar(&Args.Log, "log", fmt.Sprintf("/tmp/%s.log", Build.Name), "log file path")
	flag.BoolVar(&Args.VVV, "vvv", false, "very very verbose mode")
	flag.BoolVar(&Args.VV, "vv", false, "very verbose mode")
	flag.BoolVar(&Args.V, "v", false, "verbose mode")

	// Command line usage text
	flag.CommandLine.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "%s\n\nUsage:\n", Build.Summary)
		flag.PrintDefaults()
	}

	// Parse arguments
	flag.Parse()
}
