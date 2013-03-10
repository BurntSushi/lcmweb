package main

type config struct {
	MySQL configMysql
}

type configMysql struct {
	User, Password, Host, Database string
	Port                           int
}
