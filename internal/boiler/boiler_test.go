package boiler_test

import (
	"bytes"
	"testing"
	"time"

	"github.com/MatthiasKunnen/boiler/internal/boiler"
	"github.com/go-json-experiment/json"
	"github.com/stretchr/testify/assert"
)

func TestBoilerConfig(t *testing.T) {
	b, err := boiler.FromConfig(boiler.Config{
		DatabasePath:  "testdata/db.json",
		GamesConfPath: "testdata/games.json",
		GamesDir:      "/dev/null",
		LoginUsername: "anonymous",
		SteamCmdPath:  "/usr/bin/false",
	})
	assert.NoError(t, err)
	assert.NoError(t, b.Save())
}

func TestDatabaseJsonRoundtrip(t *testing.T) {
	input := boiler.Database{
		Collections: map[uint64]boiler.Collection{
			961618554: {
				WorkshopItems: []uint64{
					158164864,
					50,
				},
			},
		},
		WorkshopItems: map[uint64]boiler.WorkshopItem{
			463939057: {
				Requires:    []int64{120, 2},
				TimeCreated: time.Unix(1758384867, 0),
				TimeUpdated: time.Unix(1758384867, 0),
				Title:       "Hello",
			},
		},
	}
	var outJson bytes.Buffer
	err := json.MarshalWrite(&outJson, input)
	assert.NoError(t, err)

	var actual boiler.Database
	err = json.Unmarshal(outJson.Bytes(), &actual)
	assert.NoError(t, err)

	assert.Equal(t, input, actual)
}

func TestGamesConfigJsonRoundtrip(t *testing.T) {
	input := boiler.GamesConfig{
		{
			Name:                       "Arma3",
			Id:                         233780,
			WorkshopAppId:              107410,
			BetaBranch:                 "creatordlc",
			MakeWorkshopItemsLowercase: true,
			WorkshopItems: []boiler.IdWithComment{
				{463939057, "ace"},
				{2950011244, "Sail to South_Eastern Asia"},
			},
			WorkshopDependencyAdd: map[uint64][]boiler.IdWithComment{
				2950011244: {
					{11, "add this"},
				},
			},
			WorkshopDependencyRemove: map[uint64][]boiler.IdWithComment{
				463939057: {
					{12, "remove this"},
				},
			},
			WorkshopCollections: []boiler.IdWithComment{
				{18474846, "some collection"},
			},
		},
	}
	var outJson bytes.Buffer
	err := json.MarshalWrite(&outJson, input)
	assert.NoError(t, err)

	var actual boiler.GamesConfig
	err = json.Unmarshal(outJson.Bytes(), &actual)
	assert.NoError(t, err)

	assert.Equal(t, input, actual)
}

func TestGamesConfig_UpdateComments(t *testing.T) {
	actual := boiler.GamesConfig{
		{
			Name:                       "Arma3",
			Id:                         233780,
			WorkshopAppId:              107410,
			BetaBranch:                 "creatordlc",
			MakeWorkshopItemsLowercase: true,
			WorkshopItems: []boiler.IdWithComment{
				{463939057, ""},
				{2950011244, ""},
			},
			WorkshopDependencyAdd: map[uint64][]boiler.IdWithComment{
				2950011244: {
					{11, ""},
				},
			},
			WorkshopDependencyRemove: map[uint64][]boiler.IdWithComment{
				463939057: {
					{12, ""},
				},
			},
		},
	}
	expected := boiler.GamesConfig{
		{
			Name:                       "Arma3",
			Id:                         233780,
			WorkshopAppId:              107410,
			BetaBranch:                 "creatordlc",
			MakeWorkshopItemsLowercase: true,
			WorkshopItems: []boiler.IdWithComment{
				{463939057, "ace"},
				{2950011244, "Sail to South_Eastern Asia"},
			},
			WorkshopDependencyAdd: map[uint64][]boiler.IdWithComment{
				2950011244: {
					{11, "add this"},
				},
			},
			WorkshopDependencyRemove: map[uint64][]boiler.IdWithComment{
				463939057: {
					{12, "remove this"},
				},
			},
		},
	}
	db := boiler.Database{
		Collections: map[uint64]boiler.Collection{
			18474846: {},
		},
		WorkshopItems: map[uint64]boiler.WorkshopItem{
			463939057: {
				Title: "ace",
			},
			2950011244: {
				Title: "Sail to South_Eastern Asia",
			},
			11: {
				Title: "add this",
			},
			12: {
				Title: "remove this",
			},
		},
	}

	actual.UpdateComments(&db)

	assert.Equal(t, expected, actual)
}
