package boiler

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/MatthiasKunnen/boiler/pkg/filecasing"
	"github.com/MatthiasKunnen/boiler/pkg/steamcmd"
	"github.com/MatthiasKunnen/boiler/pkg/steamworkshop"
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

	_, err = os.Stat(filepath.Join(b.config.GamesDir, SteamWorkshopSubDir))
	if err == nil {
		_ = overwriteSymlink(
			filepath.Join(b.config.GamesDir, SteamWorkshopItemPrefix),
			filepath.Join(b.config.GamesDir, "workshop"),
		)
	}

	return nil
}

func (b *Boiler) changeWSItemCasing(toLower bool, items []WorkshopItemWithId) error {
	// @todo handle errors during the case changes
	if len(items) == 0 {
		return nil
	}

	basePath := filepath.Join(b.config.GamesDir, SteamWorkshopItemPrefix)
	if toLower {
		var changedPaths []string
		for _, item := range items {
			suffix := item.PathContentSuffix()
			p := filepath.Join(basePath, suffix)
			err := filecasing.MakeLowerCase(p, func(original string) {
				changedPaths = append(changedPaths, filepath.Join(suffix, original))
			})
			if err != nil {
				return err
			}
		}

		b.db.PathChanges = slices.DeleteFunc(b.db.PathChanges, func(p string) bool {
			for _, item := range items {
				if strings.HasPrefix(p, item.PathContentSuffix()) {
					return true
				}
			}
			return false
		})
		b.db.PathChanges = append(b.db.PathChanges, changedPaths...)
	} else {
		for _, p := range b.db.PathChanges {
			skip := true
			for _, item := range items {
				if strings.HasPrefix(p, item.PathContentSuffix()) {
					skip = false
					break
				}
			}
			if skip {
				continue
			}
			err := filecasing.RestoreCase(basePath, p)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

type DownloadOpts struct {
	DownloadUpToDate bool
	Logout           bool
	Validate         bool
}

func (b *Boiler) Download(ctx context.Context, opts DownloadOpts) error {
	downOpts := steamcmd.Opts{
		LoginUsername:         b.config.LoginUsername,
		InstallDir:            b.config.GamesDir,
		DownloadGames:         make([]steamcmd.DownloadGameOpts, 0, len(b.gamesConfig)),
		DownloadWorkshopItems: nil,
		Logout:                opts.Logout,
		SteamCmdPath:          b.config.SteamCmdPath,
		WorkshopInstallDir:    filepath.Join(b.config.GamesDir, SteamWorkshopSubDir),
	}

	var filenameCasingUpdates []WorkshopItemWithId

	for _, gameConfig := range b.gamesConfig {
		downOpts.DownloadGames = append(downOpts.DownloadGames, steamcmd.DownloadGameOpts{
			Id:         gameConfig.Id,
			BetaBranch: gameConfig.BetaBranch,
			Name:       gameConfig.Name,
			Validate:   opts.Validate,
		})
		ids, err := gameConfig.GetWorkshopItemsOrdered(b.db)
		if err != nil {
			return err
		}
		for _, item := range ids {
			if !opts.DownloadUpToDate && item.LastDownloaded.After(item.TimeUpdated) {
				continue
			}

			if gameConfig.MakeWorkshopItemsLowercase {
				filenameCasingUpdates = append(filenameCasingUpdates, item)
			}

			downOpts.DownloadWorkshopItems = append(
				downOpts.DownloadWorkshopItems,
				steamcmd.DownloadWorkshopItemOpts{
					GameId:         gameConfig.WorkshopAppId,
					WorkshopItemId: item.Id,
				},
			)
		}
	}

	log.Printf("%d games will be updated", len(downOpts.DownloadGames))
	log.Printf("%d workshop items will be updated", len(downOpts.DownloadWorkshopItems))

	err := b.changeWSItemCasing(false, filenameCasingUpdates)
	switch {
	case errors.Is(err, os.ErrNotExist):
	case err != nil:
		return fmt.Errorf("error restoring workshop items file casing: %w", err)
	}
	err = steamcmd.Exec(ctx, downOpts)
	if err != nil {
		caseErr := b.changeWSItemCasing(true, filenameCasingUpdates)
		return errors.Join(err, caseErr)
	}

	for _, downItem := range downOpts.DownloadWorkshopItems {
		item := b.db.WorkshopItems[downItem.WorkshopItemId]
		item.LastDownloaded = time.Now()
		b.db.WorkshopItems[downItem.WorkshopItemId] = item
	}

	err = b.changeWSItemCasing(true, filenameCasingUpdates)
	if err != nil {
		err = fmt.Errorf("error changing workshop items file casing to lower: %w", err)
	}
	return errors.Join(b.Save(), b.createSymlinks(), err)
}

func (b *Boiler) createSymlinks() error {
	var resultErr error
	for _, game := range b.gamesConfig {
		skip := true
		for _, idc := range game.WorkshopItems {
			workshopItem, ok := b.db.WorkshopItems[idc.Id]
			if !ok {
				return fmt.Errorf("workshop item %d not found", idc.Id)
			}
			if workshopItem.LastDownloaded.IsZero() {
				continue
			}
			skip = false
			break
		}
		if skip {
			continue
		}
		err := overwriteSymlink(
			filepath.Join(
				b.config.GamesDir,
				SteamWorkshopItemPrefix,
				strconv.Itoa(game.WorkshopAppId),
			),
			filepath.Join(b.config.GamesDir, game.Name, "mods"),
		)
		switch {
		case errors.Is(err, os.ErrExist):
		case err != nil:
			resultErr = errors.Join(resultErr, err)
		}
	}

	return resultErr
}

type UpdateOpts struct {
}

// UpdateDatabase updates the database based on the games configuration.
// All workshop items and collections will be fetched and updated.
func (b *Boiler) UpdateDatabase(ctx context.Context, opts UpdateOpts) error {
	nextWorkshopItems := make(map[uint64]struct{})
	nextCollections := make(map[uint64]struct{})
	collectionsSeen := make(map[uint64]struct{})
	for _, game := range b.gamesConfig {
		for _, collection := range game.WorkshopCollections {
			nextCollections[collection.Id] = struct{}{}
		}
	}

	for {
		if len(nextCollections) == 0 {
			break
		}
		collections := nextCollections
		nextCollections = make(map[uint64]struct{})
		for _, keys := range batchMapKeys(collections, 100) {
			log.Printf("Getting info of %d collections", len(keys))
			result, err := steamworkshop.CollectionDetailsApi(ctx, keys...)
			if err != nil {
				return err
			}
			for _, collectionDetails := range result {
				c := Collection{
					Items: make([]CollectionItem, 0, len(collectionDetails.Items)),
				}
				for _, item := range collectionDetails.Items {
					c.Items = append(c.Items, CollectionItem{
						Id:   item.Id,
						Type: item.Type,
					})

					switch item.Type {
					case steamworkshop.CollectionDetailFileTypeWorkshopItem:
						nextWorkshopItems[item.Id] = struct{}{}
					case steamworkshop.CollectionDetailFileTypeUnknown:
						log.Printf(
							"Unknown collection type: %d for workshop item %d. Contact the developer.",
							item.Type,
							item.Id,
						)
					case steamworkshop.CollectionDetailFileTypeCollection:
						if _, ok := collectionsSeen[item.Id]; !ok {
							collectionsSeen[item.Id] = struct{}{}
							nextCollections[item.Id] = struct{}{}
						}
					}
				}
				b.db.Collections[collectionDetails.CollectionId] = c
			}
		}
	}

	for _, config := range b.gamesConfig {
		for _, items := range config.WorkshopDependencyAdd {
			for _, item := range items {
				nextWorkshopItems[item.Id] = struct{}{}
			}
		}
		for _, items := range config.WorkshopDependencyRemove {
			for _, item := range items {
				nextWorkshopItems[item.Id] = struct{}{}
			}
		}
		for _, item := range config.WorkshopItems {
			nextWorkshopItems[item.Id] = struct{}{}
		}
	}

	for id := range nextWorkshopItems {
		item, ok := b.db.WorkshopItems[id]
		if !ok {
			continue
		}
		for _, id := range item.Requires {
			nextWorkshopItems[id] = struct{}{}
		}
		for _, id := range b.getRequiredWorkshopIds(item.Requires) {
			nextWorkshopItems[id] = struct{}{}
		}
	}

	workshopItemsSeen := make(map[uint64]struct{})
	for {
		if len(nextWorkshopItems) == 0 {
			break
		}
		workshopItems := nextWorkshopItems
		nextWorkshopItems = make(map[uint64]struct{})
		updateRequirements := make(map[uint64]struct{})

		for _, ids := range batchMapKeys(workshopItems, 100) {
			log.Printf("Getting info of %d workshop items", len(ids))
			fileDetails, err := steamworkshop.FileDetailsApi(ctx, ids...)
			if err != nil {
				return err
			}

			for _, detail := range fileDetails {
				existing, alreadyExists := b.db.WorkshopItems[detail.Id]
				newItem := WorkshopItem{
					CreatorAppId:   detail.CreatorAppId,
					LastDownloaded: time.Time{},
					LastRefreshed:  time.Now(),
					Requires:       nil,
					TimeCreated:    detail.TimeCreated,
					TimeUpdated:    detail.TimeUpdated,
					Title:          detail.Title,
				}
				if alreadyExists {
					newItem.LastDownloaded = existing.LastDownloaded
					newItem.Requires = existing.Requires
				}
				if !alreadyExists || existing.TimeUpdated.Before(detail.TimeUpdated) {
					updateRequirements[detail.Id] = struct{}{}
				}

				b.db.WorkshopItems[detail.Id] = newItem
				workshopItemsSeen[detail.Id] = struct{}{}
			}
		}

		for workshopId := range updateRequirements {
			log.Printf("Getting dependencies of workshop item %d", workshopId)
			fileDetails, err := steamworkshop.GetFileDetailsWeb(ctx, workshopId)
			if err != nil {
				return err
			}

			item, ok := b.db.WorkshopItems[workshopId]
			if !ok {
				return fmt.Errorf(
					"workshop item %d not in database on requirement update",
					workshopId,
				)
			}

			item.Requires = item.Requires[:0]
			for _, requiredItem := range fileDetails.RequiredItems {
				item.Requires = append(item.Requires, requiredItem.Id)
				if _, ok := workshopItemsSeen[requiredItem.Id]; !ok {
					nextWorkshopItems[requiredItem.Id] = struct{}{}
				}
			}

			b.db.WorkshopItems[workshopId] = item
		}
	}

	err := b.Save()
	if err != nil {
		return err
	}

	return nil
}

func (b *Boiler) GetWorkshopItemsForGame(gameName string) ([]WorkshopItemWithId, error) {
	for _, config := range b.gamesConfig {
		if config.Name != gameName {
			continue
		}
		ordered, err := config.GetWorkshopItemsOrdered(b.db)
		if err != nil {
			return nil, err
		}
		return ordered, nil
	}

	return nil, nil
}

func (b *Boiler) GetWorkshopItemsDependencyOrder(
	gameName string,
	names ...string,
) ([]WorkshopItemWithId, error) {
	var gameConfig GameConfig
	for _, config := range b.gamesConfig {
		if gameName == config.Name {
			gameConfig = config
		}
	}
	if gameConfig.Name == "" {
		return nil, nil
	}

	ids := make([]uint64, 0, len(names))
	for _, name := range names {
		var foundItem WorkshopItemWithId
		for id, item := range b.db.WorkshopItems {
			if item.Title == name {
				foundItem = WorkshopItemWithId{
					Id:           id,
					WorkshopItem: item,
				}
				break
			}
		}

		if foundItem.Id == 0 {
			return nil, fmt.Errorf("could not find workshop item %s", name)
		}
		ids = append(ids, foundItem.Id)
	}

	return gameConfig.WorkshopItemsInOrder(b.db, ids...)
}

func (b *Boiler) GetGames() []string {
	result := make([]string, 0, len(b.gamesConfig))
	for _, config := range b.gamesConfig {
		result = append(result, config.Name)
	}
	return result
}

func (b *Boiler) Logout(ctx context.Context) error {
	return steamcmd.LogOutUser(ctx, b.config.SteamCmdPath, b.config.LoginUsername)
}

func (b *Boiler) getRequiredWorkshopIds(workshopIds []uint64) []uint64 {
	var result []uint64
	for _, id := range workshopIds {
		item, ok := b.db.WorkshopItems[id]
		if !ok {
			continue
		}
		result = append(result, item.Requires...)
		result = append(result, b.getRequiredWorkshopIds(item.Requires)...)
	}
	return result
}

func (b *Boiler) loadDatabase() error {
	dbFile, err := os.Open(b.config.DatabasePath)
	switch {
	case errors.Is(err, os.ErrNotExist):
		b.db = &Database{
			Collections:   map[uint64]Collection{},
			WorkshopItems: map[uint64]WorkshopItem{},
		}
		return nil
	case err != nil:
		return err
	}
	defer dbFile.Close()
	var db Database
	err = json.UnmarshalRead(dbFile, &db)
	if err != nil {
		return err
	}
	b.db = &db
	if b.db.Collections == nil {
		b.db.Collections = map[uint64]Collection{}
	}
	if b.db.WorkshopItems == nil {
		b.db.WorkshopItems = map[uint64]WorkshopItem{}
	}
	return nil
}

func (b *Boiler) saveDatabase() error {
	dbFile, err := os.Create(b.config.DatabasePath)
	if err != nil {
		return err
	}
	defer dbFile.Close()
	err = json.MarshalWrite(dbFile, b.db)
	if err != nil {
		return err
	}

	_, err = dbFile.Write([]byte("\n"))
	return err
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
	err = json.MarshalWrite(dbFile, b.gamesConfig, jsontext.WithIndent("\t"))
	if err != nil {
		return err
	}
	_, err = dbFile.Write([]byte("\n"))
	return err
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

type ConfigOpt func(config *Config)

func WithLoginUsername(username string) ConfigOpt {
	return func(config *Config) {
		if username != "" {
			config.LoginUsername = username
		}
	}
}

func FromConfigReader(r io.Reader, opts ...ConfigOpt) (*Boiler, error) {
	var config Config
	err := json.UnmarshalRead(r, &config)
	if err != nil {
		return nil, err
	}

	for _, optFunc := range opts {
		optFunc(&config)
	}

	return FromConfig(config)
}
