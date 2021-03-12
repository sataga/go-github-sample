package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	dus "github.com/sataga/go-github-sample/domain/usersupport"
	igh "github.com/sataga/go-github-sample/infra/github"
	"github.com/sataga/go-github-sample/infra/slack"
	ius "github.com/sataga/go-github-sample/infra/usersupport"
)

var (
	ghURL   = flag.String("ghurl", "https://api.github.com", "GitHub API base URL")
	ghUser  = flag.String("ghuser", "sataga", "Github user name")
	ghMail  = flag.String("ghmail", "", "Github user email")
	ghToken = flag.String("ghtoken", "", "GitHub Personal access token")

	cnt = 0

	loc, _          = time.LoadLocation("Asia/Tokyo")
	now             = time.Now()
	oneWeekBefore   = now.Add(-7 * 24 * time.Hour)
	userSupportFlag = flag.NewFlagSet("us", flag.ExitOnError)
	sinceStr        = userSupportFlag.String("since", oneWeekBefore.Format("2006-01-02"), "Date since listing issues from")
	untilStr        = userSupportFlag.String("until", now.Format("2006-01-02"), "Date until listing issues from")

	dailyReportFlag = flag.NewFlagSet("daily-report", flag.ExitOnError)
	dailyDayAgoInt  = dailyReportFlag.Int("day-ago", 7, "Please specify a date that has not been updated")

	longtermReportFlag = flag.NewFlagSet("longterm-report", flag.ExitOnError)
	longtermKindStr    = longtermReportFlag.String("kind", "monthly", "Please choose on (weekly , monthly)")
	longtermSpanInt    = longtermReportFlag.Int("span", 4, "Please enter the span you want to get")
	longtermOriginStr  = longtermReportFlag.String("origin", now.Format("2006-01-02"), "Get the data based on the date you entered")

	analysisReportFlag = flag.NewFlagSet("analysys-report", flag.ExitOnError)
	analysisSinceStr   = analysisReportFlag.String("since", oneWeekBefore.Format("2006-01-02"), "Date since listing issues from")
	analysisUntilStr   = analysisReportFlag.String("until", now.Format("2006-01-02"), "Date until listing issues from")
	analysisStateStr   = analysisReportFlag.String("state", "created", "Please choose on (created , closed)")
	analysisSpanInt    = analysisReportFlag.Int("span", 4, "Please enter the span you want to get")

	keywordReportFlag = flag.NewFlagSet("keyword-report", flag.ExitOnError)
	keywordKindStr    = keywordReportFlag.String("kind", "monthly", "Please choose on (weekly , monthly)")
	keywordSpanInt    = keywordReportFlag.Int("span", 4, "Please enter the span you want to get")
	keywordUntilStr   = keywordReportFlag.String("until", now.Format("2006-01-02"), "Date until listing issue from")
)

func printDefaultsAll() {
	fmt.Println("usage: go-github-sample [global options] subcommand [subcommand options]")
	flag.PrintDefaults()
	fmt.Println("\nsubcommands:")
	fmt.Println("us:    getting user support info")
	userSupportFlag.PrintDefaults()
	fmt.Println("daily-report:    Notify slack of tickets that have not been updated since the specified date")
	dailyReportFlag.PrintDefaults()
	fmt.Println("longterm-report:    Output user information in Markdown format based on kind")
	dailyReportFlag.PrintDefaults()
	fmt.Println("analysis-report:    Output user information in CSV format")
	analysisReportFlag.PrintDefaults()
}

