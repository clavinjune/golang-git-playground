package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/google/go-github/v49/github"
	"golang.org/x/oauth2"
	"os"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
)

var (
	sshKeyFlag      = flag.String("key", "", "ssh key location")
	emailFlag       = flag.String("email", "", "github email")
	githubTokenFlag = flag.String("token", "", "github token")
)

func createBranchAndCommit(sshKey, email string) {
	auth, err := ssh.NewPublicKeysFromFile("git", sshKey, "")
	if err != nil {
		panic(err)
	}

	repo, err := git.PlainClone("./dst", false, &git.CloneOptions{
		URL:           "git@github.com:clavinjune/golang-git-playground.git",
		Depth:         1,
		SingleBranch:  true,
		ReferenceName: "refs/heads/main",
		Auth:          auth,
	})

	if err != nil {
		panic(err)
	}

	tree, err := repo.Worktree()
	if err != nil {
		panic(err)
	}

	if err := tree.Checkout(&git.CheckoutOptions{
		Branch: plumbing.NewBranchReferenceName("testings"),
		Create: true,
	}); err != nil {
		panic(err)
	}

	f, err := os.OpenFile("./dst/README.md", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}
	f.WriteString("\nappended")
	f.Close()

	tree.Add("README.md")
	commit, err := tree.Commit("append README.md", &git.CommitOptions{
		Author: &object.Signature{
			Email: email,
			When:  time.Now(),
		},
	})
	if err != nil {
		panic(err)
	}

	repo.CommitObject(commit)
	if err := repo.Push(&git.PushOptions{
		RemoteName: "origin",
		Auth:       auth,
	}); err != nil {
		panic(err)
	}
}

func main() {
	flag.Parse()
	defer func() {
		if err := os.RemoveAll("./dst"); err != nil {
			panic(err)
		}
	}()

	createBranchAndCommit(*sshKeyFlag, *emailFlag)

	st := oauth2.StaticTokenSource(&oauth2.Token{
		AccessToken: *githubTokenFlag,
	})

	newPR := &github.NewPullRequest{
		Title:               github.String("Bot testing"),
		Head:                github.String("refs/heads/testings"),
		Base:                github.String("refs/heads/main"),
		Body:                github.String("This is the description of the PR created with the package `github.com/google/go-github/github`"),
		MaintainerCanModify: github.Bool(true),
	}

	client := oauth2.NewClient(context.Background(), st)
	pr, _, err := github.NewClient(client).PullRequests.Create(
		context.Background(),
		"clavinjune",
		"golang-git-playground",
		newPR,
	)

	if err != nil {
		panic(err)
	}

	fmt.Println(pr.GetNumber())
}
