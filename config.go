package main

import (
	"time"
)

type config struct {
	PgSQL    configPgsql
	Email    configEmail
	Options  configOptions
	Security configSecurity
	Users    map[string]configUser
	Scores   map[string]configScoringScheme
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
	No    int
	Id    string
	Name  string
	Email string
	Admin bool
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
