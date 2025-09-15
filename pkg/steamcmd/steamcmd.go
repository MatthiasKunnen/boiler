package steamcmd

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

type Opts struct {
	// If empty, anonymous will be used.
	LoginUsername string
	// If empty, will be prompted on stdin.
	LoginPassword         string
	InstallDir            string
	DownloadGames         []DownloadGameOpts
	DownloadWorkshopItems []DownloadWorkshopItemOpts
	SteamCmdPath          string
}

type DownloadGameOpts struct {
	Id         uint64
	BetaBranch string
	Validate   bool
}

type DownloadWorkshopItemOpts struct {
	GameId         uint64
	WorkshopItemId uint64
}

var steamCmdGuardCodePrompt = []byte("Steam Guard code:")
var steamCmdPasswordPrompt = []byte("password: ")
var steamCmdReady = []byte{
	// Steam> with some ansi escapes
	0x1B, 0x5B, 0x30, 0x6D, 0x1B, 0x5B, 0x31, 0x6D, 0x0A, 0x53, 0x74, 0x65,
	0x61, 0x6D, 0x3E, 0x1B, 0x5B,
}
var largestNoDelimPrompt int

func init() {
	var noDelimPrompts = [][]byte{
		steamCmdGuardCodePrompt,
		steamCmdPasswordPrompt,
		steamCmdReady,
	}
	for _, prompt := range noDelimPrompts {
		if len(prompt) > largestNoDelimPrompt {
			largestNoDelimPrompt = len(prompt)
		}
	}
}

var steamCmdUserLoggedIn = []byte("Waiting for user info...OK")
var steamApiLoaded = []byte{
	0x4C, 0x6F, 0x61, 0x64, 0x69, 0x6E, 0x67, 0x20, 0x53, 0x74, 0x65, 0x61,
	0x6D, 0x20, 0x41, 0x50, 0x49, 0x2E, 0x2E, 0x2E, 0x1B, 0x5B, 0x30, 0x6D,
	0x4F, 0x4B, 0x0A,
}

type steamCmdState int

const (
	steamCmdWaitingForReady steamCmdState = iota
	steamCmdWaitingFor
)

func Exec(opts Opts) error {
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

	commands := make([]string, 0, 4+len(opts.DownloadGames)+len(opts.DownloadWorkshopItems))
	commands = append(
		commands,
		"@ShutdownOnFailedCommand 1",
		"force_install_dir "+opts.InstallDir,
		"login "+opts.LoginUsername,
	)

	var sb strings.Builder
	for _, game := range opts.DownloadGames {
		sb.Reset()
		sb.WriteString("app_update ")
		sb.WriteString(strconv.FormatUint(game.Id, 10))
		if game.BetaBranch != "" {
			sb.WriteString(" -beta ")
			sb.WriteString(game.BetaBranch)
		}
		if game.Validate {
			sb.WriteString(" validate")
		}
		commands = append(commands, sb.String())
	}
	for _, item := range opts.DownloadWorkshopItems {
		commands = append(commands, fmt.Sprintf(
			"download_item %d %d",
			item.GameId,
			item.WorkshopItemId,
		))
	}
	commands = append(commands, "quit")
	commandsIndex := 0

	cmd := exec.Command(opts.SteamCmdPath)

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("could not get stdin pipe: %w", err)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("could not get stdout pipe: %w", err)
	}

	f, err := os.Create("steamcmd.log")
	if err != nil {
		return err
	}
	stdout = io.NopCloser(io.TeeReader(stdout, f))

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("could not get stderr pipe: %w", err)
	}

	err = cmd.Start()
	if err != nil {
		return err
	}

	r := bufio.NewReaderSize(stdout, 4096)
	osStdin := bufio.NewReader(os.Stdin)
	for {
		peekBytes, err := r.Peek(largestNoDelimPrompt)
		log.Printf("Peeked %x", peekBytes)
		if err != nil && err != io.EOF {
			log.Println("Error peeking:", err)
			break
		}

		if bytes.HasPrefix(peekBytes, steamCmdPasswordPrompt) {
			log.Printf("steamcmd: %s", peekBytes)
			_, err := r.Discard(len(steamCmdPasswordPrompt))
			if err != nil {
				return err
			}

			var password string
			if opts.LoginPassword != "" {
				password = opts.LoginPassword
			} else {
				password, err = osStdin.ReadString('\n')
				if err != nil {
					return fmt.Errorf("error reading password from stdin: %w", err)
				}
			}

			err = writeWithNewline(stdin, password)
			if err != nil {
				return err
			}
		} else if bytes.HasPrefix(peekBytes, steamCmdGuardCodePrompt) {
			log.Printf("steamcmd: %s", peekBytes)
			_, err := r.Discard(len(steamCmdGuardCodePrompt))
			if err != nil {
				return err
			}
			code, err := osStdin.ReadString('\n')
			if err != nil {
				return fmt.Errorf("error reading steam guard code from stdin: %w", err)
			}
			err = writeWithNewline(stdin, code)
			if err != nil {
				return err
			}
		} else if bytes.HasPrefix(peekBytes, steamCmdReady) {
			_, err := r.Discard(len(steamCmdReady))
			if err != nil {
				return err
			}
			if commandsIndex >= len(commands) {
				// quit
				break
			}
			command := commands[commandsIndex]
			err = writeWithNewline(stdin, command)
			if err != nil {
				return fmt.Errorf("failed to send command %s: %w", command, err)
			}
			commandsIndex++
		} else {
			line, err := r.ReadBytes('\n')
			if err != nil {
				if err == io.EOF {
					// End of output
					break
				}
				log.Println("Error reading line:", err)
				break
			}
			log.Printf("steamcmd: %s", line)
			log.Printf("steamcmd hex: %x", line)
		}
	}

	go func() {
		s := bufio.NewScanner(stderr)
		for s.Scan() {
			log.Printf("steamcmd stderr: %s\n", s.Text())
		}
		if err := s.Err(); err != nil {
			log.Printf("steamcmd error processing stderr: %v\n", err)
		}
	}()

	return cmd.Wait()
}

func writeWithNewline(w io.Writer, content string) error {
	log.Printf("writing: %s", content)
	_, err := io.WriteString(w, content)
	if err != nil {
		return err
	}
	_, err = w.Write([]byte("\n"))
	if err != nil {
		return fmt.Errorf("error writing newline: %w", err)
	}

	return nil
}
