package main

import (
	"regexp"
	"time"

	"github.com/go-git/go-git/v5/plumbing/object"
)

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

func LimitBefore(before time.Time) object.CommitFilter {
	return func(c *object.Commit) bool {
		return c.Committer.When.After(before)
	}
}

func LimitAfter(after time.Time) object.CommitFilter {
	return func(c *object.Commit) bool {
		return !c.Committer.When.Before(after)
	}
}
