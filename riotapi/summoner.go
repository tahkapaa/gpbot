package riotapi

import (
	"errors"
	"fmt"
	"net/http"
)

// SummonerAPI implements the Riot Summoner API methods
type SummonerAPI struct {
	c *Client
}

const summonerAPIPath = "summoner"

// SummonerDTO represents a summoner
type SummonerDTO struct {
	Name          string
	ID            int
	AccountID     int
	SummonerLevel int
	ProfileIconID int
	RevisionDate  int
}

// SummonerByName gets a summoner by summoner name
func (api SummonerAPI) SummonerByName(name string) (*SummonerDTO, error) {
	if name == "" {
		return nil, errors.New("missing summoner name")
	}
	var s SummonerDTO
	if err := api.c.Request(summonerAPIPath, "summoners/by-name/"+name, &s); err != nil {
		if apiErr, ok := err.(APIError); ok {
			if apiErr.StatusCode == http.StatusNotFound {
				return nil, nil
			}
		}
		return nil, err
	}
	return &s, nil
}

// SummonerByID gets a summoner by summoner id
func (api SummonerAPI) SummonerByID(ID int) (*SummonerDTO, error) {
	var s SummonerDTO
	if err := api.c.Request(summonerAPIPath, fmt.Sprintf("summoners/%d", ID), &s); err != nil {
		if apiErr, ok := err.(APIError); ok {
			if apiErr.StatusCode == http.StatusNotFound {
				return nil, nil
			}
		}
		return nil, err
	}
	return &s, nil
}
