package main

import (
	"log"
	"path"
	"sort"
	"time"

	"github.com/BurntSushi/toml"
)

type config struct {
	PgSQL     configPgsql
	Email     configEmail
	Options   configOptions
	Security  configSecurity
	Users     map[string]configUser
	usersById map[string]configUser
	Scores    map[string]configScoringScheme
}

type configPgsql struct {
	User, Password, Host, Database string
	Port                           int
}

type configEmail struct {
	FromEmail string `toml:"from_email"`
	FromName  string `toml:"from_name"`
	SMTP      configSmtp
}

type configSmtp struct {
	Username string
	Password string
	Server   string
	Port     int
}

type configOptions struct {
	SessionTimeout string `toml:"session_timeout"`
	sessionTimeout time.Duration
}

type configSecurity struct {
	HashKey  string `toml:"hash_key"`
	BlockKey string `toml:"block_key"`
}

type configUser struct {
	Id            string
	Name          string
	Email         string
	Admin         bool
	TimeZone      string `toml:"time_zone"`
	timeZone      *time.Location
	DateFmt       string
	TimeFmt       string
	Friends       []string
	Collaborators []*lcmUser
}

type configScoringScheme struct {
	Order      []string
	Categories map[string]configScoreCategory
}

type configScoreCategory struct {
	Name     string
	Value    int
	Shortcut string
}

func newConfig() (conf config) {
	var err error

	confFile := path.Join(cwd, "config.toml")
	if _, err = toml.DecodeFile(confFile, &conf); err != nil {
		log.Fatalf("Error loading config.toml: %s", err)
	}

	// Check to make sure the session timeout is a valid duration.
	conf.Options.sessionTimeout, err = time.ParseDuration(
		conf.Options.SessionTimeout)
	if err != nil {
		log.Fatalf("Could not parse `session_timeout` '%s' as a duration: %s",
			conf.Options.SessionTimeout, err)
	}

	// And make sure the timeout is at least one minute.
	if conf.Options.sessionTimeout < time.Minute {
		log.Fatalf("Session timeout must be at least 1 minute.")
	}

	// Set the ID of each user.
	for id, user := range conf.Users {
		user.Id = id
		conf.Users[id] = user
	}

	// Now make sure we support each user's time zone.
	for k, user := range conf.Users {
		user.timeZone, err = time.LoadLocation(user.TimeZone)
		if err != nil {
			log.Fatalf("Invalid time zone '%s' for user '%s': %s",
				user.TimeZone, user.Id, err)
		}
		conf.Users[k] = user
	}

	// And now make sure each collaborator is a valid user.
	for k := range conf.Users {
		user := conf.Users[k]
		user.Collaborators = make([]*lcmUser, len(user.Friends))
		for i, friend := range user.Friends {
			if friend == k {
				log.Fatalf("A user '%s' cannot be friends with themself.", k)
			}

			if ufriend, ok := conf.Users[friend]; !ok {
				log.Fatalf("Collaborator '%s' for '%s' is not a valid user.",
					friend, k)
			} else {
				user.Collaborators[i] = newLcmUser(ufriend)
			}
		}
		sort.Sort(usersAlphabetical(user.Collaborators))
		conf.Users[k] = user
	}

	// For faster lookups.
	conf.usersById = make(map[string]configUser, len(conf.Users))
	for _, user := range conf.Users {
		if dupe, ok := conf.usersById[user.Id]; ok {
			log.Fatalf("Two users ('%s' and '%s') cannot have the same name.",
				user.Id, dupe.Id)
		}
		conf.usersById[user.Id] = user
	}

	return
}
