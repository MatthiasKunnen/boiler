package boiler

type GamesConfig []GameConfig

type GameConfig struct {
	Name                       string
	Id                         int64
	BetaBranch                 string
	WorkshopAppId              int64
	MakeWorkshopItemsLowercase bool
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
