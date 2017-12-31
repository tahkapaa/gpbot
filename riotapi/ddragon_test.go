package riotapi

import (
	"testing"
)

func TestDDragonClient_GetRunesReforged(t *testing.T) {
	c := NewDDragonClient()

	expectedVersion := "7.24.2"
	if c.version != expectedVersion {
		t.Errorf("unexpected version number: %v, expected: %v", c.version, expectedVersion)
	}
	_, err := c.GetRunesReforged()
	if err != nil {
		t.Errorf("failed to fetch runes reforged data")
	}
}

func Test_beautifyString(t *testing.T) {
	tests := []struct {
		name string
		s    string
		want string
	}{
		{name: "Basic case", s: "string with <> here", want: "string with  here"},
		{name: "Test to remove end part", s: "string with <>", want: "string with "},
		{name: "Test to remove from start", s: "<string> with ", want: " with "},
		{name: "Test to remove multiple", s: "<string> with <> and also some <other>", want: " with  and also some "},
		{name: "Test to remove with parts that should not be removed", s: "<string> with <> and also >some <other>", want: " with  and also >some "},
		{name: "Test remove <<multiple>>", s: "string <<with>> multiple ", want: "string  multiple "},

		{name: "Basic case", s: "string with @@ here", want: "string with  here"},
		{name: "Test to remove end part", s: "string with @@", want: "string with "},
		{name: "Test to remove from start", s: "@string@ with ", want: " with "},
		{name: "Test to remove multiple", s: "@string@ with @@ and also some @other@", want: " with  and also some "},
		{name: "Test to remove with parts that should not be removed", s: "@string@ with @@ and also >some @other@", want: " with  and also >some "},
		{name: "Remove all", s: "{{ perk_short_desc_NemFighter }}", want: ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := beautifyString(tt.s); got != tt.want {
				t.Errorf("beautifyString() = '%v', want '%v'", got, tt.want)
			}
		})
	}
}

func TestDDragonClient_GetProfileIconURL(t *testing.T) {
	c := NewDDragonClient()
	URL := c.GetProfileIconURL(0)
	expectedURL := "https://ddragon.leagueoflegends.com/cdn/7.24.2/img/profileicon/0.png"
	if URL != expectedURL {
		t.Errorf("invalid url received: %v, expected: %v", URL, expectedURL)
	}
}
