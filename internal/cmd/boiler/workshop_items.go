package boiler

import (
	"fmt"
	"log"
	"os"
	"slices"
	"strings"

	"github.com/MatthiasKunnen/boiler/internal/boiler"
	"github.com/spf13/cobra"
)

var workshopItemsCmd = &cobra.Command{
	Use:   "workshop-items game workshop_items...",
	Short: "Given a list of workshop items, returns the workshop items and their dependencies in dependency order.",
	Args:  cobra.MinimumNArgs(2),
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		toComplete = strings.ToLower(toComplete)
		configFile, err := os.Open(configFilePath)
		if err != nil {
			log.Printf("Error opening config file: %v", err)
			return nil, cobra.ShellCompDirectiveError
		}
		defer configFile.Close()
		b, err := boiler.FromConfigReader(configFile)
		if err != nil {
			log.Printf("Error reading config file: %v", err)
			return nil, cobra.ShellCompDirectiveError
		}

		if len(args) == 0 {
			if len(toComplete) == 0 {
				return b.GetGames(), cobra.ShellCompDirectiveNoFileComp
			}
			var result []string
			for _, gameName := range b.GetGames() {
				if strings.HasPrefix(strings.ToLower(gameName), toComplete) {
					result = append(result, gameName)
				}
			}
			return result, cobra.ShellCompDirectiveNoFileComp
		}

		var result []string
		items, err := b.GetWorkshopItemsForGame(args[0])
		if err != nil {
			return nil, cobra.ShellCompDirectiveError
		}
		for _, item := range items {
			if slices.Contains(args, item.Title) {
				continue
			}
			if toComplete == "" || strings.HasPrefix(strings.ToLower(item.Title), toComplete) {
				result = append(result, item.Title)
			}
		}

		return result, cobra.ShellCompDirectiveNoFileComp
	},
	Run: func(cmd *cobra.Command, args []string) {
		configFile, err := os.Open(configFilePath)
		if err != nil {
			log.Fatalf("failed to open config file: %v", err)
		}
		defer configFile.Close()
		b, err := boiler.FromConfigReader(configFile)
		if err != nil {
			log.Fatalf("failed to read config: %v", err)
		}
		result, err := b.GetWorkshopItemsDependencyOrder(args[0], args[1:]...)
		if err != nil {
			log.Fatalf("failed to get workshop items: %v", err)
		}

		for _, item := range result {
			fmt.Printf("%d # %s\n", item.Id, item.Title)
		}
	},
}
