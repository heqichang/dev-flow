package main

import (
	"github.com/devflow/devflow/internal/cmd"
)

var (
	Version = "0.1.0"
	Commit  = "none"
	Date    = "unknown"
)

func main() {
	cmd.SetVersion(Version, Commit, Date)
	cmd.Execute()
}
