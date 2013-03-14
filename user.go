package main

import (
	"fmt"
	"log"
	"net/smtp"
)

func findUser(userid string) configUser {
	if len(userid) == 0 {
		panic(e("No user specified."))
	}
	if user, ok := conf.Users[userid]; ok {
		return user
	}
	panic(e("Could not find user with id **%s**.", userid))
}

func (user configUser) valid() bool {
	return user.No > 0
}

func (user configUser) email(subject, message string) error {
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
