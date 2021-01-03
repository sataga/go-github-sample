package usersupport

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/google/go-github/github"
)

var (
	jp = time.FixedZone("Asia/Tokyo", 9*60*60)
)

// UserSupport is interface for getting user support info
type UserSupport interface {
	GetDailyReportStats(until time.Time) (*DailyStats, error)
	GetMonthlyReportStats(since, until time.Time) (*MonthlyStats, error)
	GetAnalysisReportStats(since, until time.Time) (*AnalysisStats, error)
	MethodTest(since, until time.Time) (*AnalysisStats, error)
	// GenMonthlyReport(data map[string]*MonthlyStats) string
}

// Repository r/w data which usersupport domain requires
type Repository interface {
	GetUpdatedSupportIssues(since, until time.Time) ([]*github.Issue, error)
	GetClosedSupportIssues(since, until time.Time) ([]*github.Issue, error)
	GetCurrentOpenNotUpdatedSupportIssues(until time.Time) ([]*github.Issue, error)
	GetCurrentOpenSupportIssues() ([]*github.Issue, error)
	GetCurrentOpenAnyLabelsSupportIssues(state string, labels []string) ([]*github.Issue, error)
	GetCurrentRepoLabels() ([]*github.Label, error)
	GetLabelsByQuery(repoID int64, query string) (*github.LabelsSearchResult, *github.Response, error)
}

type userSupport struct {
	repo Repository
}

// DailyStats is stats open data from GitHub
type DailyStats struct {
	NumNotUpdatedIssues int                  `yaml:"num_not_updated_issues"`
	NumTeamAResponse    int                  `yaml:"num_team_a_response"`
	NumTeamBResponse    int                  `yaml:"num_team_b_response"`
	UrgencyHighIssues   int                  `yaml:"num_urgency_high_issues"`
	UrgencyLowIssues    int                  `yaml:"num_urgency_low_issues"`
	DetailStats         map[int]*DetailStats `yaml:"detail_stats"`
}
type MonthlyStats struct {
	SummaryStats map[string]*SummaryStats `yaml:"summary_stats"`
	DetailStats  map[int]*DetailStats     `yaml:"detail_stats"`
}

type AnalysisStats struct {
	DetailStats map[int]*DetailStats `yaml:"detail_stats"`
}
type SummaryStats struct {
	Span                         string `yajl:"span"`
	NumCreatedIssues             int    `yaml:"num_created_issues"`
	NumClosedIssues              int    `yaml:"num_closed_issues"`
	NumGenreRequestIssues        int    `yaml:"num_genre_issues"`
	NumGenreLogSurveyIssues      int    `yaml:"num_genre_log_survey_issues"`
	NumGenreImpactSurveyIssues   int    `yaml:"num_genre_impact_survey_issues"`
	NumGenreSpecSurveyIssues     int    `yaml:"num_genre_spec_survey_issues"`
	NumGenreIncidentSurveyIssues int    `yaml:"num_genre_incident_survey_issues"`
	NumTeamAResolveIssues        int    `yaml:"num_team_a_resolve_issues"`
	NumUrgencyHighIssues         int    `yaml:"num_urgency_high_issues"`
	NumUrgencyLowIssues          int    `yaml:"num_urgency_low_issues"`
	NumScoreA                    int    `yaml:"num_score_A"`
	NumScoreB                    int    `yaml:"num_score_B"`
	NumScoreC                    int    `yaml:"num_score_C"`
	NumScoreD                    int    `yaml:"num_score_D"`
	NumScoreE                    int    `yaml:"num_score_E"`
	NumScoreF                    int    `yaml:"num_score_F"`
}

type DetailStats struct {
	Title        string `yaml:"detail_stats_of_title"`
	HTMLURL      string `yaml:"detail_stats_of_issue_url"`
	CreatedAt    string `yaml:"detail_stats_of_created_at"`
	ClosedAt     string `yaml:"detail_stats_of_closed_at"`
	State        string `yaml:"detail_stats_of_state"`
	TargetSpan   string `yaml:"detail_stats_of_target_span"`
	TeamName     string `yaml:"detail_stats_of_team_name"`
	Urgency      string `yaml:"detail_stats_of_urgency"`
	TeamAResolve bool   `yaml:"detail_stats_of_team_a_resolve"`
	Genre        string `yaml:"detail_stats_of_genre"`
	Labels       string `yaml:"detail_stats_of_labels"`
	Assignee     string `yaml:"detail_stats_of_assign"`
	NumComments  int    `yaml:"detail_stats_of_num_comment"`
	OpenDuration int    `yaml:"detail_stats_of_open_duration"`
}

