package models

import (
	i "github.com/luskaner/ageLANServer/server/internal"
)

type PresenceDefinitions interface {
	Initialize(presence i.A)
	Get(id int32) *PresenceDefinition
}

type PresenceDefinition interface {
	GetId() int32
	GetLabel() *string
}

type MainPresenceDefinitions struct {
	data map[int32]PresenceDefinition
}

func (pd *MainPresenceDefinitions) Initialize(presence i.A) {
	rawData := presence[1].(i.A)
	pd.data = make(map[int32]PresenceDefinition, len(rawData))
	for _, rawPresence := range rawData {
		rawPresenceArr := rawPresence.(i.A)
		id := int32(rawPresenceArr[0].(float64))
		var label *string
		if tmpLabel := rawPresenceArr[2].(string); tmpLabel != "" {
			label = &tmpLabel
		}
		pd.data[id] = &MainPresenceDefinition{
			id:    id,
			label: label,
		}
	}
}

func (pd *MainPresenceDefinitions) Get(id int32) *PresenceDefinition {
	presenceDefinition, ok := pd.data[id]
	if !ok {
		return nil
	}
	return &presenceDefinition
}

type MainPresenceDefinition struct {
	id    int32
	label *string
}

func (pd *MainPresenceDefinition) GetId() int32 {
	return pd.id
}

func (pd *MainPresenceDefinition) GetLabel() *string {
	return pd.label
}
