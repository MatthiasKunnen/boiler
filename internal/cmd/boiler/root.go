package boiler

import (
	"log"

	"github.com/MatthiasKunnen/boiler/internal/boiler"
	"github.com/spf13/cobra"
)

var configFilePath string

var rootCmd = &cobra.Command{
	Use:               "boiler",
	Short:             "boiler manages steam games and workshop items on a server",
	DisableAutoGenTag: true,
	Run: func(cmd *cobra.Command, args []string) {
		err := cmd.Help()
		if err != nil {
			log.Fatalf("Error printing help information: %v\n", err)
		}
	},
}

func Execute() error {
	return rootCmd.Execute()
}

func GetCommand() *cobra.Command {
	return rootCmd
}

func init() {
	rootCmd.PersistentFlags().StringVar(
		&configFilePath,
		"config",
		boiler.ConfigFilePath,
		"Path to the config file.",
	)
	rootCmd.AddCommand(logoutCmd)
	rootCmd.AddCommand(updateCmd)
	rootCmd.AddCommand(workshopItemsCmd)
}