// NewUserSupport creates UserSupport
func NewUserSupport(repo Repository) UserSupport {
	return &userSupport{
		repo: repo,
	}
}

// GetDailryReport
func (us *userSupport) GetDailyReportStats(until time.Time) (*DailyStats, error) {
	startEnd := fmt.Sprintf("%s", until.Format("2006-01-02"))
	opi, err := us.repo.GetCurrentOpenNotUpdatedSupportIssues(until)
	if err != nil {
		return nil, fmt.Errorf("get open issues : %s", err)
	}
	DailyStats := &DailyStats{
		NumNotUpdatedIssues: len(opi),
		DetailStats:         make(map[int]*DetailStats, len(opi)),
	}
	for i, issue := range opi {
		DailyStats.DetailStats[i] = &DetailStats{}
		if labelContains(issue.Labels, "緊急度:高") {
			DailyStats.UrgencyHighIssues++
			DailyStats.DetailStats[i].Urgency = "高"
		}
		if labelContains(issue.Labels, "緊急度:中") {
			DailyStats.UrgencyHighIssues++
			DailyStats.DetailStats[i].Urgency = "中"
		}
		if labelContains(issue.Labels, "緊急度:低") {
			DailyStats.UrgencyLowIssues++
			DailyStats.DetailStats[i].Urgency = "低"
		}
		if labelContains(issue.Labels, "Team-A") {
			DailyStats.NumTeamAResponse++
			DailyStats.DetailStats[i].TeamName = "Team-A"
		}
		if labelContains(issue.Labels, "Team-B") {
			DailyStats.NumTeamBResponse++
			DailyStats.DetailStats[i].TeamName = "Team-B"
		}
		DailyStats.DetailStats[i].writeDetailStats(issue, startEnd)
	}
	return DailyStats, nil
}

func (s *DailyStats) GetDailyReportStats() string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("=== サマリー ===\n"))
	sb.WriteString(fmt.Sprintf("総未更新チケット数, %d\n", s.NumNotUpdatedIssues))
	sb.WriteString(fmt.Sprintf("Team-A 未更新チケット数, %d\n", s.NumTeamAResponse))
	sb.WriteString(fmt.Sprintf("Team-B 未更新チケット数, %d\n", s.NumTeamBResponse))
	sb.WriteString(fmt.Sprintf("=== 詳細 ===\n"))
	for _, issue := range s.DetailStats {
		dates := issue.OpenDuration / 24
		hours := issue.OpenDuration % 24
		sb.WriteString(fmt.Sprintf("- <%s|%s> ", issue.HTMLURL, issue.Title))
		sb.WriteString(fmt.Sprintf("%dd%dh ", dates, hours))
		sb.WriteString(fmt.Sprintf("%s/", issue.TeamName))
		sb.WriteString(fmt.Sprintf("%s\n", issue.Assignee))
	}
	return sb.String()
}

