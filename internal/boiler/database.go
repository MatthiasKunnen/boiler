package boiler

import "time"

type Database struct {
	Collections   map[uint64]Collection
	WorkshopItems map[uint64]WorkshopItem
}

type WorkshopItem struct {
	Requires    []int64 `json:",string"`
	TimeCreated time.Time
	TimeUpdated time.Time
	Title       string
}

type Collection struct {
	WorkshopItems []uint64 `json:",string"`
}
