package main

import (
	"fmt"
	"io"
	"sort"
	"strconv"
	"text/tabwriter"
)

type Stats struct {
	Name, Type                 string
	Commit, Addition, Deletion int
}

func (s Stats) Key() string {
	return s.Type + "-" + s.Name
}

func (s Stats) rank() string {
	r := "0"
	switch s.Type {
	case emailType:
		r = "1"
	case authorType:
		r = "2"
	}
	return r + strconv.Itoa(s.Commit) + strconv.Itoa(s.Addition) + strconv.Itoa(s.Deletion)
}

func (s Stats) String() string {
	return fmt.Sprintf(
		"%s\t%s\t%d\t%d\t%d\t",
		s.Name, s.Type,
		s.Commit, s.Addition, s.Deletion,
	)
}

func (s Stats) WriteTo(w io.Writer) error {
	_, err := fmt.Fprintln(w, s.String())
	return err
}

func (s Stats) MoreThan(o Stats) bool {
	if s.rank() == o.rank() {
		return s.Name > o.Name
	}
	return s.rank() > o.rank()
}

func (s Stats) Add(o Stats) Stats {
	if s.Key() != o.Key() {
		return s
	}
	return Stats{
		Name:     s.Name,
		Type:     s.Type,
		Commit:   s.Commit + o.Commit,
		Addition: s.Addition + o.Addition,
		Deletion: s.Deletion + o.Deletion,
	}
}

type Report map[string]Stats

func (r Report) add(s Stats) {
	if _, ok := r[s.Key()]; !ok {
		r[s.Key()] = s
		return
	}
	r[s.Key()] = r[s.Key()].Add(s)
}

func (r Report) list() []Stats {
	entries := make([]Stats, 0, len(r))
	for _, entry := range r {
		entries = append(entries, entry)
	}
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].MoreThan(entries[j])
	})
	return entries
}

func (r Report) writeHeader(w io.Writer) error {
	if _, err := fmt.Fprintln(w, "name\ttype\tcommit\taddition\tdeletion\t"); err != nil {
		return err
	}
	_, err := fmt.Fprintln(w, "---\t---\t---\t---\t---\t")
	return err
}

func (r Report) Print(w io.Writer) error {
	writer := tabwriter.NewWriter(w, 0, 0, 0, ' ', tabwriter.Debug)
	if err := r.writeHeader(writer); err != nil {
		return err
	}
	for _, entry := range r.list() {
		if err := entry.WriteTo(writer); err != nil {
			return err
		}
	}
	return writer.Flush()
}
