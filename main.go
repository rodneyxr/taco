package main

import "github.com/rodneyxr/taco/cmd"

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	cmd.Execute(cmd.BuildInfo{
		Version: version,
		Commit:  commit,
		Date:    date,
	})
}
