package boiler_test

import (
	"testing"

	"github.com/MatthiasKunnen/boiler/internal/boiler"
	"github.com/MatthiasKunnen/boiler/pkg/steamworkshop"
	"github.com/stretchr/testify/assert"
)

func TestGameConfig_GetWorkshopItemsOrdered_WorkshopOnly(t *testing.T) {
	db := &boiler.Database{
		Collections: map[uint64]boiler.Collection{},
		WorkshopItems: map[uint64]boiler.WorkshopItem{
			1:   {Requires: []uint64{3}},
			3:   {Requires: []uint64{4}},
			4:   {Requires: []uint64{5}},
			5:   {Requires: []uint64{50}},
			10:  {Requires: []uint64{11, 13, 50, 14}},
			11:  {Requires: []uint64{12}},
			12:  {},
			13:  {Requires: []uint64{12}},
			14:  {},
			50:  {},
			100: {Requires: []uint64{110, 160}},
			110: {Requires: []uint64{111, 112}},
			111: {},
			112: {},
			140: {},
			160: {},
			180: {},
		},
	}
	gc := boiler.GameConfig{
		Name:                       "",
		Id:                         0,
		BetaBranch:                 "",
		WorkshopAppId:              0,
		MakeWorkshopItemsLowercase: false,
		WorkshopItems: []boiler.IdWithComment{
			{1, ""},
			{10, ""},
			{100, ""},
		},
		WorkshopDependencyAdd: map[uint64][]boiler.IdWithComment{
			100: {{140, ""}},
			110: {{112, ""}},
			160: {{180, ""}},
		},
		WorkshopDependencyRemove: map[uint64][]boiler.IdWithComment{
			100: {{110, ""}},
		},
		WorkshopCollections: nil,
	}
	expectedIds := []uint64{
		50, 5, 4, 3, 1, 12, 11, 13, 14, 10, 180, 160, 140, 100,
	}
	actual, err := gc.GetWorkshopItemsOrdered(db)
	actualIds := make([]uint64, 0, len(actual))
	for _, id := range actual {
		actualIds = append(actualIds, id.Id)
	}
	assert.NoError(t, err)
	assert.Equal(t, expectedIds, actualIds)
}

func TestGameConfig_GetWorkshopItemsOrdered_WithCollections(t *testing.T) {
	db := &boiler.Database{
		Collections: map[uint64]boiler.Collection{
			12: {
				Items: []boiler.CollectionItem{
					{
						Id:   5,
						Type: steamworkshop.CollectionDetailFileTypeWorkshopItem,
					},
					{
						Id:   71,
						Type: steamworkshop.CollectionDetailFileTypeCollection,
					},
					{
						Id:   72,
						Type: steamworkshop.CollectionDetailFileTypeCollection,
					},
					{
						Id:   90,
						Type: steamworkshop.CollectionDetailFileTypeWorkshopItem,
					},
				},
			},
			71: {
				Items: []boiler.CollectionItem{
					{
						Id:   80,
						Type: steamworkshop.CollectionDetailFileTypeCollection,
					},
				},
			},
			72: {
				Items: []boiler.CollectionItem{},
			},
			80: {
				Items: []boiler.CollectionItem{
					{
						Id:   81,
						Type: steamworkshop.CollectionDetailFileTypeCollection,
					},
				},
			},
		},
		WorkshopItems: map[uint64]boiler.WorkshopItem{
			1:   {Requires: []uint64{3}},
			3:   {Requires: []uint64{4}},
			4:   {Requires: []uint64{5}},
			5:   {Requires: []uint64{50}},
			10:  {Requires: []uint64{11, 13, 50, 14}},
			11:  {Requires: []uint64{12}},
			13:  {Requires: []uint64{12}},
			14:  {},
			50:  {},
			81:  {},
			90:  {Requires: []uint64{91, 92}},
			91:  {},
			92:  {},
			93:  {},
			100: {Requires: []uint64{110, 160}},
			110: {Requires: []uint64{111, 112}},
			111: {},
			112: {},
			140: {},
			160: {},
			180: {},
		},
	}
	gc := boiler.GameConfig{
		Name:                       "",
		Id:                         0,
		BetaBranch:                 "",
		WorkshopAppId:              0,
		MakeWorkshopItemsLowercase: false,
		WorkshopItems: []boiler.IdWithComment{
			{1, ""},
			{10, ""},
			{100, ""},
		},
		WorkshopDependencyAdd: map[uint64][]boiler.IdWithComment{
			90:  {{93, ""}},
			100: {{140, ""}},
			110: {{112, ""}},
			160: {{180, ""}},
		},
		WorkshopDependencyRemove: map[uint64][]boiler.IdWithComment{
			100: {{110, ""}},
		},
		WorkshopCollections: nil,
	}
	expectedIds := []uint64{
		50, 5, 4, 3, 1, 81, 91, 92, 93, 90, 11, 13, 14, 10, 180, 160, 140, 100,
	}
	actual, err := gc.GetWorkshopItemsOrdered(db)
	actualIds := make([]uint64, 0, len(actual))
	for _, id := range actual {
		actualIds = append(actualIds, id.Id)
	}
	assert.NoError(t, err)
	assert.Equal(t, expectedIds, actualIds)
}
