package boiler

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/MatthiasKunnen/boiler/internal/boiler"
	"github.com/spf13/cobra"
)

var downloadUpToDate bool
var loginUsername string
var logout bool
var skipDatabaseUpdate bool
var skipDownload bool

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Updates all games, collections and workshop items in the games configuration",
	Long: `By default, this downloads all games, and fetches information of the collections and
workshop items in the games.json or dependencies thereof. All out-of-date workshop items that
are in games.json or are dependencies will be downloaded.
`,
	Args: cobra.NoArgs,
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

		stopSig := make(chan os.Signal, 1)
		signal.Notify(stopSig, os.Interrupt, syscall.SIGTERM)
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		go func() {
			<-stopSig
			cancel()
		}()

		if !skipDatabaseUpdate {
			err = b.UpdateDatabase(ctx, boiler.UpdateOpts{})
			if err != nil {
				log.Fatalf("failed to update: %v", err)
			}
		}
		if !skipDownload {
			err = b.Download(ctx, boiler.DownloadOpts{
				DownloadUpToDate: downloadUpToDate,
				LoginUsername:    loginUsername,
				Logout:           logout,
			})
			if err != nil {
				log.Fatalf("failed to update: %v", err)
			}
		}
		log.Println("Update successful")
	},
}

func init() {
	updateCmd.Flags().StringVar(
		&loginUsername,
		"login-username",
		"",
		`Username to use to log in with steamcmd.`,
	)
	updateCmd.Flags().BoolVar(
		&logout,
		"logout",
		false,
		`Log out of steamcmd after the operation completes.`,
	)
	updateCmd.Flags().BoolVar(
		&skipDatabaseUpdate,
		"skip-database-update",
		false,
		`Do not check if workshop items are up-to-date. Note that unless
--download-up-to-date is specified, only out-of-date workshop items are downloaded.`,
	)
	updateCmd.Flags().BoolVar(
		&skipDownload,
		"skip-download",
		false,
		"Do not download any of the games or workshop items. Only updates the database.",
	)
	updateCmd.Flags().BoolVar(
		&downloadUpToDate,
		"download-up-to-date",
		false,
		`Additionally, force download required workshop items that are up-to-date.`,
	)
}
