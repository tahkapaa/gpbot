package riotapi

import (
	"fmt"
	"net/http"
)

// MatchAPI implements the Riot Match API methods
type MatchAPI struct {
	c *Client
}

const matchAPIPath = "match"

// MatchListDTO contains list of match references
type MatchListDTO struct {
	Matches    []MatchReferenceDTO
	totalGames int
	startIndex int
	endIndex   int
}

// MatchReferenceDTO contains reference information about a match
type MatchReferenceDTO struct {
	Lane       string
	GameID     int
	Champion   int
	PlatformID string
	Season     int
	Queue      int
	Role       string
	TimeStamp  int
}

// MatchDTO contais all information of one match
type MatchDTO struct {
	seasonID              int
	QueueID               int
	GameID                int
	ParticipantIdentities []ParticipantIdentityDTO
	GameVersion           string
	PlatformID            string
	GameMode              string
	MapID                 int
	GameType              string
	Teams                 []TeamStatsDTO
	Participants          []ParticipantDTO
	GameDuration          int
	GameCreation          int
}

// ParticipantIdentityDTO contains information about a participant
type ParticipantIdentityDTO struct {
	Player        PlayerDTO
	ParticipantID int
}

// PlayerDTO contains information about the player
type PlayerDTO struct {
	CurrentPlatformID string
	SummonerName      string
	MatchHistoryURI   string
	PlatformID        string
	CurrentAccountID  int
	ProfileIcon       int
	SummonerID        int
	AccountID         int
}

// TeamStatsDTO contains information about the team stats
type TeamStatsDTO struct {
	FirstDragon          bool
	FirstInhibitor       bool
	Bans                 []TeamBansDTO
	BaronKills           int
	FirstRiftHerald      bool
	FirstBaron           bool
	RiftHeraldKills      int
	FirstBlood           bool
	TeamID               int
	FirstTower           bool
	VilemawKills         int
	InhibitorKills       int
	TowerKills           int
	DominionVictoryScore int
	Win                  string
	DragonKills          int
}

// TeamBansDTO contains the champion bans
type TeamBansDTO struct {
	PickTurn   int
	ChampionID int
}

// ParticipantDTO contains information about a game participant
type ParticipantDTO struct {
	Stats                     ParticipantStatsDTO
	ParticipantID             int
	Runes                     []RuneDTO
	Timeline                  ParticipantTimelineDTO
	TeamID                    int
	Spell1ID                  int
	Spell2ID                  int
	Masteries                 []MasteryDTO
	HighestAchievedSeasonTier string
	ChampionID                int
}

// ParticipantStatsDTO contains information about the participants stats
type ParticipantStatsDTO struct {
	PhysicalDamageDealt             int
	NeutralMinionsKilledTeamJungle  int
	MagicDamageDealt                int
	TotalPlayerScore                int
	Deaths                          int
	Win                             bool
	NeutralMinionsKilledEnemyJungle int
	AltarsCaptured                  int
	LargestCriticalStrike           int
	TotalDamageDealt                int
	MagicDamageDealtToChampion      int
	VisionWardsBoughtInGame         int
	DamageDealtToObjectives         int
	LargestKillingSpree             int
	Item                            int
	Item1                           int
	Item2                           int
	Item3                           int
	Item4                           int
	Item5                           int
	Item6                           int
	FirstBloodAssist                bool
	VisionScore                     int
	WardsPlaced                     int
	TurretKills                     int
	TripleKills                     int
	DamageSelfMitigated             int
	ChampLevel                      int
	NodeNeutralizeAssist            int
	FirstInhibitorKill              bool
	GoldEarned                      int
	MagicalDamageTaken              int
	Kills                           int
	DoubleKills                     int
	NodeCaptureAssist               bool
	TrueDamageTaken                 int
	NodeNeutralize                  int
	FirstInhibitorAssist            bool
	Assists                         int
	UnrealKills                     int
	NeutralMinionsKilled            int
	ObjectivePlayerScore            int
	CombatPlayerScore               int
	DamageDealtToTurrets            int
	AltarsNeutralized               int
	PhysicalDamageDealtToChampions  int
	GoldSpent                       int
	TrueDamageDealt                 int
	TrueDamageDealtToChampions      int
	ParticipantID                   int
	PentaKills                      int
	TotalHeal                       int
	TotalMinionsKilled              int
	FirstBloodKill                  bool
	NodeCapture                     int
	LargestMultiKill                int
	SightWardsBoughtInGame          int
	TotalDamageDealtToChampions     int
	TotalUnitsHealed                int
	InhibitorKills                  int
	TotalScoreRank                  int
	TotalDamageTaken                int
	KillingSprees                   int
	TimeCCingOthers                 int
	PhysicalDamageTaken             int
}

// RuneDTO contains information about player runes
type RuneDTO struct {
	RuneID int
	Rank   int
}

// ParticipantTimelineDTO contains information about participants doings
type ParticipantTimelineDTO struct {
	Lane                        string
	ParticipantID               int
	CsDiffPerMinDeltas          map[string]float64
	GoldPerMinDeltas            map[string]float64
	XpDiffPerMinDelts           map[string]float64
	CreepsPerMinDeltas          map[string]float64
	XpPerMinDeltas              map[string]float64
	Role                        string
	DamageTakenDiffPerMinDeltas map[string]float64
	DamageTakenPerMinDeltas     map[string]float64
}

// MasteryDTO contains information about layer masteries
type MasteryDTO struct {
	MasteryID int
	Rank      int
}

// MatchByID returns match by given id
func (api MatchAPI) MatchByID(ID int) (*MatchDTO, error) {
	var m MatchDTO
	if err := api.c.Request(matchAPIPath, fmt.Sprintf("matches/%d", ID), &m); err != nil {
		if apiErr, ok := err.(APIError); ok {
			if apiErr.StatusCode == http.StatusNotFound {
				return nil, nil
			}
		}
		return nil, err
	}
	return &m, nil
}

// RecentMatchesByAccountID gets matchlist for last 20 matches played on given account ID and platform ID.
func (api MatchAPI) RecentMatchesByAccountID(ID int) (*MatchListDTO, error) {
	var m MatchListDTO
	if err := api.c.Request(matchAPIPath, fmt.Sprintf("matchlists/by-account/%d/recent", ID), &m); err != nil {
		if apiErr, ok := err.(APIError); ok {
			if apiErr.StatusCode == http.StatusNotFound {
				// TODO: This makes the api hard to use - should be an error if nil is returned
				return nil, nil
			}
		}
		return nil, err
	}
	return &m, nil
}
