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
	GetCurrentOpenAnyLabelsSupportIssues(labels []string) ([]*github.Issue, error)
	GetCurrentRepoLabels() ([]*github.Label, error)
	GetLabelsByQuery(repoID int64, query string) (*github.LabelsSearchResult, *github.Response, error)
}

type userSupport struct {
	repo Repository
}

// Stats is stats data from GitHub
type Stats struct {
	NumCreatedIssues                int                      `yaml:"num_created_issues"`
	NumClosedIssues                 int                      `yaml:"num_closed_issues"`
	UrgencyHighIssues               int                      `yaml:"num_urgency_high_issues"`
	UrgencyLowIssues                int                      `yaml:"num_urgency_low_issues"`
	UrgencyHighDifficultyHighIssues int                      `yaml:"num_urgency_high_difficulty_high_issues"`
	UrgencyHighDifficultyLowIssues  int                      `yaml:"num_urgency_high_difficulty_low_issues"`
	UrgencyLowDifficultyHighIssues  int                      `yaml:"num_urgency_low_difficulty_high_issues"`
	UrgencyLowDifficultyLowIssues   int                      `yaml:"num_urgency_low_difficulty_low_issues"`
	NumOpenIssues                   int                      `yaml:"num_open_issues"`
	NumUpdatedIssues                int                      `yaml:"num_updated_issues"`
	NumCommentsPerIssue             map[string]int           `yaml:"num_comments_per_issue"`
	OpenDurationPerIssue            map[string]time.Duration `yaml:"open_duration_per_issue"`
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
	labels, err := us.repo.GetCurrentRepoLabels()
	if err != nil {
		return nil, fmt.Errorf("get label list : %s", err)
	}
	fmt.Println(labels)

	slabels, _, _ := us.repo.GetLabelsByQuery(233058313, "keyword:")
	fmt.Println(slabels)

	usStats := &Stats{
		NumUpdatedIssues:     len(upi),
		NumOpenIssues:        len(opi),
		NumCommentsPerIssue:  make(map[string]int, len(upi)),
		OpenDurationPerIssue: make(map[string]time.Duration, len(upi)),
	}
	for _, issue := range opi {
		if labelContains(issue.Labels, "緊急度:高") || labelContains(issue.Labels, "緊急度:中") {
			usStats.UrgencyHighIssues++
			if labelContains(issue.Labels, "難易度:高") {
				usStats.UrgencyHighDifficultyHighIssues++
			}
			if labelContains(issue.Labels, "難易度:低") {
				usStats.UrgencyHighDifficultyLowIssues++
			}
		}
		if labelContains(issue.Labels, "緊急度:低") {
			usStats.UrgencyLowIssues++
			if labelContains(issue.Labels, "難易度:高") {
				usStats.UrgencyLowDifficultyHighIssues++
			}
			if labelContains(issue.Labels, "難易度:低") {
				usStats.UrgencyLowDifficultyLowIssues++
			}
		}
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
	sb.WriteString(fmt.Sprintf("Openチケット数, %d\n", s.NumOpenIssues))
	sb.WriteString(fmt.Sprintf("新規作成されたチケット数, %d\n", s.NumCreatedIssues))
	sb.WriteString(fmt.Sprintf("情報更新されたチケット数, %d\n", s.NumUpdatedIssues))
	sb.WriteString(fmt.Sprintf("クローズしたチケット数, %d\n", s.NumClosedIssues))
	sb.WriteString(fmt.Sprintf("緊急度:高 のチケット数, %d\n", s.UrgencyHighIssues))
	sb.WriteString(fmt.Sprintf("かつ、難易度:高 のチケット数, %d\n", s.UrgencyHighDifficultyHighIssues))
	sb.WriteString(fmt.Sprintf("かつ、難易度:低 のチケット数, %d\n", s.UrgencyHighDifficultyLowIssues))
	sb.WriteString(fmt.Sprintf("緊急度:低 のチケット数, %d\n", s.UrgencyLowIssues))
	sb.WriteString(fmt.Sprintf("かつ、難易度:高 のチケット数, %d\n", s.UrgencyLowDifficultyHighIssues))
	sb.WriteString(fmt.Sprintf("かつ、難易度:低 のチケット数, %d\n", s.UrgencyLowDifficultyLowIssues))
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
		for _, kv := range kvArr {
			// if >=10 {
			// 	break
			// }
			totalHours := int(kv.Val.Hours())
			dates := totalHours / 24
			hours := totalHours % 24
			sb.WriteString(fmt.Sprintf("%s, %dd %dh\n", kv.Key, dates, hours))
		}
	}

	return sb.String()
}

//配列の中に特定の文字列が含まれるかを返す
func labelContains(arr []github.Label, str string) bool {
	for _, v := range arr {
		if *v.Name == str {
			return true
		}
	}
	return false
}
