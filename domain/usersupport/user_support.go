package usersupport

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/go-github/github"
)

// UserSupport is interface for getting user support info
type UserSupport interface {
	GetUserSupportStats(since, until time.Time) (*Stats, error)
}

// Repository r/w data which usersupport domain requires
type Repository interface {
	GetCurrentOpenSupportIssues() ([]*github.Issue, error)
}

type userSupport struct {
	repo Repository
}

// Stats is stats data from GitHub
type Stats struct {
	NumCreatedIssues int `yaml:"num_created_issues"`
}

// NewUserSupport creates UserSupport
func NewUserSupport(repo Repository) UserSupport {
	return &userSupport{
		repo: repo,
	}
}

// GetUserSupportStats
func (us *userSupport) GetUserSupportStats(since, until time.Time) (*Stats, error) {
	opi, err := us.repo.GetCurrentOpenSupportIssues()
	if err != nil {
		return nil, fmt.Errorf("get open issues : %s", err)
	}
	usStats := &Stats{
		NumCreatedIssues: len(opi),
	}
	return usStats, nil
}

func (s *Stats) GenReport() string {
	var sb strings.Builder
	sb.WriteString("項目, 情報\n")
	sb.WriteString(fmt.Sprintf("新規作成されたチケット数, %d\n", s.NumCreatedIssues))

	return sb.String()
}
