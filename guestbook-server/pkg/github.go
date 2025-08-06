package guestbook_server

import (
	"context"
	"fmt"
	"time"

	"github.com/google/go-github/v66/github"
	"golang.org/x/oauth2"
	"gopkg.in/yaml.v3"
)

type GitHubClient struct {
	client *github.Client
	owner  string
	repo   string
	branch string
}

func NewGitHubClient(token, owner, repo, branch string) *GitHubClient {
	if token == "" {
		return nil
	}

	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)

	return &GitHubClient{
		client: client,
		owner:  owner,
		repo:   repo,
		branch: branch,
	}
}

func (g *GitHubClient) CreateGuestbookEntry(ctx context.Context, req GuestbookRequest) error {
	if g == nil {
		return fmt.Errorf("GitHub client not configured")
	}

	entry := req.ToEntry()

	// Generate filename
	filename := fmt.Sprintf("data/guestbook/entry%d.yml", time.Now().Unix())

	// Convert entry to YAML
	yamlData, err := yaml.Marshal(entry)
	if err != nil {
		return fmt.Errorf("failed to marshal YAML: %w", err)
	}

	// Get current main branch
	ref, _, err := g.client.Git.GetRef(ctx, g.owner, g.repo, "refs/heads/"+g.branch)
	if err != nil {
		return fmt.Errorf("failed to get %s branch ref: %w", g.branch, err)
	}

	// Create new branch
	branchName := fmt.Sprintf("guestbook-entry-%d", time.Now().Unix())
	newRef := &github.Reference{
		Ref: github.String("refs/heads/" + branchName),
		Object: &github.GitObject{
			SHA: ref.Object.SHA,
		},
	}

	_, _, err = g.client.Git.CreateRef(ctx, g.owner, g.repo, newRef)
	if err != nil {
		return fmt.Errorf("failed to create branch: %w", err)
	}

	// Create file content
	fileContent := &github.RepositoryContentFileOptions{
		Message: github.String(fmt.Sprintf("New Guestbook Post from %s", entry.Name)),
		Content: yamlData,
		Branch:  github.String(branchName),
	}

	// Create the file
	_, _, err = g.client.Repositories.CreateFile(ctx, g.owner, g.repo, filename, fileContent)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}

	// Create pull request
	title := fmt.Sprintf("New Guestbook Entry from %s", entry.Name)
	body := fmt.Sprintf("New guestbook entry submission:\n\n**Name:** %s\n**Message:** %s\n\nSubmitted on: %s",
		entry.Name, entry.Message, time.Unix(entry.Date, 0).Format("January 2, 2006 15:04:05"))

	pr := &github.NewPullRequest{
		Title: github.String(title),
		Head:  github.String(branchName),
		Base:  github.String(g.branch),
		Body:  github.String(body),
	}

	_, _, err = g.client.PullRequests.Create(ctx, g.owner, g.repo, pr)
	if err != nil {
		return fmt.Errorf("failed to create pull request: %w", err)
	}

	return nil
}
