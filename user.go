package main

import (
	"fmt"
	"log"
	"net/smtp"
)

type lcmUser struct {
	configUser
}

func newLcmUser(confUser configUser) *lcmUser {
	return &lcmUser{confUser}
}

func findUserById(userid string) *lcmUser {
	if len(userid) == 0 {
		panic(e("No user specified."))
	}
	if user, ok := conf.Users[userid]; ok {
		return newLcmUser(user)
	}
	panic(e("Could not find user with id **%s**.", userid))
}

// findUserByNo finds a user by a given user number. We don't panic here
// because we only look for users by number to match things in the DB. If
// we don't find a match, we want to be free to ignore it.
func findUserByNo(userno int) *lcmUser {
	if user, ok := conf.usersById[userno]; ok {
		return newLcmUser(user)
	}
	return nil
}

func (user *lcmUser) String() string {
	return user.Name
}

func (user *lcmUser) valid() bool {
	return user != nil && user.No > 0
}

func (user *lcmUser) lock() {
	locker.lock(user.Id)
}

func (user *lcmUser) unlock() {
	locker.unlock(user.Id)
}

func (user *lcmUser) email(subject, message string) error {
	auth := smtp.PlainAuth(
		"",
		conf.Email.SMTP.Username,
		conf.Email.SMTP.Password,
		conf.Email.SMTP.Server,
	)
	err := smtp.SendMail(
		fmt.Sprintf("%s:%d", conf.Email.SMTP.Server, conf.Email.SMTP.Port),
		auth,
		conf.Email.FromEmail,
		[]string{user.Email},
		[]byte(fmt.Sprintf(`From: "%s" <%s>
To: "%s" <%s>
Subject: %s

%s
`,
			conf.Email.FromName, conf.Email.FromEmail,
			user.Name, user.Email,
			subject, message)))
	if err != nil {
		log.Printf("ERROR [email]: %s", err)
		return err
	}
	return nil
}

type usersAlphabetical []*lcmUser

func (us usersAlphabetical) Less(i, j int) bool {
	return us[i].Name < us[j].Name
}

func (us usersAlphabetical) Swap(i, j int) {
	us[i], us[j] = us[j], us[i]
}

func (us usersAlphabetical) Len() int {
	return len(us)
}