// GetMonthlyReport
func (us *userSupport) GetMonthlyReportStats(since, until time.Time) (*MonthlyStats, error) {
	//
	span := 4
	MonthlyStats := &MonthlyStats{
		SummaryStats: make(map[string]*SummaryStats, span),
		DetailStats:  make(map[int]*DetailStats, span),
	}
	cnt := 0
	for i := 1; i <= span; i++ {
		startEnd := fmt.Sprintf("%s~%s", since.Format("2006-01-02"), until.Format("2006-01-02"))
		upi, err := us.repo.GetUpdatedSupportIssues(since, until)
		if err != nil {
			return nil, fmt.Errorf("get open issues : %s", err)
		}
		numCreated, numClosed := 0, 0
		for _, issue := range upi {
			if issue.State != nil && *issue.State == "closed" {
				numClosed++
			}
			if issue.CreatedAt != nil && issue.CreatedAt.After(since) && issue.CreatedAt.Before(until) {
				numCreated++
			}
		}

		cli, err := us.repo.GetClosedSupportIssues(since, until)
		if err != nil {
			return nil, fmt.Errorf("get updated issues : %s", err)
		}
		MonthlyStats.SummaryStats[startEnd] = &SummaryStats{
			Span:             startEnd,
			NumClosedIssues:  numClosed,
			NumCreatedIssues: numCreated,
		}
		for _, issue := range cli {
			MonthlyStats.DetailStats[cnt] = &DetailStats{
				TeamAResolve: false,
			}
			if labelContains(issue.Labels, "genre:影響調査") {
				MonthlyStats.SummaryStats[startEnd].NumGenreImpactSurveyIssues++
			}
			if labelContains(issue.Labels, "genre:要望") {
				MonthlyStats.SummaryStats[startEnd].NumGenreRequestIssues++
			}
			if labelContains(issue.Labels, "genre:ログ調査") {
				MonthlyStats.SummaryStats[startEnd].NumGenreLogSurveyIssues++
			}
			if labelContains(issue.Labels, "genre:仕様調査") {
				MonthlyStats.SummaryStats[startEnd].NumGenreSpecSurveyIssues++
			}
			if labelContains(issue.Labels, "genre:障害調査") {
				MonthlyStats.SummaryStats[startEnd].NumGenreIncidentSurveyIssues++
			}
			if labelContains(issue.Labels, "TeamA単体解決") {
				MonthlyStats.SummaryStats[startEnd].NumTeamAResolveIssues++
			}
			if labelContains(issue.Labels, "緊急度:高") || labelContains(issue.Labels, "緊急度:中") {
				MonthlyStats.SummaryStats[startEnd].NumUrgencyHighIssues++
			}
			if labelContains(issue.Labels, "緊急度:低") {
				MonthlyStats.SummaryStats[startEnd].NumUrgencyLowIssues++
			}

			var totalTime int
			if issue.State != nil && *issue.State == "closed" {
				totalTime = int(issue.ClosedAt.Sub(*issue.CreatedAt).Hours())
			} else {
				totalTime = int(issue.UpdatedAt.Sub(*issue.CreatedAt).Hours())
			}
			switch {
			case totalTime <= 2:
				MonthlyStats.SummaryStats[startEnd].NumScoreA++
			case totalTime <= 5:
				MonthlyStats.SummaryStats[startEnd].NumScoreB++
			case totalTime <= 10:
				MonthlyStats.SummaryStats[startEnd].NumScoreC++
			case totalTime <= 20:
				MonthlyStats.SummaryStats[startEnd].NumScoreD++
			case totalTime <= 30:
				MonthlyStats.SummaryStats[startEnd].NumScoreE++
			default:
				MonthlyStats.SummaryStats[startEnd].NumScoreF++
			}

			MonthlyStats.DetailStats[cnt].writeDetailStats(issue, startEnd)
			cnt++
		}
		MonthlyStats.SummaryStats[startEnd].NumClosedIssues = len(cli)
		since = since.AddDate(0, 0, -7)
		until = until.AddDate(0, 0, -7)

	}
	return MonthlyStats, nil
}

