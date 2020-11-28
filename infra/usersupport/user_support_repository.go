package usersupport

import (
	"fmt"
	"time"

	"github.com/google/go-github/github"
	dus "github.com/sataga/go-github-sample/domain/usersupport"
	igh "github.com/sataga/go-github-sample/infra/github"
)

type userSupportRepository struct {
	ghClient igh.Client
}

// NewUsersupportRepository creates UsersupportRepository implementation
func NewUsersupportRepository(ghClient igh.Client) dus.Repository {
	return &userSupportRepository{
		ghClient: ghClient,
	}
}

func (r *userSupportRepository) GetUpdatedSupportIssues(since, until time.Time) ([]*github.Issue, error) {
	issues, err := r.ghClient.ListRepoIssues("sataga", "issue-warehouse", "all", []string{"support"})
	if err != nil {
		return nil, fmt.Errorf("list repo issues: %s", err)
	}
	iss := make([]*github.Issue, 0, len(issues))
	for _, is := range issues {
		if is.UpdatedAt.Before(until) {
			iss = append(iss, is)
		}
	}
	return iss, nil
}

func (r *userSupportRepository) GetCurrentOpenSupportIssues() ([]*github.Issue, error) {
	return r.ghClient.ListRepoIssues("sataga", "issue-warehouse", "open", []string{"support"})
}
