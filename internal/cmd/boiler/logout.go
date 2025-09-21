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

var logoutCmd = &cobra.Command{
	Use:   "logout [username]",
	Short: "Logs the given user out of steamcmd",
	Args:  cobra.RangeArgs(0, 1),
	Run: func(cmd *cobra.Command, args []string) {
		var username string
		if len(args) > 0 {
			username = args[0]
		}

		stopSig := make(chan os.Signal, 1)
		signal.Notify(stopSig, os.Interrupt, syscall.SIGTERM)
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		go func() {
			<-stopSig
			cancel()
		}()

		configFile, err := os.Open(configFilePath)
		if err != nil {
			log.Fatalf("failed to open config file: %v", err)
		}
		defer configFile.Close()
		b, err := boiler.FromConfigReader(configFile, boiler.WithLoginUsername(username))
		if err != nil {
			log.Fatalf("failed to read config: %v", err)
		}

		err = b.Logout(ctx)
		if err != nil {
			log.Fatalf("failed to logout: %v", err)
		}
		log.Printf("Successfully logged out")
	},
}
