package main

import (
	"bytes"
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

func (m mockCommit) commit(repo *git.Repository, message string) error {
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
			When: time.Now(),
		},
	}); err != nil {
		return err
	}
	return nil
}

func setupRepo(t *testing.T, commits []mockCommit) *git.Repository {
	t.Helper()
	repo, err := git.Init(memory.NewStorage(), memfs.New())
	require.NoError(t, err)
	for i, c := range commits {
		require.NoError(t, c.commit(repo, fmt.Sprintf("commit %d", i)))
	}
	return repo
}

func Test_Repo(t *testing.T) {
	cases := map[string]struct {
		commits                        []mockCommit
		wantAuthorStats, wantFileStats string
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
			wantAuthorStats: `name|commit|addition|deletion|
a   |3     |5       |2       |
b   |1     |0       |1       |
`,
			wantFileStats: `name        |commit|addition|deletion|
test        |3     |3       |2       |
dir/test    |2     |1       |1       |
dir/dir/test|1     |1       |0       |
`,
		},
	}
	for name := range cases {
		tc := cases[name]
		t.Run(name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			repo := setupRepo(t, tc.commits)
			r := New(buf, repo)
			require.NoError(t, r.AuthorStats())
			require.Equal(t, tc.wantAuthorStats, buf.String(), "author")
			buf.Reset()
			require.NoError(t, r.FileStats())
			require.Equal(t, tc.wantFileStats, buf.String(), "files")
		})
	}
}
