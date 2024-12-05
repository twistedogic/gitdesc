package main

import (
	"flag"
	"log"
	"os"
)

type gitdescCmd struct {
	c    *commitFilterFlag
	f    *flag.FlagSet
	help bool
}

func NewCmd(f *flag.FlagSet) *gitdescCmd {
	c := NewCommitFilterFlag()
	g := &gitdescCmd{c: c, f: f}
	c.SetFlag(f)
	f.BoolVar(&g.help, "help", false, "print help")
	return g
}

func (g *gitdescCmd) Execute(args []string) {
	if err := g.f.Parse(args); err != nil {
		log.Println(err)
		os.Exit(1)
	}
	if g.help {
		g.f.Usage()
		os.Exit(0)
	}
	repo, err := NewFromDir()
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
	report, err := repo.Report(g.c.Valid(), g.c.Limit())
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
	report.Print(os.Stdout, byNumberOfCommitsDesc)
}
