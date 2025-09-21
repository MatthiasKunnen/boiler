package boiler

import (
	"time"

	"github.com/MatthiasKunnen/boiler/pkg/steamworkshop"
)

type Database struct {
	Collections   map[uint64]Collection
	WorkshopItems map[uint64]WorkshopItem
}

type WorkshopItem struct {
	// The ID of the game that the workshop item relates to.
	CreatorAppId int
	// Time when the workshop item was last downloaded.
	LastDownloaded time.Time
	// Time when the details of the workshop item were last retrieved.
	LastRefreshed time.Time
	Requires      []uint64 `json:",string"`
	// Time when the workshop item was created.
	TimeCreated time.Time
	// Time when the workshop item was last updated.
	TimeUpdated time.Time
	Title       string
}

type WorkshopItemWithId struct {
	Id uint64
	WorkshopItem
}

type Collection struct {
	Items []CollectionItem
}

type CollectionItem struct {
	Id   uint64 `json:",string"`
	Type steamworkshop.CollectionDetailFileType
}
