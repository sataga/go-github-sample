// Package usersupport is a domain logic for user support
//go:generate mockgen -source=$GOFILE -destination=mock_$GOFILE -package=$GOPACKAGE

package usersupport

import (
	"reflect"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/google/go-github/github"
)

var (
	now         = time.Now()
	oneHourAgo  = now.Add(-1 * time.Hour)
	threeDayAgo = now.Add(-3 * 24 * time.Hour)
	fiveDayAgo  = now.Add(-5 * 24 * time.Hour)
	sevenDayAgo = now.Add(-7 * 24 * time.Hour)
	tenDayAgo   = now.Add(-10 * 24 * time.Hour)

	issuePatterns = []*github.Issue{
		{
			ID:        github.Int64(1),
			Title:     github.String("issue 1"),
			CreatedAt: &tenDayAgo,
			ClosedAt:  &threeDayAgo,
			State:     github.String("closed"),
			Body:      github.String("test 1"),
			Comments:  github.Int(1),
			Labels: []github.Label{
				{Name: github.String("PF_Support")},
				{Name: github.String("緊急度：低")},
				{Name: github.String("CaaS-A 対応中")},
			},
			HTMLURL: github.String("https://github.com/sataga/issue-warehouse/issues/1"),
		},
		{
			ID:        github.Int64(2),
			Title:     github.String("issue 2"),
			CreatedAt: &sevenDayAgo,
			ClosedAt:  &threeDayAgo,
			State:     github.String("closed"),
			Body: github.String(`
				test2
				hogehoge
			`),
			Comments: github.Int(2),
			Labels: []github.Label{
				{Name: github.String("PF_Support")},
				{Name: github.String("緊急度：中")},
				{Name: github.String("CaaS-A 対応中")},
			},
			HTMLURL: github.String("https://github.com/sataga/issue-warehouse/issues/2"),
		},
		{
			ID:        github.Int64(3),
			Title:     github.String("issue 3"),
			CreatedAt: &fiveDayAgo,
			UpdatedAt: &oneHourAgo,
			State:     github.String("open"),
			Body:      github.String("test 3"),
			Comments:  github.Int(3),
			Labels: []github.Label{
				{Name: github.String("PF_Support")},
				{Name: github.String("緊急度：高")},
				{Name: github.String("CaaS-A 対応中")},
			},
			HTMLURL: github.String("https://github.com/sataga/issue-warehouse/issues/3"),
		},
		{
			ID:        github.Int64(4),
			Title:     github.String("issue 4"),
			CreatedAt: &threeDayAgo,
			UpdatedAt: &oneHourAgo,
			State:     github.String("open"),
			Body:      github.String("test 4"),
			Comments:  github.Int(4),
			Labels: []github.Label{
				{Name: github.String("PF_Support")},
				{Name: github.String("緊急度：低")},
				{Name: github.String("CaaS-B 対応中")},
			},
			HTMLURL: github.String("https://github.com/sataga/issue-warehouse/issues/4"),
		},
	}
)

