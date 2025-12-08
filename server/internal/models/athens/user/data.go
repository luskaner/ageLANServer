package user

import (
	"maps"
	"slices"

	"github.com/luskaner/ageLANServer/server/internal/models/playfab/data"
)

type PunchCardProgress struct {
	Holes                     uint8
	DateOfMostRecentHolePunch data.CustomTime
}

type MissionRewards struct {
	Amount  int
	Scaling string
	ItemId  string
}

type WorldTwist struct {
	Id              string
	ArenaEffectName string
	Title           string
	Description     string
	OwnerIcon       string
	OwnerPortrait   string
	Visualization   string
}

type Opponent struct {
	Civ              string
	Team             int
	Personality      string
	DifficultyOffset int
}

type ChallengeMission struct {
	RowIndex                int `json:"-"`
	Id                      string
	Predecessors            []string
	PositionX               int
	PositionY               int
	Visualization           string
	Map                     string
	Size                    string
	VictoryCondition        string
	GameType                string
	MapVisibility           string
	StartingResources       string
	AllowTitans             bool
	WorldTwists             []WorldTwist
	Opponents               []Opponent
	OpponentsFor2PlayerCoop []Opponent
	Rewards                 []MissionRewards
	MinimapImage            string
	MapPreviewImage         string
}

func (m *ChallengeMission) Equals(other ChallengeMission) bool {
	return m.Id == other.Id &&
		slices.Equal(m.Predecessors, other.Predecessors) &&
		m.PositionX == other.PositionX &&
		m.PositionY == other.PositionY &&
		m.Visualization == other.Visualization &&
		m.Map == other.Map &&
		m.Size == other.Size &&
		m.VictoryCondition == other.VictoryCondition &&
		m.GameType == other.GameType &&
		m.MapVisibility == other.MapVisibility &&
		m.StartingResources == other.StartingResources &&
		m.AllowTitans == other.AllowTitans &&
		slices.Equal(m.WorldTwists, other.WorldTwists) &&
		slices.Equal(m.Opponents, other.Opponents) &&
		slices.Equal(m.OpponentsFor2PlayerCoop, other.OpponentsFor2PlayerCoop) &&
		slices.Equal(m.Rewards, other.Rewards) &&
		m.MinimapImage == other.MinimapImage &&
		m.MapPreviewImage == other.MapPreviewImage
}

type Labyrinth struct {
	Id        int
	Dfficulty string
	Missions  []ChallengeMission
}

func (l *Labyrinth) Equals(other *Labyrinth) bool {
	if l == nil && other == nil {
		return true
	}
	if l == nil || other == nil {
		return false
	}
	return l.Id == other.Id &&
		l.Dfficulty == other.Dfficulty &&
		slices.EqualFunc(l.Missions, other.Missions, func(mission ChallengeMission, mission2 ChallengeMission) bool {
			return mission.Equals(mission2)
		})
}

type ProgressInventory struct {
	SeasonId string
	Item     string
	Rarity   int
}

type Progress struct {
	Lives                      int
	CompletedMissions          []string
	Inventory                  []ProgressInventory
	MissionBeingPlayedRightNow string
}

func (p *Progress) Equals(other *Progress) bool {
	return p.Lives == other.Lives &&
		slices.Equal(p.CompletedMissions, other.CompletedMissions) &&
		slices.Equal(p.Inventory, other.Inventory) &&
		p.MissionBeingPlayedRightNow == other.MissionBeingPlayedRightNow
}

type Challenge struct {
	Labyrinth *data.BaseValue[Labyrinth]
	Progress  *data.BaseValue[Progress]
}

func (c *Challenge) Equals(other *Challenge) bool {
	if (c.Labyrinth == nil) != (other.Labyrinth == nil) {
		return false
	}
	if c.Labyrinth != nil && !c.Labyrinth.Value.Equals(&other.Labyrinth.Value) {
		return false
	}
	if (c.Progress == nil) != (other.Progress == nil) {
		return false
	}
	if c.Progress != nil && !c.Progress.Value.Equals(&other.Progress.Value) {
		return false
	}
	return true
}

type StoryMission struct {
	State                 string
	RewardsAwarded        string
	CompletionCountEasy   uint32
	CompletionCountMedium uint32
	CompletionCountHard   uint32
}

type Data struct {
	Challenge         Challenge
	StoryMissions     map[string]data.BaseValue[StoryMission]
	PunchCardProgress data.BaseValue[PunchCardProgress]
	DataVersion       uint32
}

func (d *Data) Equals(other *Data) bool {
	if d == nil && other == nil {
		return true
	}
	if d == nil || other == nil {
		return false
	}
	return d.Challenge.Equals(&other.Challenge) && maps.Equal(d.StoryMissions, other.StoryMissions) &&
		d.PunchCardProgress == other.PunchCardProgress &&
		d.DataVersion == other.DataVersion
}