func (ms *MonthlyStats) GenMonthlyReport() string {
	var sb strings.Builder
	var Span []string
	var NumCreatedIssues []string
	var NumClosedIssues []string
	var NumGenreRequestIssues []string
	var NumGenreLogSurveyIssues []string
	var NumGenreImpactSurveyIssues []string
	var NumGenreSpecSurveyIssues []string
	var NumGenreIncidentSurveyIssues []string
	var NumTeamAResolveIssues []string
	var NumTeamAResolvePercentage []string
	var NumScoreA []string
	var NumScoreB []string
	var NumScoreC []string
	var NumScoreD []string
	var NumScoreE []string
	var NumScoreF []string

	type kv struct {
		Key string
		Val *SummaryStats
	}
	var kvArr []kv
	for k, v := range ms.SummaryStats {
		kvArr = append(kvArr, kv{k, v})
	}
	// sort by Span
	sort.Slice(kvArr, func(i, j int) bool {
		return kvArr[i].Val.Span < kvArr[j].Val.Span
	})

	for _, d := range kvArr {
		Span = append(Span, d.Val.Span)
		NumCreatedIssues = append(NumCreatedIssues, strconv.Itoa(d.Val.NumCreatedIssues))
		NumClosedIssues = append(NumClosedIssues, strconv.Itoa(d.Val.NumClosedIssues))
		NumGenreRequestIssues = append(NumGenreRequestIssues, strconv.Itoa(d.Val.NumGenreRequestIssues))
		NumGenreLogSurveyIssues = append(NumGenreLogSurveyIssues, strconv.Itoa(d.Val.NumGenreLogSurveyIssues))
		NumGenreImpactSurveyIssues = append(NumGenreImpactSurveyIssues, strconv.Itoa(d.Val.NumGenreLogSurveyIssues))
		NumGenreSpecSurveyIssues = append(NumGenreSpecSurveyIssues, strconv.Itoa(d.Val.NumGenreSpecSurveyIssues))
		NumGenreIncidentSurveyIssues = append(NumGenreIncidentSurveyIssues, strconv.Itoa(d.Val.NumGenreIncidentSurveyIssues))
		NumTeamAResolveIssues = append(NumTeamAResolveIssues, strconv.Itoa(d.Val.NumTeamAResolveIssues))
		if d.Val.NumTeamAResolveIssues != 0 {
			if d.Val.NumClosedIssues != 0 {
				NumTeamAResolvePercentage = append(NumTeamAResolvePercentage, fmt.Sprintf("%.1f", (float64(d.Val.NumTeamAResolveIssues)/float64(d.Val.NumClosedIssues)*100)))
			} else {
				NumTeamAResolvePercentage = append(NumTeamAResolvePercentage, "0")
			}
		} else {
			NumTeamAResolvePercentage = append(NumTeamAResolvePercentage, "0")
		}

		NumScoreA = append(NumScoreA, strconv.Itoa(d.Val.NumScoreA))
		NumScoreB = append(NumScoreB, strconv.Itoa(d.Val.NumScoreB))
		NumScoreC = append(NumScoreC, strconv.Itoa(d.Val.NumScoreC))
		NumScoreD = append(NumScoreD, strconv.Itoa(d.Val.NumScoreD))
		NumScoreE = append(NumScoreE, strconv.Itoa(d.Val.NumScoreE))
		NumScoreF = append(NumScoreF, strconv.Itoa(d.Val.NumScoreF))
	}
	sb.WriteString(fmt.Sprintf("## サマリー \n"))
	sb.WriteString(fmt.Sprintf("|項目|"))
	sb.WriteString(fmt.Sprintf("%s|\n", strings.Join(Span, "|")))
	sb.WriteString(fmt.Sprintf("|----|"))
	for i := 0; i < len(kvArr); i++ {
		sb.WriteString(fmt.Sprintf("----|"))
	}
	sb.WriteString(fmt.Sprintf("\n"))
	sb.WriteString(fmt.Sprintf("|起票件数|%s|\n", strings.Join(NumCreatedIssues, "|")))
	sb.WriteString(fmt.Sprintf("|Team-A完結件数|%s|\n", strings.Join(NumTeamAResolveIssues, "|")))
	sb.WriteString(fmt.Sprintf("|クローズ件数|%s|\n", strings.Join(NumClosedIssues, "|")))
	sb.WriteString(fmt.Sprintf("|Team-A完結率(％)|%s|\n", strings.Join(NumTeamAResolvePercentage, "|")))
	sb.WriteString(fmt.Sprintf("|ジャンル:要望件数|%s|\n", strings.Join(NumGenreRequestIssues, "|")))
	sb.WriteString(fmt.Sprintf("|ジャンル:ログ調査件数|%s|\n", strings.Join(NumGenreLogSurveyIssues, "|")))
	sb.WriteString(fmt.Sprintf("|ジャンル:影響調査件数|%s|\n", strings.Join(NumGenreImpactSurveyIssues, "|")))
	sb.WriteString(fmt.Sprintf("|ジャンル:仕様調査件数|%s|\n", strings.Join(NumGenreSpecSurveyIssues, "|")))
	sb.WriteString(fmt.Sprintf("|ジャンル:障害調査件数|%s|\n", strings.Join(NumGenreIncidentSurveyIssues, "|")))
	sb.WriteString(fmt.Sprintf("|スコアA|%s|\n", strings.Join(NumScoreA, "|")))
	sb.WriteString(fmt.Sprintf("|スコアB|%s|\n", strings.Join(NumScoreB, "|")))
	sb.WriteString(fmt.Sprintf("|スコアC|%s|\n", strings.Join(NumScoreC, "|")))
	sb.WriteString(fmt.Sprintf("|スコアD|%s|\n", strings.Join(NumScoreD, "|")))
	sb.WriteString(fmt.Sprintf("|スコアE|%s|\n", strings.Join(NumScoreE, "|")))
	sb.WriteString(fmt.Sprintf("|スコアF|%s|\n", strings.Join(NumScoreF, "|")))
	sb.WriteString(fmt.Sprintf("\n"))

	sb.WriteString(fmt.Sprintf("## 詳細 \n"))
	for _, d := range ms.DetailStats {
		dates := d.OpenDuration / 24
		hours := d.OpenDuration % 24
		sb.WriteString(fmt.Sprintf("- [%s](%s),%s,%s,%s/%s,comment数:%d,経過時間:%dd%dh,解決フラグ:%t,span:	%s\n", d.Title, d.HTMLURL, d.Urgency, d.Genre, d.TeamName, d.Assignee, d.NumComments, dates, hours, d.TeamAResolve, d.TargetSpan))
	}
	return sb.String()
}

