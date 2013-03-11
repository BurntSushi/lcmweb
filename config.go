package main

type config struct {
	PgSQL  configPgsql
	Users  map[string]configUser
	Scores map[string]configScoringScheme
}

type configPgsql struct {
	User, Password, Host, Database string
	Port                           int
}

type configUser struct {
	Id    int
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
