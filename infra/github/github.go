package github

import (
	"context"
	"errors"
	"fmt"
	"path"

	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

// Client is a interface that handle about github
type Client interface {
	ListRepoIssues(owner, repo string, state string, labels []string) ([]*github.Issue, error)
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