func (us *userSupport) GetAnalysisReportStats(since, until time.Time) (*AnalysisStats, error) {
	span := 4
	loc, _ := time.LoadLocation("Asia/Tokyo")
	since = time.Date(since.Year(), since.Month(), 1, 0, 0, 0, 0, loc)

	AnalysisStats := &AnalysisStats{
		DetailStats: make(map[int]*DetailStats, span),
	}

	cnt := 0
	for i := 1; i <= span; i++ {
		startEnd := fmt.Sprintf("%s~%s", since.Format("2006-01-02"), until.Format("2006-01-02"))
		upi, err := us.repo.GetUpdatedSupportIssues(since, until)
		if err != nil {
			return nil, fmt.Errorf("get open issues : %s", err)
		}
		for _, issue := range upi {
			AnalysisStats.DetailStats[cnt] = &DetailStats{
				TeamAResolve: false,
			}
			AnalysisStats.DetailStats[cnt].writeDetailStats(issue, startEnd)
			cnt++
		}
		since = time.Date(since.Year(), since.Month()-1, 1, 0, 0, 0, 0, loc)
		until = since.AddDate(0, +1, -1)
	}

	return AnalysisStats, nil
}

// GenReport generate analysis report
func (as *AnalysisStats) GenAnalysisReport() string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("期間,Title,起票日,クローズ日,ステータス,担当チーム,担当アサイン,緊急度,問い合わせ種別,コメント数,経過時間,Keywordラベル,URL\n"))
	for _, d := range as.DetailStats {
		sb.WriteString(fmt.Sprintf("%s,%s,%s,%s,%s,%s,%s,%s,%s,%d,%d,%s,%s \n", d.TargetSpan, d.Title, d.CreatedAt, d.ClosedAt, d.State, d.TeamName, d.Assignee, d.Urgency, d.Genre, d.NumComments, d.OpenDuration, d.Labels, d.HTMLURL))
	}
	return sb.String()
}

func (us *userSupport) MethodTest(since, until time.Time) (*AnalysisStats, error) {
	startEnd := fmt.Sprintf("%s~%s", since.Format("2006-01-02"), until.Format("2006-01-02"))
	cli, err := us.repo.GetUpdatedSupportIssues(since, until)
	if err != nil {
		return nil, fmt.Errorf("get updated issues : %s", err)
	}
	AnalysisStats := &AnalysisStats{
		DetailStats: make(map[int]*DetailStats, len(cli)),
	}
	for i, issue := range cli {
		AnalysisStats.DetailStats[i] = &DetailStats{
			TeamAResolve: false,
		}
		AnalysisStats.DetailStats[i].writeDetailStats(issue, startEnd)
	}
	return AnalysisStats, nil
}

func (ds *DetailStats) writeDetailStats(issue *github.Issue, startEnd string) {
	var labels []string
	if issue.Labels != nil {
		for _, label := range issue.Labels {
			if strings.Contains(*label.Name, "keyword") {
				labels = append(labels, strings.Replace(*label.Name, "keyword:", "", -1))
			}
			if strings.Contains(*label.Name, "緊急度") {
				ds.Urgency = strings.Replace(*label.Name, "緊急度:", "", -1)
			}
			if strings.Contains(*label.Name, "Team-") {
				ds.TeamName = *label.Name
			}
			if strings.Contains(*label.Name, "単体解決") {
				ds.TeamAResolve = true
			}
			if strings.Contains(*label.Name, "genre") {
				ds.Genre = strings.Replace(*label.Name, "genre:", "", -1)
			}
		}
	}
	var assigns []string
	if issue.Assignees != nil {
		for _, assign := range issue.Assignees {
			assigns = append(assigns, *assign.Login)
		}
	}

	var totalTime int
	if issue.State != nil && *issue.State == "closed" {
		totalTime = int(issue.ClosedAt.Sub(*issue.CreatedAt).Hours())
		ds.ClosedAt = issue.ClosedAt.In(jp).Format("2006-01-02")
	} else {
		totalTime = int(issue.UpdatedAt.Sub(*issue.CreatedAt).Hours())
	}

	ds.Assignee = strings.Join(assigns, " ")
	ds.Title = *issue.Title
	ds.HTMLURL = *issue.HTMLURL
	ds.NumComments = *issue.Comments
	ds.State = *issue.State
	ds.CreatedAt = issue.CreatedAt.In(jp).Format("2006-01-02")
	ds.OpenDuration = totalTime
	ds.Assignee = strings.Join(assigns, " ")
	ds.Labels = strings.Join(labels, " ")
	ds.TargetSpan = startEnd
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
