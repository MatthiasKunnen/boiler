package boiler

import (
	"fmt"
	"io"
	"os"

	"github.com/go-json-experiment/json"
	"github.com/go-json-experiment/json/jsontext"
)

type Boiler struct {
	config      Config
	db          *Database
	gamesConfig GamesConfig
}

type SyncOpts struct {
	// Overwrites the username set in the config.
	LoginUsername string

	// If true, the user will be logged out after execution completes.
	Logout bool
	// Whether updates to workshop files and collections will be performed.
	UpdateDatabase bool
}

func (b *Boiler) Save() error {
	err := b.saveDatabase()
	if err != nil {
		return err
	}
	err = b.saveGamesConfig()
	if err != nil {
		return err
	}
	return nil
}

func (b *Boiler) loadDatabase() error {
	dbFile, err := os.Open(b.config.DatabasePath)
	if err != nil {
		return err
	}
	defer dbFile.Close()
	var db Database
	err = json.UnmarshalRead(dbFile, &db)
	if err != nil {
		return err
	}
	b.db = &db
	return nil
}

func (b *Boiler) saveDatabase() error {
	dbFile, err := os.Create(b.config.DatabasePath)
	if err != nil {
		return err
	}
	defer dbFile.Close()
	return json.MarshalWrite(dbFile, b.db)
}

func (b *Boiler) loadGamesConfig() error {
	dbFile, err := os.Open(b.config.GamesConfPath)
	if err != nil {
		return err
	}
	defer dbFile.Close()
	var games GamesConfig
	err = json.UnmarshalRead(dbFile, &games)
	if err != nil {
		return err
	}
	b.gamesConfig = games
	return nil
}

func (b *Boiler) saveGamesConfig() error {
	b.gamesConfig.UpdateComments(b.db)
	dbFile, err := os.Create(b.config.GamesConfPath)
	if err != nil {
		return err
	}
	defer dbFile.Close()
	return json.MarshalWrite(dbFile, b.gamesConfig, jsontext.WithIndent("\t"))
}

func FromConfig(config Config) (*Boiler, error) {
	if config.DatabasePath == "" {
		return nil, fmt.Errorf("config.DatabasePath is not set")
	}
	if config.GamesConfPath == "" {
		return nil, fmt.Errorf("config.GamesConfPath is not set")
	}
	b := &Boiler{config: config}

	err := b.loadDatabase()
	if err != nil {
		return nil, err
	}

	err = b.loadGamesConfig()
	if err != nil {
		return nil, err
	}

	return b, nil
}

func FromConfigReader(r io.Reader) (*Boiler, error) {
	var config Config
	err := json.UnmarshalRead(r, &config)
	if err != nil {
		return nil, err
	}

	return FromConfig(config)
}
