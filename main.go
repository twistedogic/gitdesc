package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sort"
	"text/tabwriter"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
)

type Stats struct {
	Key                        string
	Commit, Addition, Deletion int
}

func (s Stats) String() string {
	return fmt.Sprintf("%s\t%d\t%d\t%d\t", s.Key, s.Commit, s.Addition, s.Deletion)
}

func (s *Stats) Add(add, del int) {
	s.Commit += 1
	s.Addition += add
	s.Deletion += del
}

type Report map[string]*Stats

func (r Report) WriteTo(w io.Writer) error {
	if _, err := fmt.Fprintln(w, "name\tcommit\taddition\tdeletion\t"); err != nil {
		return err
	}
	entries := make([]*Stats, 0, len(r))
	for _, entry := range r {
		entries = append(entries, entry)
	}
	sort.Slice(entries, func(i, j int) bool {
		if entries[i].Commit == entries[j].Commit {
			return entries[i].Addition > entries[j].Addition
		}
		return entries[i].Commit > entries[j].Commit
	})
	for _, entry := range entries {
		if _, err := fmt.Fprintln(w, entry.String()); err != nil {
			return err
		}
	}
	return nil
}

type CommitStats struct {
	Author, Hash string
	Files        object.FileStats
}

func ToCommitStats(commit *object.Commit) (CommitStats, error) {
	stats := CommitStats{}
	stats.Author = commit.Committer.Name
	if name := commit.Author.Name; name != "" {
		stats.Author = name
	}
	stats.Hash = commit.Hash.String()
	files, err := commit.Stats()
	if err != nil {
		return stats, err
	}
	stats.Files = files
	return stats, nil
}

func (c CommitStats) Diff() (add, del int) {
	for _, f := range c.Files {
		add += f.Addition
		del += f.Deletion
	}
	return
}

type CommitFilter func(*object.Commit) bool

type Repo struct {
	repo *git.Repository
	w    *tabwriter.Writer
}

func New(w io.Writer, repo *git.Repository) Repo {
	writer := tabwriter.NewWriter(w, 0, 0, 0, ' ', tabwriter.Debug)
	return Repo{repo: repo, w: writer}
}

func NewFromDir() (Repo, error) {
	dir, err := os.Getwd()
	if err != nil {
		return Repo{}, err
	}
	repo := &git.Repository{}
	for dir != "/" {
		repo, err = git.PlainOpen(dir)
		if err == nil {
			return New(os.Stdout, repo), nil
		}
		dir = filepath.Dir(dir)
	}
	return Repo{}, err
}

func (r Repo) commitStats(ch chan CommitStats, filters ...CommitFilter) error {
	defer close(ch)
	iter, err := r.repo.CommitObjects()
	if err != nil {
		return err
	}
	return iter.ForEach(func(c *object.Commit) error {
		for _, f := range filters {
			if !f(c) {
				return nil
			}
		}
		stats, err := ToCommitStats(c)
		if err != nil {
			return err
		}
		ch <- stats
		return err
	})
}

func (r Repo) AuthorStats(filters ...CommitFilter) error {
	ch := make(chan CommitStats)
	go r.commitStats(ch, filters...)
	s := make(Report)
	for stats := range ch {
		key := stats.Author
		if _, ok := s[key]; !ok {
			s[key] = &Stats{Key: key}
		}
		s[key].Add(stats.Diff())
	}
	if err := s.WriteTo(r.w); err != nil {
		return err
	}
	return r.w.Flush()
}

func (r Repo) FileStats(filters ...CommitFilter) error {
	ch := make(chan CommitStats)
	go r.commitStats(ch, filters...)
	s := make(Report)
	for stats := range ch {
		for _, f := range stats.Files {
			key := f.Name
			if _, ok := s[key]; !ok {
				s[key] = &Stats{Key: key}
			}
			s[key].Add(f.Addition, f.Deletion)
		}
	}
	if err := s.WriteTo(r.w); err != nil {
		return err
	}
	return r.w.Flush()
}

func main() {
	repo, err := NewFromDir()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("# Author statistics")
	if err := repo.AuthorStats(); err != nil {
		log.Fatal(err)
	}
	fmt.Println("\n# Files statistics")
	if err := repo.FileStats(); err != nil {
		log.Fatal(err)
	}
}
