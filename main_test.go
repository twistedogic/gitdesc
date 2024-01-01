package main

import (
	"fmt"
	"testing"
	"time"

	"github.com/go-git/go-billy/v5/memfs"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/storage/memory"
	"github.com/stretchr/testify/require"
)

type mockFile struct {
	path, content string
}

func (m mockFile) add(tree *git.Worktree) error {
	fs := tree.Filesystem
	if len(m.content) == 0 {
		return fs.Remove(m.path)
	}
	file, err := fs.Create(m.path)
	if err != nil {
		return err
	}
	if _, err := file.Write([]byte(m.content)); err != nil {
		return err
	}
	if err := file.Close(); err != nil {
		return err
	}
	if _, err := tree.Add(m.path); err != nil {
		return err
	}
	return nil
}

type mockCommit struct {
	author string
	files  []mockFile
}

func (m mockCommit) commit(repo *git.Repository, message string, date time.Time) error {
	tree, err := repo.Worktree()
	if err != nil {
		return err
	}
	for _, f := range m.files {
		if err := f.add(tree); err != nil {
			return err
		}
	}
	if _, err := tree.Commit(message, &git.CommitOptions{
		All: true,
		Author: &object.Signature{
			Name: m.author,
			When: date,
		},
	}); err != nil {
		return err
	}
	return nil
}

func setupRepo(t *testing.T, start time.Time, commits []mockCommit) *git.Repository {
	t.Helper()
	repo, err := git.Init(memory.NewStorage(), memfs.New())
	require.NoError(t, err)
	for i, c := range commits {
		commitTime := start.AddDate(0, 0, i)
		require.NoError(t, c.commit(repo, fmt.Sprintf("commit %d", i), commitTime))
	}
	return repo
}

func Test_Repo(t *testing.T) {
	cases := map[string]struct {
		start        time.Time
		commits      []mockCommit
		valid, limit *object.CommitFilter
		want         []Stats
	}{
		"base": {
			commits: []mockCommit{
				{
					author: "a",
					files: []mockFile{
						{path: "test", content: "something"},
						{path: "dir/test", content: "something"},
						{path: "dir/dir/test", content: "something"},
					},
				},
				{
					author: "a",
					files: []mockFile{
						{path: "test", content: "otherthing"},
					},
				},
				{
					author: "a",
					files: []mockFile{
						{path: "test", content: "other"},
					},
				},
				{
					author: "b",
					files: []mockFile{
						{path: "dir/test", content: ""},
					},
				},
			},
			want: []Stats{
				{Name: "a", Type: authorType, Commit: 3, Addition: 5, Deletion: 2},
				{Name: "b", Type: authorType, Commit: 1, Addition: 0, Deletion: 1},
				{Name: "test", Type: fileType, Commit: 3, Addition: 3, Deletion: 2},
				{Name: "dir/test", Type: fileType, Commit: 2, Addition: 1, Deletion: 1},
				{Name: "dir/dir/test", Type: fileType, Commit: 1, Addition: 1, Deletion: 0},
			},
		},
		"a only": {
			commits: []mockCommit{
				{
					author: "a",
					files: []mockFile{
						{path: "test", content: "something"},
						{path: "dir/test", content: "something"},
						{path: "dir/dir/test", content: "something"},
					},
				},
				{
					author: "a",
					files: []mockFile{
						{path: "test", content: "otherthing"},
					},
				},
				{
					author: "a",
					files: []mockFile{
						{path: "test", content: "other"},
					},
				},
				{
					author: "b",
					files: []mockFile{
						{path: "dir/test", content: ""},
					},
				},
			},
			valid: compose(MatchAuthorRegex("a")),
			want: []Stats{
				{Name: "a", Type: authorType, Commit: 3, Addition: 5, Deletion: 2},
				{Name: "test", Type: fileType, Commit: 3, Addition: 3, Deletion: 2},
				{Name: "dir/test", Type: fileType, Commit: 1, Addition: 1, Deletion: 0},
				{Name: "dir/dir/test", Type: fileType, Commit: 1, Addition: 1, Deletion: 0},
			},
		},
		"limit 2 last commits only": {
			commits: []mockCommit{
				{
					author: "a",
					files: []mockFile{
						{path: "test", content: "something"},
						{path: "dir/test", content: "something"},
						{path: "dir/dir/test", content: "something"},
					},
				},
				{
					author: "a",
					files: []mockFile{
						{path: "test", content: "otherthing"},
					},
				},
				{
					author: "a",
					files: []mockFile{
						{path: "test", content: "other"},
					},
				},
				{
					author: "b",
					files: []mockFile{
						{path: "dir/test", content: ""},
					},
				},
			},
			limit: compose(LimitN(2)),
			want: []Stats{
				{Name: "a", Type: authorType, Commit: 2, Addition: 2, Deletion: 2},
				{Name: "b", Type: authorType, Commit: 1, Addition: 0, Deletion: 1},
				{Name: "test", Type: fileType, Commit: 2, Addition: 2, Deletion: 2},
				{Name: "dir/test", Type: fileType, Commit: 1, Addition: 0, Deletion: 1},
			},
		},
	}
	for name := range cases {
		tc := cases[name]
		t.Run(name, func(t *testing.T) {
			repo := setupRepo(t, tc.start, tc.commits)
			r := New(repo)
			report, err := r.Report(tc.valid, tc.limit)
			require.NoError(t, err)
			require.Equal(t, tc.want, report.list(), "stats")
		})
	}
}
