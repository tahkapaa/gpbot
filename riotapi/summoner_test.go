package riotapi

import (
	"testing"
)

func TestSummonerAPI_SummonerByName(t *testing.T) {
	api := SummonerAPI{newClient(t)}
	type args struct {
		name string
	}
	tests := []struct {
		name    string
		args    args
		want    *SummonerDTO
		wantErr bool
	}{
		{
			name:    "Get summoner by name",
			args:    args{name: "uxipaxa"},
			want:    &SummonerDTO{Name: "Uxipaxa", ID: 24749077, AccountID: 29325268},
			wantErr: false,
		},
		{
			name:    "Get nonexistant summoner by name",
			args:    args{name: "non_existing_name_512035kmsadfu815ij"},
			want:    nil,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := api.SummonerByName(tt.args.name)
			if (err != nil) != tt.wantErr {
				t.Errorf("SummonerAPI.SummonerByName() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.want != nil && (got.ID != tt.want.ID || got.Name != tt.want.Name || got.AccountID != tt.want.AccountID) {
				t.Errorf("SummonerAPI.SummonerByName() = %v, want %v", got, tt.want)
			}

			if tt.want == nil && got != nil {
				t.Errorf("SummonerAPI.SummonerByName() = %v, want %v", got, tt.want)
			}
		})
	}
}