func Test_userSupport_GetDailyReportStats(t *testing.T) {
	var c *gomock.Controller

	updatedIssues := []*github.Issue{
		issuePatterns[2],
		issuePatterns[3],
	}

	type fields struct {
		repo Repository
	}
	type args struct {
		dayAgo int
	}
	tests := []struct {
		name       string
		fields     fields
		args       args
		want       *DailyStats
		wantErr    bool
		beforeFunc func(*fields)
		afterFunc  func()
	}{
		// TODO: Add test cases.
		{
			name: "since < until",
			args: args{
				dayAgo: 5,
			},
			want: &DailyStats{
				NumNotUpdatedIssues: 2,
				NumTeamAResponse:    1,
				NumTeamBResponse:    1,
				NumTeamAHighIssues:  0,
				NumTeamBLowIssues:   0,
				UrgencyHighIssues:   1,
				UrgencyLowIssues:    1,
				dayAgo:              5,
				DetailStats: map[int]*DetailStats{
					0: {
						Title:        "issue 3",
						HTMLURL:      "https://github.com/sataga/issue-warehouse/issues/3",
						CreatedAt:    fiveDayAgo.Format("2006-01-02"),
						State:        "open",
						TargetSpan:   fiveDayAgo.Format("2006-01-02"),
						TeamName:     "CaaS-A",
						Urgency:      "高",
						NumComments:  3,
						OpenDuration: 119,
						Escalation:   false,
					},
					1: {
						Title:        "issue 4",
						HTMLURL:      "https://github.com/sataga/issue-warehouse/issues/4",
						CreatedAt:    threeDayAgo.Format("2006-01-02"),
						State:        "open",
						TargetSpan:   fiveDayAgo.Format("2006-01-02"),
						TeamName:     "CaaS-B",
						Urgency:      "低",
						NumComments:  4,
						OpenDuration: 71,
						Escalation:   false,
					},
				},
			},
			wantErr: false,
			beforeFunc: func(f *fields) {
				c = gomock.NewController(t)
				musr := NewMockRepository(c)
				musr.EXPECT().GetCurrentOpenNotUpdatedSupportIssues(gomock.Any()).Return(updatedIssues, nil)
				f.repo = musr
			},
			afterFunc: func() {
				c.Finish()
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.beforeFunc != nil {
				tt.beforeFunc(&tt.fields)
			}
			if tt.afterFunc != nil {
				defer tt.afterFunc()
			}
			us := &userSupport{
				repo: tt.fields.repo,
			}
			got, err := us.GetDailyReportStats(now, tt.args.dayAgo)
			// fmt.Printf("got: %+v %+v\n ", got.DetailStats[0], got.DetailStats[1])
			// fmt.Printf("want: %+v %+v\n ", tt.want.DetailStats[0], tt.want.DetailStats[1])
			if (err != nil) != tt.wantErr {
				t.Errorf("userSupport.GetDailyReportStats() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("userSupport.GetDailyReportStats() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDailyStats_GetDailyReportStats(t *testing.T) {
	type fields struct {
		dayAgo              int
		NumNotUpdatedIssues int
		NumTeamAResponse    int
		NumTeamBResponse    int
		NumTeamAHighIssues  int
		NumTeamBLowIssues   int
		UrgencyHighIssues   int
		UrgencyLowIssues    int
		DetailStats         map[int]*DetailStats
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		// TODO: Add test cases.
		{
			name: "print daily-report",
			fields: fields{
				NumNotUpdatedIssues: 2,
				NumTeamAResponse:    1,
				NumTeamBResponse:    1,
				NumTeamAHighIssues:  0,
				NumTeamBLowIssues:   0,
				UrgencyHighIssues:   1,
				UrgencyLowIssues:    1,
				dayAgo:              5,
				DetailStats: map[int]*DetailStats{
					0: {
						Title:        "issue 3",
						HTMLURL:      "https://github.com/sataga/issue-warehouse/issues/3",
						CreatedAt:    fiveDayAgo.Format("2006-01-02"),
						State:        "open",
						TargetSpan:   fiveDayAgo.Format("2006-01-02"),
						TeamName:     "CaaS-A",
						Urgency:      "高",
						NumComments:  3,
						OpenDuration: 119,
						Escalation:   false,
					},
					1: {
						Title:        "issue 4",
						HTMLURL:      "https://github.com/sataga/issue-warehouse/issues/4",
						CreatedAt:    threeDayAgo.Format("2006-01-02"),
						State:        "open",
						TargetSpan:   fiveDayAgo.Format("2006-01-02"),
						TeamName:     "CaaS-B",
						Urgency:      "低",
						NumComments:  4,
						OpenDuration: 71,
						Escalation:   false,
					},
				},
			},
			want: `■ *5日間* 以上更新がなかったチケット一覧
=== サマリー ===
総未更新チケット数: 2 件
    緊急度：高・中: 1 件
    緊急度：低: 1 件
=== 詳細 ===
- <https://github.com/sataga/issue-warehouse/issues/3|issue 3> 経過時間:4d23h 緊急度：高 
- <https://github.com/sataga/issue-warehouse/issues/4|issue 4> 経過時間:2d23h 緊急度：低 
`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ds := &DailyStats{
				dayAgo:              tt.fields.dayAgo,
				NumNotUpdatedIssues: tt.fields.NumNotUpdatedIssues,
				NumTeamAResponse:    tt.fields.NumTeamAResponse,
				NumTeamBResponse:    tt.fields.NumTeamBResponse,
				NumTeamAHighIssues:  tt.fields.NumTeamAHighIssues,
				NumTeamBLowIssues:   tt.fields.NumTeamBLowIssues,
				UrgencyHighIssues:   tt.fields.UrgencyHighIssues,
				UrgencyLowIssues:    tt.fields.UrgencyLowIssues,
				DetailStats:         tt.fields.DetailStats,
			}
			// fmt.Printf("got: %+v\n ", ds.GetDailyReportStats())
			// fmt.Printf("want:%+v\n ", tt.want)
			if got := ds.GetDailyReportStats(); got != tt.want {
				t.Errorf("DailyStats.GetDailyReportStats() = %v, want %v", got, tt.want)
			}
		})
	}
}