func main() {
	jst, _ := time.LoadLocation("Asia/Tokyo")
	flag.Parse()
	if os.Getenv("GITHUB_TOKEN") != "" {
		*ghToken = os.Getenv("GITHUB_TOKEN")
	}
	if os.Getenv("GITHUB_MAIL") != "" {
		*ghMail = os.Getenv("GITHUB_MAIL")
	}
	ghcli, err := igh.NewGitHubClient(*ghURL, *ghToken, *ghUser, *ghMail)
	if err != nil {
		log.Fatalf("github client: %s", err)
	}
	subCommandArgs := os.Args[1+flag.NFlag():]
	if len(subCommandArgs) == 0 {
		printDefaultsAll()
		log.Fatalln("specify subcommand")
	}
	switch subCommand := subCommandArgs[0]; subCommand {
	case "daily-report":
		if err := dailyReportFlag.Parse(subCommandArgs[1:]); err != nil {
			log.Fatalf("parsing daily report flag: %s", err)
		}
		usrepo := ius.NewUserSupportRepository(ghcli)
		us := dus.NewUserSupport(usrepo)
		dairyStats, err := us.GetDailyReportStats(now, *dailyDayAgoInt)
		if err != nil {
			log.Fatalf("get user support stats: %s", err)
		}
		fmt.Printf("%s", dairyStats.GetDailyReportStats())
		// channel := "times_t-sataga"
		// username := "t-sataga"
		// _, err = slack.PostMessage(channel, username, dairyStats.GetDailyReportStats())
		// if err != nil {
		// 	log.Fatalf("slack post message failed: %s", err)
		// }
	case "longterm-report":
		if err := userSupportFlag.Parse(subCommandArgs[1:]); err != nil {
			log.Fatalf("parsing longterm report flag: %s", err)
		}
		usrepo := ius.NewUserSupportRepository(ghcli)
		us := dus.NewUserSupport(usrepo)
		var since, until time.Time
		var err error
		if until, err = time.ParseInLocation("2006-01-02", *longtermOriginStr, jst); err != nil {
			log.Fatalf("could not parse: %s", *longtermOriginStr)
		}
		switch *longtermKindStr {
		case "weekly":
			since = until.AddDate(0, 0, -7)
		case "monthly":
			since = time.Date(until.Year(), until.Month(), 1, 0, 0, 0, 0, loc)
			until = since.AddDate(0, +1, -1)
		}
		LongTermStats := &dus.LongTermStats{
			SummaryStats: make(map[string]*dus.SummaryStats, *longtermSpanInt),
			DetailStats:  make(map[int]*dus.DetailStats),
		}
		for i := 1; i <= *longtermSpanInt; i++ {
			result, err := us.GetLongTermReportStats(since, until)
			for key, val := range result.SummaryStats {
				LongTermStats.SummaryStats[key] = val
			}
			for _, val := range result.DetailStats {
				LongTermStats.DetailStats[cnt] = val
				cnt++
			}
			if err != nil {
				log.Fatalf("get longterm stats: %s", err)
			}
			switch *longtermKindStr {
			case "weekly":
				since = since.AddDate(0, 0, -7)
				until = until.AddDate(0, 0, -7)
			case "monthly":
				since = time.Date(since.Year(), since.Month()-1, 1, 0, 0, 0, 0, loc)
				until = since.AddDate(0, +1, -1)
			}
		}
		fmt.Printf("%s", LongTermStats.GenLongTermReport())
	case "analysis-report":
		if err := analysisReportFlag.Parse(subCommandArgs[1:]); err != nil {
			log.Fatalf("parsing analysis support flag: %s", err)
		}
		usrepo := ius.NewUserSupportRepository(ghcli)
		us := dus.NewUserSupport(usrepo)
		var since, until time.Time
		var err error
		if since, err = time.Parse("2006-01-02", *analysisSinceStr); err != nil {
			log.Fatalf("could not parse: %s", *analysisSinceStr)
		}
		if until, err = time.Parse("2006-01-02", *analysisUntilStr); err != nil {
			log.Fatalf("could not parse: %s", *analysisUntilStr)
		}
		since = time.Date(since.Year(), since.Month(), 1, 0, 0, 0, 0, loc)
		until = since.AddDate(0, +1, -1)
		AnalysisStats := &dus.AnalysisStats{
			DetailStats: make(map[int]*dus.DetailStats),
		}
		for i := 1; i <= *analysisSpanInt; i++ {
			result, err := us.GetAnalysisReportStats(since, until, *analysisStateStr)
			if err != nil {
				log.Fatalf("get user support stats: %s", err)
			}
			for _, val := range result.DetailStats {
				AnalysisStats.DetailStats[cnt] = val
				cnt++
			}
			since = time.Date(since.Year(), since.Month()-1, 1, 0, 0, 0, 0, loc)
			until = since.AddDate(0, +1, -1)
		}
		// fmt.Printf("Reporting Stats From: %s, Until: %s\n", since, until)
		fmt.Printf("%s", AnalysisStats.GenAnalysisReport())
	case "keyword-report":
		if err := keywordReportFlag.Parse(subCommandArgs[1:]); err != nil {
			log.Fatalf("parsing keyword report flag: %s", err)
		}
		usrepo := ius.NewUserSupportRepository(ghcli)
		us := dus.NewUserSupport(usrepo)
		var since, until time.Time
		var err error
		if until, err = time.ParseInLocation("2006-01-02", *keywordUntilStr, jst); err != nil {
			log.Fatalf("could not parse: %s", err)
		}
		switch *keywordKindStr {
		case "weekly":
			since = until.AddDate(0, 0, -7)
		case "monthly":
			since = time.Date(until.Year(), until.Month(), 1, 0, 0, 0, 0, loc)
			until = since.AddDate(0, +1, -1)
		}
		KeywordStats := &dus.KeywordStats{
			KeywordSummary: make(map[string]*dus.KeywordSummary),
		}
		for i := 1; i <= *longtermSpanInt; i++ {
			result, err := us.GetKeywordReportStats(since, until)
			if err != nil {
				log.Fatalf("get keyword stats: %s", err)
			}
			for key, val := range result.KeywordSummary {
				KeywordStats.KeywordSummary[key] = val
			}
			if err != nil {
				log.Fatalf("get longterm stats: %s", err)
			}
			switch *longtermKindStr {
			case "weekly":
				since = since.AddDate(0, 0, -7)
				until = until.AddDate(0, 0, -7)
			case "monthly":
				since = time.Date(since.Year(), since.Month()-1, 1, 0, 0, 0, 0, loc)
				until = since.AddDate(0, +1, -1)
			}
		}
		fmt.Printf("%s", KeywordStats.GenKeywordReport())

	case "slacktest":
		channel := "times_t-sataga"
		username := "t-sataga"
		text := "hogehoge"
		_, err := slack.PostMessage(channel, username, text)
		if err != nil {
			log.Fatalf("slack post message failed: %s", err)
		}
	case "methodtest":
		if err := userSupportFlag.Parse(subCommandArgs[1:]); err != nil {
			log.Fatalf("parsing user support flag: %s", err)
		}
		usrepo := ius.NewUserSupportRepository(ghcli)
		us := dus.NewUserSupport(usrepo)
		var since, until time.Time
		var err error
		if since, err = time.Parse("2006-01-02", *sinceStr); err != nil {
			log.Fatalf("could not parse: %s", *sinceStr)
		}
		if until, err = time.Parse("2006-01-02", *untilStr); err != nil {
			log.Fatalf("could not parse: %s", *untilStr)
		}
		fmt.Printf("Reporting Stats From: %s, Until: %s\n", since, until)
		testStats, err := us.MethodTest(since, until)
		if err != nil {
			log.Fatalf("get user support stats: %s", err)
		}
		fmt.Printf("%v", testStats)
	case "get-pj-board":
		usrepo := ius.NewUserSupportRepository(ghcli)
		us := dus.NewUserSupport(usrepo)
		result := us.GetProjectBoard()
		fmt.Printf("%s", result)
	}

}
