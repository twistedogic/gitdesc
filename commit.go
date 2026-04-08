package main

import (
	"github.com/go-git/go-git/v5/plumbing/object"
)

const (
	FileType  = "file"
	AuthorType = "author"
	EmailType  = "email"
)

func authorName(c *object.Commit) string {
	if c.Committer.Name != "" {
		return c.Committer.Name
	}
	return c.Author.Name
}

func authorEmail(c *object.Commit) string {
	if c.Committer.Email != "" {
		return c.Committer.Email
	}
	return c.Author.Email
}

func fileStats(f object.FileStat) *Stats {
	return &Stats{
		Name: f.Name, Type: FileType,
		Commit: 1, Addition: f.Addition, Deletion: f.Deletion,
	}
}

func ToStats(commit *object.Commit) ([]*Stats, error) {
	if commit == nil {
		return nil, nil
	}
	files, err := commit.Stats()
	if err != nil {
		return nil, err
	}
	stats := make([]*Stats, 0, len(files)+2)
	totalAdditions, totalDeletions := 0, 0
	for _, f := range files {
		stats = append(stats, fileStats(f))
		totalAdditions += f.Addition
		totalDeletions += f.Deletion
	}
	if name := authorName(commit); name != "" {
		stats = append(stats, &Stats{
			Name: name, Type: AuthorType,
			Commit: 1, Addition: totalAdditions, Deletion: totalDeletions,
		})
	}
	if email := authorEmail(commit); email != "" {
		stats = append(stats, &Stats{
			Name: email, Type: EmailType,
			Commit: 1, Addition: totalAdditions, Deletion: totalDeletions,
		})
	}
	return stats, nil
}
