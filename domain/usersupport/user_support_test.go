package usersupport

import (
	"reflect"
	"testing"
	"time"
)

func Test_userSupport_GetDailyReportStats(t *testing.T) {
	type fields struct {
		repo Repository
	}
	type args struct {
		until time.Time
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *dailyStats
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			us := &userSupport{
				repo: tt.fields.repo,
			}
			got, err := us.GetDailyReportStats(tt.args.until)
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
