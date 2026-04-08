package main

import (
	"log"
	"os"
	"path/filepath"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
)

type Repo struct {
	repo *git.Repository
}

func NewRepo(repo *git.Repository) Repo {
	return Repo{repo: repo}
}

func NewRepoFromDir() (Repo, error) {
	dir, err := os.Getwd()
	if err != nil {
		return Repo{}, err
	}
	repo := &git.Repository{}
	for dir != "/" {
		repo, err = git.PlainOpen(dir)
		if err == nil {
			return NewRepo(repo), nil
		}
		dir = filepath.Dir(dir)
	}
	return Repo{}, err
}

func (r Repo) commitStats(isValid, isLimit *object.CommitFilter) (<-chan *Stats, error) {
	ref, err := r.repo.Head()
	if err != nil {
		return nil, err
	}
	head, err := r.repo.CommitObject(ref.Hash())
	if err != nil {
		return nil, err
	}
	statsChan := make(chan *Stats)
	iter := object.NewFilterCommitIter(head, isValid, isLimit)
	go func() {
		defer iter.Close()
		defer close(statsChan)
		if err := iter.ForEach(func(commit *object.Commit) error {
			stats, err := ToStats(commit)
			if err != nil {
				return err
			}
			for _, s := range stats {
				statsChan <- s
			}
			return err
		}); err != nil {
			log.Fatal(err)
		}
	}()
	return statsChan, nil

}

func (r Repo) Report(isValid, isLimit *object.CommitFilter) (Report, error) {
	statsCh, err := r.commitStats(isValid, isLimit)
	if err != nil {
		return nil, err
	}
	report := make(Report)
	for stats := range statsCh {
		report.add(stats)
	}
	return report, nil
}
