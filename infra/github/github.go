package github

import (
	"context"
	"errors"
	"fmt"
	"log"
	"path"
	"time"

	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
	"gopkg.in/src-d/go-git.v4/plumbing/transport/http"
)

// Client is a interface that handle about github
type Client interface {
	Clone(repoURI string, dir string) (*git.Repository, error)
	Commit(r *git.Repository, msg string) error
	Push(r *git.Repository) error
	PullRequest(owner, repo, title, head, body, baseBranch string) (string, error)
	ListRepoIssuesSince(owner, repo string, since time.Time, state string, labels []string) ([]*github.Issue, error)
	ListRepoIssues(owner, repo string, state string, labels []string) ([]*github.Issue, error)
	ListRepoLabels(owner, repo string) ([]*github.Label, error)
	SearchRepoLabels(repoID int64, query string) (*github.LabelsSearchResult, *github.Response, error)
}

type ghclient struct {
	client *github.Client
	ctx    context.Context
	user   string
	mail   string
	token  string
}

// NewGitHubClient create GitHubClient implementation
func NewGitHubClient(baseURL string, token string, user string, mail string) (Client, error) {
	if baseURL == "" {
		return nil, errors.New("need to set baseURL")
	}
	uploadURL := path.Join(baseURL, "upload")
	ctx := context.Background()

	if token == "" {
		return nil, errors.New("need to set token")
	}
	if user == "" || mail == "" {
		return nil, errors.New("need to set user, mail for git operation")
	}

	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)

	cli, err := github.NewEnterpriseClient(baseURL, uploadURL, tc)
	if err != nil {
		return nil, fmt.Errorf("creating github enterprise client: %s", err)
	}

	c := &ghclient{
		client: cli,
		ctx:    ctx,
		user:   user,
		mail:   mail,
		token:  token,
	}
	return c, nil
}

func listRepoIssues(listFunc func(pageIdx int) ([]*github.Issue, *github.Response, error)) ([]*github.Issue, error) {
	maxTry := 20 // limit requests for safety
	pageIdx := 1
	issues := make([]*github.Issue, 0)
	for ; maxTry > 0; maxTry-- {
		iss, resp, err := listFunc(pageIdx)
		if err != nil {
			return nil, fmt.Errorf("list issues from repo: %s, pageIdx %d, lastPageIdx %d", err, pageIdx, resp.LastPage)
		}
		issues = append(issues, iss...)
		// last page index is 0 when no more pagination
		if resp.LastPage == 0 {
			break
		}
		pageIdx = resp.NextPage
	}
	if maxTry == 0 {
		return issues, fmt.Errorf("list issues reached to max try: %d", maxTry)
	}
	return issues, nil
}

// ListRepoIssues lists issues since
func (c *ghclient) ListRepoIssuesSince(owner, repo string, since time.Time, state string, labels []string) ([]*github.Issue, error) {
	return listRepoIssues(func(pageIdx int) ([]*github.Issue, *github.Response, error) {
		return c.client.Issues.ListByRepo(c.ctx, owner, repo, &github.IssueListByRepoOptions{
			State:  state,
			Labels: labels,
			Since:  since,
			ListOptions: github.ListOptions{
				Page:    pageIdx,
				PerPage: 30,
			},
		})
	})
}

// ListRepoIssues lists issues
func (c *ghclient) ListRepoIssues(owner, repo string, state string, labels []string) ([]*github.Issue, error) {
	return listRepoIssues(func(pageIdx int) ([]*github.Issue, *github.Response, error) {
		return c.client.Issues.ListByRepo(c.ctx, owner, repo, &github.IssueListByRepoOptions{
			State:  state,
			Labels: labels,
			ListOptions: github.ListOptions{
				Page:    pageIdx,
				PerPage: 30,
			},
		})
	})
}

func (c *ghclient) ListRepoLabels(owner, repo string) ([]*github.Label, error) {
	return listRepoLabels(func(pageIdx int) ([]*github.Label, *github.Response, error) {
		return c.client.Issues.ListLabels(c.ctx, owner, repo, &github.ListOptions{
			Page:    pageIdx,
			PerPage: 30,
		})
	})
}

func (c *ghclient) SearchRepoLabels(repoID int64, query string) (*github.LabelsSearchResult, *github.Response, error) {
	return c.client.Search.Labels(c.ctx, repoID, query, &github.SearchOptions{
		Sort: "created",
	})
}

func listRepoLabels(listFunc func(pageIdx int) ([]*github.Label, *github.Response, error)) ([]*github.Label, error) {
	maxTry := 20 // limit requests for safety
	pageIdx := 1
	labels := make([]*github.Label, 0)
	for ; maxTry > 0; maxTry-- {
		iss, resp, err := listFunc(pageIdx)
		if err != nil {
			return nil, fmt.Errorf("list labels from repo: %s, pageIdx %d, lastPageIdx %d", err, pageIdx, resp.LastPage)
		}
		labels = append(labels, iss...)
		// last page index is 0 when no more pagination
		if resp.LastPage == 0 {
			break
		}
		pageIdx = resp.NextPage
	}
	if maxTry == 0 {
		return labels, fmt.Errorf("list labels reached to max try: %d", maxTry)
	}
	return labels, nil
}

func (c *ghclient) Clone(repoURI string, dir string) (*git.Repository, error) {
	o := &git.CloneOptions{
		Auth: &http.BasicAuth{
			Username: c.user,
			Password: c.token,
		},
		URL: repoURI,
	}
	return git.PlainClone(dir, false, o)
}

func (c *ghclient) Commit(repo *git.Repository, msg string) error {
	w, err := repo.Worktree()
	if err != nil {
		return err
	}
	o := &git.CommitOptions{
		Author: &object.Signature{
			Name:  c.user,
			Email: c.mail,
			When:  time.Now(),
		},
	}
	_, err = w.Commit(msg, o)
	return err
}

func (c *ghclient) Push(repo *git.Repository) error {
	o := &git.PushOptions{
		Auth: &http.BasicAuth{
			Username: c.user,
			Password: c.token,
		},
	}
	return repo.Push(o)
}

func (c *ghclient) PullRequest(owner, repo, title, head, body, baseBranch string) (string, error) {
	npr := github.NewPullRequest{
		Title: &title,
		Head:  &head,
		Base:  &baseBranch,
		Body:  &body,
	}
	pr, _, err := c.client.PullRequests.Create(c.ctx, owner, repo, &npr)
	if err != nil {
		return "", fmt.Errorf("creating PullRequest : %s", err)
	}
	prURL := pr.GetHTMLURL()
	log.Printf("PullRequest created: %s", prURL)
	return prURL, nil
}
