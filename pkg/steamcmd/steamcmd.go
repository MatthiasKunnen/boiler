package steamcmd

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
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
	// Workshop items will be installed in this directory.
	WorkshopInstallDir string
}

type DownloadGameOpts struct {
	Id         int
	BetaBranch string
	// Used as the subdirectory to download the files to.
	Name     string
	Validate bool
}

type DownloadWorkshopItemOpts struct {
	GameId         int
	WorkshopItemId uint64
}

// Exec executes the given steamcmd Opts.
// Depending on the options specified, steamcmd will be run multiple times.
// This is because games need to be installed in their own directories and force_install_dir
// after login is discouraged.
func Exec(ctx context.Context, opts Opts) error {
	r := newRunner(
		withLogout(opts.Logout),
		withSteamcmd(opts.SteamCmdPath),
	)

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
		err := r.Add(runOpts{
			LoginUsername: opts.LoginUsername,
			InstallDir:    filepath.Join(opts.InstallDir, game.Name),
			ExtraCommands: []string{sb.String()},
		})
		if err != nil {
			return err
		}
	}

	var workshopDownloadCommands []string
	for _, item := range opts.DownloadWorkshopItems {
		workshopDownloadCommands = append(workshopDownloadCommands, fmt.Sprintf(
			// @todo this does not error on exit, unlike download_item though download_item
			//       downloads to ~/.steam/steamcmd/linux32\steamapps\content\app_APP_ID\item_ID
			"workshop_download_item %d %d",
			item.GameId,
			item.WorkshopItemId,
		))
	}

	if len(workshopDownloadCommands) > 0 {
		err := r.Add(runOpts{
			LoginUsername: opts.LoginUsername,
			InstallDir:    opts.WorkshopInstallDir,
			ExtraCommands: workshopDownloadCommands,
		})
		if err != nil {
			return err
		}
	}

	return r.Run(ctx)
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

type runner struct {
	todo         []runOpts
	steamCmdPath string
	// If true, the logged-in user will be logged out at the end of the last run.
	logout bool
}

type runOpts struct {
	// If empty, anonymous will be used.
	LoginUsername string
	// If true, the logged-in user will be logged out at the end of the run.
	Logout        bool
	InstallDir    string
	ExtraCommands []string
}

type runnerOpt func(*runner)

func withLogout(logout bool) runnerOpt {
	return func(r *runner) {
		r.logout = logout
	}
}
func withSteamcmd(steamcmd string) runnerOpt {
	return func(r *runner) {
		r.steamCmdPath = steamcmd
	}
}

func newRunner(opts ...runnerOpt) *runner {
	r := &runner{
		steamCmdPath: "steamcmd",
	}
	for _, opt := range opts {
		opt(r)
	}

	return r
}

func (r *runner) Add(opts runOpts) error {
	if opts.InstallDir == "" {
		return fmt.Errorf("InstallDir not set")
	}

	r.todo = append(r.todo, opts)
	return nil
}

func (r *runner) Run(ctx context.Context) error {
	var resultErr error
	logOutUsers := make(map[string]struct{})
	for _, conf := range r.todo {
		resultErr = r.runSingle(ctx, conf)
		if conf.LoginUsername != "" && (r.logout || (resultErr != nil && conf.Logout)) {
			logOutUsers[conf.LoginUsername] = struct{}{}
		}
		if resultErr != nil {
			break
		}
	}

	if len(logOutUsers) == 0 {
		return resultErr
	}

	for username, _ := range logOutUsers {
		err := LogOutUser(ctx, r.steamCmdPath, username)
		if err != nil {
			log.Printf("WARNING: steamcmd exited unexpectedly while attempting to logout. Your credentials may still be stored on the system. %v", err)
		}
	}

	return resultErr
}

// runSingle creates a script for steamcmd to process, writes it to a temporary file, and then
// executes it. The runscript is removed after execution.
// Stdin, stdout, and stderr are connected to their respective stream.
func (r *runner) runSingle(ctx context.Context, opts runOpts) error {
	username := opts.LoginUsername
	if username == "" {
		username = "anonymous"
	}

	_, err := os.Stat(filepath.Dir(opts.InstallDir))
	if err != nil {
		return fmt.Errorf("error checking if InstallDir's parent dir exists: %w", err)
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
	_, err = io.WriteString(temp, "login "+username+"\n")
	if err != nil {
		return err
	}

	for _, command := range opts.ExtraCommands {
		_, err = io.WriteString(temp, command+"\n")
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
		r.steamCmdPath,
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

	return cmd.Wait()
}
