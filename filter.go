package main

import (
	"flag"
	"fmt"
	"regexp"
	"time"

	"github.com/go-git/go-git/v5/plumbing/object"
)

type dateValue struct {
	date time.Time
}

func (d dateValue) Date() time.Time { return d.date }

func (d *dateValue) String() string {
	return d.date.Format(time.DateOnly)
}

func (d *dateValue) Set(s string) error {
	t, err := time.Parse(time.DateOnly, s)
	if err != nil {
		return err
	}
	d.date = t
	return nil
}

type commitFilterFlag struct {
	author, email, file string
	before, after       dateValue
	lastN               int
}

func NewCommitFilterFlag() *commitFilterFlag {
	return &commitFilterFlag{lastN: 100}
}

func (c *commitFilterFlag) SetFlag(f *flag.FlagSet) {
	f.StringVar(&c.author, "author", "", "commit author regex pattern")
	f.StringVar(&c.email, "email", "", "commit email regex pattern")
	f.StringVar(&c.file, "file", "", "commit file regex pattern")
	f.Var(&c.before, "before", fmt.Sprintf("commit before date (format: %s)", time.DateOnly))
	f.Var(&c.after, "after", fmt.Sprintf("commit after date (format: %s)", time.DateOnly))
	f.IntVar(&c.lastN, "n", 100, "number of commits")
}

func (c *commitFilterFlag) Valid() *object.CommitFilter {
	filters := []object.CommitFilter{func(c *object.Commit) bool { return true }}
	if c.author != "" {
		filters = append(filters, MatchAuthorRegex(c.author))
	}
	if c.email != "" {
		filters = append(filters, MatchEmailRegex(c.email))
	}
	if c.file != "" {
		filters = append(filters, MatchFileRegex(c.file))
	}
	return compose(filters...)
}

func (c *commitFilterFlag) Limit() *object.CommitFilter {
	filters := []object.CommitFilter{LimitN(c.lastN)}
	if before := c.before.Date(); !before.IsZero() {
		filters = append(filters, isBefore(before))
	}
	if after := c.after.Date(); !after.IsZero() {
		filters = append(filters, isAfter(after))
	}
	return compose(filters...)
}

func compose(filters ...object.CommitFilter) *object.CommitFilter {
	filter := object.CommitFilter(func(c *object.Commit) bool {
		for _, f := range filters {
			if !f(c) {
				return false
			}
		}
		return true
	})
	return &filter
}

func not(f object.CommitFilter) object.CommitFilter {
	return func(c *object.Commit) bool {
		return !f(c)
	}
}

func MatchAuthorRegex(expr string) object.CommitFilter {
	re := regexp.MustCompile(expr)
	return func(c *object.Commit) bool {
		return re.MatchString(c.Author.Name)
	}
}

func MatchEmailRegex(expr string) object.CommitFilter {
	re := regexp.MustCompile(expr)
	return func(c *object.Commit) bool {
		return re.MatchString(c.Author.Email)
	}
}

func MatchMessageRegex(expr string) object.CommitFilter {
	re := regexp.MustCompile(expr)
	return func(c *object.Commit) bool {
		return re.MatchString(c.Message)
	}
}

func LimitN(n int) object.CommitFilter {
	seen := 0
	return func(c *object.Commit) bool {
		if seen >= n {
			return true
		}
		seen += 1
		return false
	}
}

func isBefore(before time.Time) object.CommitFilter {
	return func(c *object.Commit) bool {
		return c.Committer.When.Before(before)
	}
}

func isAfter(after time.Time) object.CommitFilter {
	return func(c *object.Commit) bool {
		return !c.Committer.When.After(after)
	}
}

func MatchFileRegex(expr string) object.CommitFilter {
	re := regexp.MustCompile(expr)
	return func(c *object.Commit) bool {
		stats, err := c.Stats()
		if err != nil {
			return false
		}
		for _, s := range stats {
			if re.MatchString(s.Name) {
				return true
			}
		}
		return false
	}
}
