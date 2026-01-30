package models

type ProfilePropertiesUpgradableDefaultData struct {
	InitialUpgradableDefaultData[*map[string]string]
}

func NewProfilePropertiesUpgradableDefaultData() *ProfilePropertiesUpgradableDefaultData {
	return &ProfilePropertiesUpgradableDefaultData{
		InitialUpgradableDefaultData: InitialUpgradableDefaultData[*map[string]string]{},
	}
}

func (p *ProfilePropertiesUpgradableDefaultData) Default() *map[string]string {
	return &map[string]string{}
}
