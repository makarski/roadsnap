package main

import (
	"flag"
	"fmt"
	"os"
	"path"

	"github.com/makarski/roadsnap/cmd"
	"github.com/makarski/roadsnap/config"
)

var fls *flag.FlagSet

func init() {
	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	fls = flag.NewFlagSet("", flag.ExitOnError)
	fls.StringVar(&cmd.InArgs.Dir, "dir", wd, "Work directory for cache and reports")
	fls.BoolVar(&cmd.InArgs.Interactive, "i", false, "Run the tool in the interactive mode")
	fls.StringVar(&cmd.InArgs.ConfigFile, "config", path.Join(wd, config.DefaultFileName), "Full config filekey")

	fls.Usage = printHelp
}

func printHelp() {
	txt := `
Roadsnap - fetches jira project snapshots by epic

USAGE:
  roadsnap [OPTIONS] [SUBCOMMAND]

SUBCOMMANDS:
  cache - Cache JIRA epics
  list  - Generate report
  chart - Generate stacked bar chart for with stats

OPTIONS:
`

	fmt.Fprint(fls.Output(), txt)
	fls.PrintDefaults()
}

func main() {
	// parse flags
	fls.Parse(os.Args[1:])

	cmdName := fls.Arg(0)

	if err := cmd.Run(cmdName); err != nil {
		panic(err)
	}
}
