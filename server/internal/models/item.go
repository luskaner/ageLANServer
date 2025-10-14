package models

type MainItemLoadout struct {
	id           int32
	name         string
	typ          int32
	itemOrLocIDs []int32
}
