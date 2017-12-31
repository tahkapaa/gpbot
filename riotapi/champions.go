package riotapi

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
)

// LoLStaticDataAPI implements the Riot LoL Static Data API methods
type LoLStaticDataAPI struct {
	c         *Client
	champions *Champions
}

const staticDataAPIPath = "static-data"

// ChampionsDTO contains the Champions data
type ChampionsDTO struct {
	Type    string              `json:"type,omitempty"`
	Version string              `json:"version,omitempty"`
	Data    map[string]Champion `json:"data,omitempty"`
}

// Champions contains the Champions indexed by their id's
type Champions struct {
	Data map[int]Champion
}

// Champion is the League of Legends champion
type Champion struct {
	ID    int    `json:"id,omitempty"`
	Name  string `json:"name,omitempty"`
	Title string `json:"title,omitempty"`
	Key   string `json:"key,omitempty"`
}

func (champs *Champions) fromDTO(champsDTO *ChampionsDTO) {
	champs.Data = make(map[int]Champion, len(champsDTO.Data))
	for _, v := range champsDTO.Data {
		champs.Data[v.ID] = v
	}
}

func (cdto *ChampionsDTO) toChampions() *Champions {
	var champs Champions
	champs.Data = make(map[int]Champion, len(cdto.Data))
	for _, v := range cdto.Data {
		champs.Data[v.ID] = v
	}
	return &champs
}

func readChamps(fn string) (*ChampionsDTO, error) {
	champsJSON, err := ioutil.ReadFile(fn)
	if err != nil {
		return nil, err
	}

	var champs ChampionsDTO
	err = json.Unmarshal(champsJSON, &champs)
	if err != nil {
		return nil, err
	}

	return &champs, nil
}

// Champions retrieves champion list lol/static-data/v3/champions
func (api *LoLStaticDataAPI) Champions() (*Champions, error) {
	if api.champions != nil {
		return api.champions, nil
	}
	var cdto *ChampionsDTO
	err := api.c.Request(staticDataAPIPath, "champions", &cdto)
	if err != nil {
		fmt.Printf("unable to fetch champions from riot: %v\n", err)
		cdto, err = readChamps("riotapi/champions.json")
		if err != nil {
			fmt.Printf("unable to fetch champions from file: %v\n", err)
			return nil, err
		}
	}
	api.champions = cdto.toChampions()
	return api.champions, nil
}
