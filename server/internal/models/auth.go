package models

import (
	"time"
)

type AuthUpgradableDefaultData struct {
	InitialUpgradableDefaultData[*time.Time]
}

func NewAuthUpgradableDefaultData() *AuthUpgradableDefaultData {
	return &AuthUpgradableDefaultData{
		InitialUpgradableDefaultData: InitialUpgradableDefaultData[*time.Time]{},
	}
}

func (a *AuthUpgradableDefaultData) Default() *time.Time {
	return new(time.Time)
}
