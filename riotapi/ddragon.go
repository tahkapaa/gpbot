package riotapi

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// DDragonClient fetches data from the ddragon
type DDragonClient struct {
	c        *http.Client
	version  string
	host     string
	language string
	rr       *RunesReforged
}

// NewDDragonClient creates a new client for DDragon
func NewDDragonClient() *DDragonClient {

	return &DDragonClient{
		c:        &http.Client{Timeout: time.Second * 20},
		version:  getLatestVersion(),
		host:     "ddragon.leagueoflegends.com",
		language: "en_US",
	}
}

func getLatestVersion() string {
	var versions []string
	resp, err := http.Get("https://ddragon.leagueoflegends.com/api/versions.json")
	if err != nil {
		panic(fmt.Sprintf("failed to fetch riot api versions: %v", err))
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		panic(fmt.Sprintf("failed to fetch riot api version, status: %s", resp.Status))
	}

	if err := json.NewDecoder(resp.Body).Decode(&versions); err != nil {
		panic(fmt.Sprintf("failed to unmarshall version json: %v", err))
	}

	if len(versions) == 0 {
		panic("invalid data received for versions")
	}
	return versions[0]
}

// Request sends a new request to the given api endpoint and unmarshalls the response to given data
func (c *DDragonClient) Request(dataPath string, data interface{}) error {
	u := url.URL{
		Host:   c.host,
		Scheme: "https",
		Path:   fmt.Sprintf("cdn/%s/data/%s/%s", c.version, c.language, dataPath),
	}

	resp, err := c.c.Get(u.String())
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return handleErrorStatus(resp)
	}

	if err := json.NewDecoder(resp.Body).Decode(data); err != nil {
		return err
	}
	return nil
}

// GetProfileIconURL returns url to given profile icon
func (c *DDragonClient) GetProfileIconURL(id int) string {
	u := url.URL{
		Host:   c.host,
		Scheme: "https",
		Path:   fmt.Sprintf("cdn/%s/img/profileicon/%d.png", c.version, id),
	}
	return u.String()
}

// GetRunesReforged returns runes data
func (c *DDragonClient) GetRunesReforged() (*RunesReforged, error) {
	if c.rr != nil {
		return c.rr, nil
	}

	rr, err := readRunes("riotapi/runesreforged.json")
	if err != nil {
		return nil, err
	}
	return rr.ToRunesReforged(), nil
}

func readRunes(fn string) (*RunesReforgedDTO, error) {
	runesJSON, err := ioutil.ReadFile(fn)
	if err != nil {
		return nil, err
	}

	var rr RunesReforgedDTO
	err = json.Unmarshal(runesJSON, &rr)
	if err != nil {
		return nil, err
	}

	return &rr, nil
}

// RunesReforgedDTO contains the runes data as fetched from ddragon
type RunesReforgedDTO []struct {
	ID    int    `json:"id"`
	Key   string `json:"key"`
	Name  string `json:"name"`
	Icon  string `json:"icon"`
	Slots []struct {
		Runes []struct {
			ID        int    `json:"id"`
			Key       string `json:"key"`
			Name      string `json:"name"`
			ShortDesc string `json:"shortDesc"`
			LongDesc  string `json:"longDesc"`
			Icon      string `json:"icon"`
		} `json:"runes"`
	} `json:"slots"`
}

// RunesReforged contains the runes data in more usable form
type RunesReforged struct {
	PerkStyles map[int]PerkStyle
}

// PerkStyle contains information about a Rune
type PerkStyle struct {
	ID    int
	Key   string
	Name  string
	Icon  string
	Runes map[int]Rune
}

// Rune contains information about one rune
type Rune struct {
	ID        int
	Key       string
	Name      string
	ShortDesc string
	LongDesc  string
	Icon      string
}

// ToRunesReforged converts DTO to RunesReforged
func (dtos RunesReforgedDTO) ToRunesReforged() *RunesReforged {
	rr := RunesReforged{PerkStyles: make(map[int]PerkStyle)}
	for _, dto := range dtos {
		ps := PerkStyle{
			ID:    dto.ID,
			Key:   dto.Key,
			Name:  dto.Name,
			Icon:  dto.Icon,
			Runes: make(map[int]Rune),
		}

		for _, slot := range dto.Slots {
			for _, r := range slot.Runes {
				longDesc := beautifyString(r.LongDesc)
				if len(longDesc) == 0 {
					longDesc = beautifyString(r.ShortDesc)
				}
				ps.Runes[r.ID] = Rune{
					ID:        r.ID,
					Key:       r.Key,
					Name:      r.Name,
					ShortDesc: beautifyString(r.ShortDesc),
					LongDesc:  longDesc,
					Icon:      r.Icon,
				}
			}
		}
		rr.PerkStyles[dto.ID] = ps
	}
	return &rr
}

// AllRunes returns all runes in a single map
func (rr *RunesReforged) AllRunes() map[int]Rune {
	runes := make(map[int]Rune)
	for _, perks := range rr.PerkStyles {
		for _, rune := range perks.Runes {
			runes[rune.ID] = rune
		}
	}
	return runes
}

// beautifyString removes all data between <>, @@ or {} from string
func beautifyString(s string) string {
	s = removeStringPart(s, '@', '@')
	s = removeStringPart(s, '<', '>')
	s = removeStringPart(s, '{', '}')

	return s
}

func removeStringPart(s string, r1, r2 rune) string {
	lenS := len(s)
	for {
		s = strings.Map(mapper(r1, r2, false), s)
		if len(s) == lenS {
			break
		}
		lenS = len(s)
	}
	return s
}

func mapper(r1 rune, r2 rune, remove bool) func(r rune) rune {
	return func(r rune) rune {
		if r == r1 {
			if remove && r1 == r2 {
				remove = false
				return -1
			} else if remove {
				return r
			}
			remove = true
		}
		if r == r2 && r1 != r2 {
			if remove {
				remove = false
				return -1
			}
		}
		if remove {
			return -1
		}
		return r
	}
}
