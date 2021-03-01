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
func NewUserSupportRepository(ghClient igh.Client) dus.Repository {
	return &userSupportRepository{
		ghClient: ghClient,
	}
}

func (r *userSupportRepository) GetUpdatedSupportIssues(since, until time.Time) ([]*github.Issue, error) {
	issues, err := r.ghClient.ListRepoIssuesSince("sataga", "issue-warehouse", since, "all", []string{"PF_Support"})
	if err != nil {
		return nil, fmt.Errorf("list repo issues: %s", err)
	}
	iss := make([]*github.Issue, 0, len(issues))
	for _, is := range issues {
		if is.UpdatedAt.After(since) && is.UpdatedAt.Before(until) {
			iss = append(iss, is)
		}
	}
	return iss, nil
}

func (r *userSupportRepository) GetClosedSupportIssues(since, until time.Time) ([]*github.Issue, error) {
	issues, err := r.ghClient.ListRepoIssuesSince("sataga", "issue-warehouse", since, "closed", []string{"PF_Support"})
	if err != nil {
		return nil, fmt.Errorf("list repo issues: %s", err)
	}
	iss := make([]*github.Issue, 0, len(issues))
	for _, is := range issues {
		if is.ClosedAt.After(since) && is.ClosedAt.Before(until) {
			iss = append(iss, is)
		}
	}
	return iss, nil
}

func (r *userSupportRepository) GetCurrentOpenNotUpdatedSupportIssues(until time.Time) ([]*github.Issue, error) {
	issues, err := r.ghClient.ListRepoIssues("sataga", "issue-warehouse", "open", []string{"PF_Support"})
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
	return r.ghClient.ListRepoIssues("sataga", "issue-warehouse", "open", []string{"PF_Support"})
}

func (r *userSupportRepository) GetCreatedSupportIssues(since, until time.Time) ([]*github.Issue, error) {
	query := fmt.Sprintf("repo:sataga/issue-warehouse is:issue created:%s..%s label:PF_Support", since.Format("2006-01-02"), until.Format("2006-01-02"))
	result, _ := r.ghClient.SearchIssuesByQuery(query)
	iss := make([]*github.Issue, 0, len(result))
	for _, is := range result {
		iss = append(iss, &github.Issue{
			Title:     is.Title,
			CreatedAt: is.CreatedAt,
			UpdatedAt: is.UpdatedAt,
			ClosedAt:  is.ClosedAt,
			State:     is.State,
			Assignees: is.Assignees,
			Labels:    is.Labels,
			HTMLURL:   is.HTMLURL,
			Body:      is.Body,
			Comments:  is.Comments,
		})
	}
	return iss, nil
}

func (r *userSupportRepository) GetLabelsByQuery(query string) ([]*github.LabelResult, error) {
	repoID, _ := r.ghClient.GetRepoID("sataga", "issue-warehouse")
	return r.ghClient.SearchLabelsByQuery(repoID, query)
}

func (r *userSupportRepository) GetIssueComments(number int) ([]*github.IssueComment, error) {
	return r.ghClient.ListIssueComments("sataga", "issue-warehouse", number)
}
