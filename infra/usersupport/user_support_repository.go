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

func (r *userSupportRepository) GetUpdatedSupportIssues(state string, since, until time.Time) ([]*github.Issue, error) {
	issues, err := r.ghClient.ListRepoIssuesSince("sataga", "issue-warehouse", since, state, []string{"support"})
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

func (r *userSupportRepository) GetClosedSupportIssues(since, until time.Time) ([]*github.Issue, error) {
	issues, err := r.ghClient.ListRepoIssuesSince("sataga", "issue-warehouse", since, "closed", []string{"support"})
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

func (r *userSupportRepository) GetCurrentOpenNotUpdatedSupportIssues(until time.Time) ([]*github.Issue, error) {
	issues, err := r.ghClient.ListRepoIssues("sataga", "issue-warehouse", "open", []string{"support"})
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

func (r *userSupportRepository) GetCurrentOpenAnyLabelsSupportIssues(state string, labels []string) ([]*github.Issue, error) {
	labels = append(labels, "support")
	return r.ghClient.ListRepoIssues("sataga", "issue-warehouse", state, labels)
}

func (r *userSupportRepository) GetCurrentRepoLabels() ([]*github.Label, error) {
	return r.ghClient.ListRepoLabels("sataga", "issue-warehouse")
}

func (r *userSupportRepository) GetLabelsByQuery(repoID int64, query string) (*github.LabelsSearchResult, *github.Response, error) {
	return r.ghClient.SearchRepoLabels(repoID, query)
}
