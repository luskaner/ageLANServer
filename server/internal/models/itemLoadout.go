package models

import (
	"iter"
	"maps"

	mapset "github.com/deckarep/golang-set/v2"
	i "github.com/luskaner/ageLANServer/server/internal"
)

type ItemLoadout interface {
	Encode(userId int32) i.A
	Update(name string, typ int32, itemOrLocIds mapset.Set[int32])
}

type ItemLoadouts interface {
	Get(id int32) ItemLoadout
	NewItemLoadout(name string, typ int32, itemOrLocIds mapset.Set[int32], userId int32) i.A
	Iter() iter.Seq[ItemLoadout]
}

type MainItemLoadout struct {
	Id           int32
	Name         string
	ItemOrLocIds mapset.Set[int32]
	Type         int32
	// TODO: Implement fields
	// AtributeKeys  []any
	// RecurseLevels []int32
}

func (l *MainItemLoadout) Encode(userId int32) i.A {
	return i.A{
		l.Id,
		userId,
		l.Name,
		l.Type,
		// FIXME: Change to appropriate data
		"[]",
	}
}

func (l *MainItemLoadout) Update(name string, typ int32, itemOrLocIds mapset.Set[int32]) {
	l.Name = name
	l.Type = typ
	l.ItemOrLocIds = itemOrLocIds
}

type MainItemLoadouts struct {
	ItemLoadouts map[int32]ItemLoadout `json:"itemLoadouts"`
}

func (l *MainItemLoadouts) Get(id int32) ItemLoadout {
	itemLoadout, ok := l.ItemLoadouts[id]
	if !ok {
		return nil
	}
	return itemLoadout
}

func (l *MainItemLoadouts) NewItemLoadout(name string, typ int32, itemOrLocIds mapset.Set[int32], userId int32) i.A {
	var itemloadoutId int32
	i.WithRng(func(rand *i.RandReader) {
		for itemloadoutId = rand.Int32(); ; {
			if _, exists := l.ItemLoadouts[itemloadoutId]; !exists {
				break
			}
		}
	})
	itemLoadout := &MainItemLoadout{
		Id:           itemloadoutId,
		Name:         name,
		ItemOrLocIds: itemOrLocIds,
		Type:         typ,
	}
	l.ItemLoadouts[itemloadoutId] = itemLoadout
	return itemLoadout.Encode(userId)
}

func (l *MainItemLoadouts) Iter() iter.Seq[ItemLoadout] {
	return maps.Values(l.ItemLoadouts)
}

type ItemLoadoutsUpgradableDefaultData struct {
	InitialUpgradableDefaultData[ItemLoadouts]
}

func NewItemLoadoutsUpgradableDefaultData() *ItemLoadoutsUpgradableDefaultData {
	return &ItemLoadoutsUpgradableDefaultData{
		InitialUpgradableDefaultData: InitialUpgradableDefaultData[ItemLoadouts]{},
	}
}

func (i *ItemLoadoutsUpgradableDefaultData) Default() ItemLoadouts {
	return &MainItemLoadouts{
		ItemLoadouts: make(map[int32]ItemLoadout),
	}
}
