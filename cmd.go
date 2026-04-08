package main

import (
	"flag"
	"log"
	"os"
)

type cmd struct {
	filterFlag *commitFilter
	flagSet    *flag.FlagSet
	help       bool
}

func NewCmd(f *flag.FlagSet) *cmd {
	filterFlag := NewCommitFilter()
	c := &cmd{filterFlag: filterFlag, flagSet: f}
	filterFlag.SetFlag(f)
	f.BoolVar(&c.help, "help", false, "print help")
	return c
}

func (c *cmd) Execute(args []string) {
	if err := c.flagSet.Parse(args); err != nil {
		log.Println(err)
		os.Exit(1)
	}
	if c.help {
		c.flagSet.Usage()
		os.Exit(0)
	}
	repo, err := NewRepoFromDir()
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
	report, err := repo.Report(c.filterFlag.Valid(), c.filterFlag.Limit())
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
	report.Print(os.Stdout, byNumberOfCommitsDesc)
}
