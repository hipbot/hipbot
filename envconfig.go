package hipbot

import (
	"fmt"
	"os"
	"strings"
)

const (
	// EnvJabberID is the env variable for the JabberID
	EnvJabberID = "HIPCHAT_JABBERID"
	// EnvHost is the env variable for the Host
	EnvHost = "HIPCHAT_HOST"
	// EnvPassword is the env variable for the password
	EnvPassword = "HIPCHAT_PASSWORD"
	// EnvNick is the env variable for the nick name (short name)
	EnvNick = "HIPCHAT_NICK"
	// EnvRooms is the env variable for a comma separated list of rooms to join
	EnvRooms = "HIPCHAT_ROOMS"
	// EnvDebug is the env variable that if set causes XMPP messages to be printed on stdout
	EnvDebug = "HIPCHAT_DEBUG"
	// EnvFullName is the env variable for the full name of the bot account
	EnvFullName = "HIPCHAT_FULLNAME"
)

// EnvConfig builds a configuration from environment variables.
// See the constants section for variable names
func EnvConfig() (Config, error) {
	cfg := Config{
		JabberID: os.Getenv(EnvJabberID),
		Host:     os.Getenv(EnvHost),
		Password: os.Getenv(EnvPassword),
		Nick:     os.Getenv(EnvNick),
		FullName: os.Getenv(EnvFullName),
		Rooms:    strings.Split(os.Getenv(EnvRooms), ","),
	}
	if cfg.JabberID == "" || cfg.Host == "" || cfg.Password == "" || cfg.Nick == "" {
		return cfg, fmt.Errorf("missing env vars required - %s %s %s %s optional - %s %s", EnvJabberID, EnvHost, EnvPassword, EnvNick, EnvRooms, EnvFullName)
	}
	_, debug := os.LookupEnv(EnvDebug)
	cfg.Debug = debug
	return cfg, nil
}
