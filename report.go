package main

import (
	"cmp"
	"io"
	"slices"
	"strconv"
	"strings"

	"github.com/olekukonko/tablewriter"
)

type Stats struct {
	Name, Type                 string
	Commit, Addition, Deletion int
}

func (s *Stats) add(o *Stats) {
	if s.Type != o.Type || s.Name != o.Name {
		return
	}
	s.Addition += o.Addition
	s.Deletion += o.Deletion
	s.Commit += o.Commit
}

func (s *Stats) append(w *tablewriter.Table) {
	w.SetHeader([]string{"name", "commits", "additions", "deletions"})
	w.Append([]string{s.Name, strconv.Itoa(s.Commit), strconv.Itoa(s.Addition), strconv.Itoa(s.Deletion)})
}

func (s *Stats) lineChanges() int { return s.Addition + s.Deletion }

type sortFunc func(*Stats, *Stats) int

func byNumberOfCommitsDesc(a, b *Stats) int { return -cmp.Compare(a.Commit, b.Commit) }
func byLineChangeDesc(a, b *Stats) int      { return -cmp.Compare(a.lineChanges(), b.lineChanges()) }

func printTable(w io.Writer, title string, content map[string]*Stats, f sortFunc) {
	w.Write([]byte(strings.ToUpper(title) + "\n"))
	t := tablewriter.NewWriter(w)
	s := make([]*Stats, 0, len(content))
	for _, stats := range content {
		s = append(s, stats)
	}
	slices.SortFunc(s, f)
	for _, stats := range s {
		stats.append(t)
	}
	t.Render()
}

type Report map[string]map[string]*Stats

func NewReport() Report {
	return make(map[string]map[string]*Stats)
}

func (r Report) add(s *Stats) {
	if _, ok := r[s.Type]; !ok {
		r[s.Type] = make(map[string]*Stats)
	}
	if _, ok := r[s.Type][s.Name]; !ok {
		r[s.Type][s.Name] = s
	} else {
		r[s.Type][s.Name].add(s)
	}
}

func (r Report) Print(w io.Writer, s sortFunc) {
	for t, content := range r {
		printTable(w, t, content, s)
	}
}
