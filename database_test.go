package main

import (
	"reflect"
	"testing"
)

func TestFireBaseDB(t *testing.T) {
	fb := New("test")
	testData := ChannelData{
		ID: "test channel",
		Summoners: map[string]Player{
			"123456789": Player{
				Name:          "Summoner",
				ID:            123456789,
				CurrentGameID: 0,
				Rank:          "Rank",
			},
			"1234567891": Player{
				Name:          "Summoner2",
				ID:            1234567891,
				CurrentGameID: 0,
				Rank:          "Rank2",
			},
			"1234567892": Player{
				Name:          "Summoner3",
				ID:            1234567892,
				CurrentGameID: 0,
				Rank:          "Rank3",
			},
			"1234567893": Player{
				Name:          "Summoner4",
				ID:            1234567893,
				CurrentGameID: 0,
				Rank:          "Rank4",
			},
		},
	}

	if err := fb.Save(&testData); err != nil {
		t.Fatalf("Unable to save channel data: %v", err)
	}

	tests := []struct {
		name    string
		want    []*ChannelData
		wantErr bool
	}{
		{name: "test that all data can be fetched", want: []*ChannelData{&testData}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := fb.Get()
			if (err != nil) != tt.wantErr {
				t.Errorf("FireBaseDB.Get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FireBaseDB.Get() = %v, want %v", got, tt.want)
			}
		})
	}
}
