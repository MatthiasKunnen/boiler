package boiler

type Config struct {
	// Points to the path containing the [Database] JSON config.
	DatabasePath string
	// Points to the path containing the [GamesConfig].
	GamesConfPath string
	// The path where games and workshop items will be downloaded.
	GamesDir string
	// Username used to log in with steamcmd.
	LoginUsername string
	SteamCmdPath  string
}
