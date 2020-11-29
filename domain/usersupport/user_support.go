package usersupport

import (
	"fmt"
	"sort"
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
	GetUpdatedSupportIssues(since, until time.Time) ([]*github.Issue, error)
	GetCurrentOpenSupportIssues() ([]*github.Issue, error)
}

type userSupport struct {
	repo Repository
}

// Stats is stats data from GitHub
type Stats struct {
	NumCreatedIssues     int                      `yaml:"num_created_issues"`
	NumClosedIssues      int                      `yaml:"num_closed_issues"`
	NumOpenIssues        int                      `yaml:"num_open_issues`
	NumUpdatedIssues     int                      `yaml:"num_updated_issues`
	NumCommentsPerIssue  map[string]int           `yaml:"num_comments_per_issue`
	OpenDurationPerIssue map[string]time.Duration `yaml:"open_duration_per_issue"`
}

// NewUserSupport creates UserSupport
func NewUserSupport(repo Repository) UserSupport {
	return &userSupport{
		repo: repo,
	}
}

// GetUserSupportStats
func (us *userSupport) GetUserSupportStats(since, until time.Time) (*Stats, error) {
	upi, err := us.repo.GetUpdatedSupportIssues(since, until)
	if err != nil {
		return nil, fmt.Errorf("get updated issues : %s", err)
	}
	opi, err := us.repo.GetCurrentOpenSupportIssues()
	if err != nil {
		return nil, fmt.Errorf("get open issues : %s", err)
	}
	usStats := &Stats{
		NumUpdatedIssues:     len(upi),
		NumOpenIssues:        len(opi),
		NumCommentsPerIssue:  make(map[string]int, len(upi)),
		OpenDurationPerIssue: make(map[string]time.Duration, len(upi)),
	}
	numCreated, numClosed := 0, 0
	for _, issue := range upi {
		if issue.State != nil && *issue.State == "closed" {
			numClosed++
		}
		if issue.CreatedAt != nil && issue.CreatedAt.After(since) && issue.CreatedAt.Before(until) {
			numCreated++
		}
		title := fmt.Sprintf("[%s]%s", *issue.State, *issue.Title)
		if issue.Comments != nil {
			usStats.NumCommentsPerIssue[title] = *issue.Comments
		}
		if issue.State != nil && *issue.State == "closed" {
			usStats.OpenDurationPerIssue[title] = issue.ClosedAt.Sub(*issue.CreatedAt)
		} else {
			usStats.OpenDurationPerIssue[title] = issue.UpdatedAt.Sub(*issue.CreatedAt)
		}
	}
	usStats.NumClosedIssues = numClosed
	usStats.NumCreatedIssues = numCreated
	return usStats, nil
}

// GenReport generate report
func (s *Stats) GenReport() string {
	var sb strings.Builder
	sb.WriteString("項目, 情報\n")
	sb.WriteString(fmt.Sprintf("対応が必要なチケット数, %d\n", s.NumOpenIssues))
	sb.WriteString(fmt.Sprintf("新規作成されたチケット数, %d\n", s.NumCreatedIssues))
	sb.WriteString(fmt.Sprintf("対応したチケット数, %d\n", s.NumUpdatedIssues))
	sb.WriteString(fmt.Sprintf("クローズしたチケット数, %d\n", s.NumClosedIssues))

	{
		// 経過時間を出すロジック
		type kv struct {
			Key string
			Val time.Duration
		}
		var kvArr []kv
		for k, v := range s.OpenDurationPerIssue {
			kvArr = append(kvArr, kv{k, v})
		}
		// sort by duration
		sort.Slice(kvArr, func(i, j int) bool {
			return kvArr[i].Val > kvArr[j].Val
		})
		for _, kv := range kvArr{
			// if >=10 {
			// 	break
			// }
			totalHours := int(kv.Val.Hours())
			dates := totalHours / 24
			hours := totalHours % 24
			sb.WriteString(fmt.Sprintf("%s, %dd %dh\n",kv.Key, dates, hours))
		}
	}

	return sb.String()
}
