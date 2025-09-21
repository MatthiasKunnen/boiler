package steamcmd

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

type Opts struct {
	// If empty, anonymous will be used.
	LoginUsername         string
	InstallDir            string
	DownloadGames         []DownloadGameOpts
	DownloadWorkshopItems []DownloadWorkshopItemOpts
	// If true, the logged-in user will be logged out at the end of the run.
	Logout       bool
	SteamCmdPath string
}

type DownloadGameOpts struct {
	Id         int
	BetaBranch string
	Validate   bool
}

type DownloadWorkshopItemOpts struct {
	GameId         int
	WorkshopItemId uint64
}

// Exec creates a script for steamcmd to process, writes it to a temporary file, and then
// executes it. The runscript is removed after execution.
// Stdin, stdout, and stderr are connected to their respective stream.
func Exec(ctx context.Context, opts Opts) error {
	if opts.SteamCmdPath == "" {
		return fmt.Errorf("SteamCmdPath not set")
	}
	if opts.InstallDir == "" {
		return fmt.Errorf("InstallDir not set")
	}
	if opts.LoginUsername == "" {
		opts.LoginUsername = "anonymous"
	}

	_, err := os.Stat(opts.InstallDir)
	if err != nil {
		return fmt.Errorf("error checking if InstallDir exists: %w", err)
	}

	temp, err := os.CreateTemp("", "steamcmd_script")
	if err != nil {
		return fmt.Errorf("error creating temp file: %w", err)
	}
	defer os.Remove(temp.Name())
	defer temp.Close()

	_, err = io.WriteString(temp, "@ShutdownOnFailedCommand 1\n")
	if err != nil {
		return err
	}
	_, err = io.WriteString(temp, "force_install_dir "+opts.InstallDir+"\n")
	if err != nil {
		return err
	}
	_, err = io.WriteString(temp, "login "+opts.LoginUsername+"\n")
	if err != nil {
		return err
	}

	var sb strings.Builder
	for _, game := range opts.DownloadGames {
		sb.Reset()
		sb.WriteString("app_update ")
		sb.WriteString(strconv.Itoa(game.Id))
		if game.BetaBranch != "" {
			sb.WriteString(" -beta ")
			sb.WriteString(game.BetaBranch)
		}
		if game.Validate {
			sb.WriteString(" validate")
		}
		sb.WriteString("\n")
		_, err = io.WriteString(temp, sb.String())
		if err != nil {
			return err
		}
	}
	for _, item := range opts.DownloadWorkshopItems {
		_, err = io.WriteString(temp, fmt.Sprintf(
			"workshop_download_item %d %d\n",
			item.GameId,
			item.WorkshopItemId,
		))
		if err != nil {
			return err
		}
	}

	if opts.Logout {
		_, err = io.WriteString(temp, "logout\n")
		if err != nil {
			return err
		}
	}

	_, err = io.WriteString(temp, "quit\n")
	if err != nil {
		return err
	}

	cmd := exec.CommandContext(
		ctx,
		opts.SteamCmdPath,
		"+runscript",
		temp.Name(),
	)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err = cmd.Start()
	if err != nil {
		return fmt.Errorf("error starting steamcmd: %w", err)
	}

	steamCmdErr := cmd.Wait()
	if steamCmdErr == nil || !opts.Logout || opts.LoginUsername == "anonymous" {
		return steamCmdErr
	}

	log.Print("An error occurred and logout may not have occurred, attempting to logout...")
	logoutCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	err = LogOutUser(logoutCtx, opts.SteamCmdPath, opts.LoginUsername)
	if err == nil {
		log.Print("Logout successful")
	} else {
		log.Printf("WARNING: steamcmd exited unexpectedly and attempting logout failed. Your credentials may still be stored on the system. %v", err)
	}

	return steamCmdErr
}

func LogOutUser(ctx context.Context, steamCmdPath string, username string) error {
	logoutTmp, err := os.CreateTemp("", "steamcmd_logout")
	if err != nil {
		return fmt.Errorf("error creating temp file: %w", err)
	}
	defer os.Remove(logoutTmp.Name())
	defer logoutTmp.Close()
	_, err = fmt.Fprintf(logoutTmp, `@ShutdownOnFailedCommand 1
login %s
logout
quit
`, username)
	if err != nil {
		return err
	}

	logoutCmd := exec.CommandContext(
		ctx,
		steamCmdPath,
		"+runscript",
		logoutTmp.Name(),
	)
	return logoutCmd.Run()
}
