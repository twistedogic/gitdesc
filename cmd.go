package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/go-git/go-git/v5/plumbing/object"
)

const dateFormat = "2006-01-02"

type DateValue struct {
	date time.Time
}

func (d *DateValue) String() string {
	return d.date.Format(dateFormat)
}

func (d *DateValue) Set(s string) error {
	t, err := time.Parse(s, dateFormat)
	if err != nil {
		return err
	}
	d.date = t
	return nil
}

type reportType string

func (r *reportType) String() string { return string(*r) }
func (r *reportType) Set(s string) error {
	switch s {
	case emailType, fileType, authorType:
		*r = reportType(s)
		return nil
	}
	return fmt.Errorf("must be one of the following: %v", []string{
		emailType, fileType, authorType,
	})
}

type gitdescCmd struct {
	f             *flag.FlagSet
	before, after *DateValue
	report        reportType
	last          int
	help          bool
	messageRegex  string
}

func NewCmd(f *flag.FlagSet) *gitdescCmd {
	g := &gitdescCmd{f: f, before: &DateValue{}, after: &DateValue{}}
	f.BoolVar(&g.help, "help", false, "print help")
	f.Var(g.before, "before", "commit before provided date")
	f.Var(g.after, "after", "commit after provided date")
	f.IntVar(&g.last, "last", 0, "last N commits")
	f.StringVar(&g.messageRegex, "re", "", "regex for commit message match")
	f.Var(&g.report, "filter", "filter on type")
	return g
}

func (g *gitdescCmd) valid() *object.CommitFilter {
	filters := []object.CommitFilter{}
	if g.messageRegex != "" {
		filters = append(filters, MatchMessageRegex(g.messageRegex))
	}
	if len(filters) == 0 {
		return nil
	}
	return compose(filters...)
}

func (g *gitdescCmd) limit() *object.CommitFilter {
	filters := []object.CommitFilter{}
	if g.last > 0 {
		filters = append(filters, LimitN(g.last))
	}
	if before := g.before.date; !before.IsZero() {
		filters = append(filters, LimitBefore(before))
	}
	if after := g.after.date; !after.IsZero() {
		filters = append(filters, LimitAfter(after))
	}
	if len(filters) == 0 {
		return nil
	}
	return compose(filters...)
}

func (g *gitdescCmd) filter() StatsFilter {
	if g.report != "" {
		return func(s Stats) bool { return s.Type == g.report.String() }
	}
	return func(s Stats) bool { return true }
}

func (g *gitdescCmd) Execute(args []string) {
	if err := g.f.Parse(args); err != nil {
		log.Println(err)
		os.Exit(1)
	}
	if g.help {
		g.f.Usage()
		return
	}
	repo, err := NewFromDir()
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
	report, err := repo.Report(g.valid(), g.limit())
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
	if err := report.Print(os.Stdout, g.filter()); err != nil {
		log.Println(err)
		os.Exit(1)
	}
}
