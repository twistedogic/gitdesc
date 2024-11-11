package main

import (
	"github.com/go-git/go-git/v5/plumbing/object"
)

const (
	fileType   = "file"
	authorType = "author"
	emailType  = "email"
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

func fileStat(f object.FileStat) *Stats {
	return &Stats{
		Name: f.Name, Type: fileType,
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
	totalAdd, totalDel := 0, 0
	for _, f := range files {
		stats = append(stats, fileStat(f))
		totalAdd += f.Addition
		totalDel += f.Deletion
	}
	if name := authorName(commit); name != "" {
		stats = append(stats, &Stats{
			Name: name, Type: authorType,
			Commit: 1, Addition: totalAdd, Deletion: totalDel,
		})
	}
	if email := authorEmail(commit); email != "" {
		stats = append(stats, &Stats{
			Name: email, Type: emailType,
			Commit: 1, Addition: totalAdd, Deletion: totalDel,
		})
	}
	return stats, nil
}
