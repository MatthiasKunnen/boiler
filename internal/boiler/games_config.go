package boiler

import (
	"fmt"
	"slices"
)

type GamesConfig []GameConfig

type GameConfig struct {
	Name                       string
	Id                         int
	BetaBranch                 string
	WorkshopAppId              int
	MakeWorkshopItemsLowercase bool
	PostInstall                string
	WorkshopItems              []IdWithComment
	WorkshopDependencyAdd      map[uint64][]IdWithComment
	WorkshopDependencyRemove   map[uint64][]IdWithComment
	WorkshopCollections        []IdWithComment
}

func (config GamesConfig) UpdateComments(db *Database) {
	for _, gameConfig := range config {
		gameConfig.UpdateComments(db)
	}
}

func (gc GameConfig) UpdateComments(db *Database) {
	for key, item := range gc.WorkshopItems {
		workshopItem, ok := db.WorkshopItems[item.Id]
		if !ok {
			continue
		}
		item.Comment = workshopItem.Title
		gc.WorkshopItems[key] = item
	}
	for _, items := range gc.WorkshopDependencyAdd {
		for i, item := range items {
			workshopItem, ok := db.WorkshopItems[item.Id]
			if !ok {
				continue
			}
			item.Comment = workshopItem.Title
			items[i] = item
		}
	}
	for _, items := range gc.WorkshopDependencyRemove {
		for i, item := range items {
			workshopItem, ok := db.WorkshopItems[item.Id]
			if !ok {
				continue
			}
			item.Comment = workshopItem.Title
			items[i] = item
		}
	}
}

func (gc GameConfig) GetWorkshopItemsOrdered(db *Database) ([]WorkshopItemWithId, error) {
	ids := make([]uint64, 0, len(gc.WorkshopItems))
	for _, item := range gc.WorkshopItems {
		ids = append(ids, item.Id)
	}
	return gc.WorkshopItemsInOrder(db, ids...)
}

func (gc GameConfig) WorkshopItemsInOrder(db *Database, ids ...uint64) ([]WorkshopItemWithId, error) {
	var results []WorkshopItemWithId
	idsSeen := make(map[uint64]struct{})
	results, err := gc.getWorkshopItemsOrdered(db, results, idsSeen, ids...)
	if err != nil {
		return nil, err
	}

	return results, nil
}

func (gc GameConfig) getWorkshopItemsOrdered(
	db *Database,
	result []WorkshopItemWithId,
	skip map[uint64]struct{},
	ids ...uint64,
) ([]WorkshopItemWithId, error) {
	var err error
	for _, id := range ids {
		if item, ok := db.WorkshopItems[id]; ok {
			var add []uint64
			exclude := gc.WorkshopDependencyRemove[id]

			for _, requiredId := range item.Requires {
				if slices.ContainsFunc(exclude, func(s IdWithComment) bool {
					return s.Id == requiredId
				}) {
					continue
				}
				add = append(add, requiredId)
			}

			for _, idc := range gc.WorkshopDependencyAdd[id] {
				add = append(add, idc.Id)
			}

			result, err = gc.getWorkshopItemsOrdered(db, result, skip, add...)
			if err != nil {
				return nil, err
			}

			if _, ok := skip[id]; ok {
				continue
			}
			skip[id] = struct{}{}
			result = append(result, WorkshopItemWithId{
				Id:           id,
				WorkshopItem: item,
			})
		} else if collection, ok := db.Collections[id]; ok {
			if _, ok := skip[id]; ok {
				continue
			}
			skip[id] = struct{}{}
			for _, collectionItem := range collection.Items {
				result, err = gc.getWorkshopItemsOrdered(db, result, skip, collectionItem.Id)
				if err != nil {
					return nil, err
				}
			}
		} else {
			return nil, fmt.Errorf("%d is not a collection nor a workshopitem", id)
		}

	}

	return result, nil
}
