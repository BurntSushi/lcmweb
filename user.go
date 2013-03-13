package main

func findUser(userid string) configUser {
	if user, ok := conf.Users[userid]; ok {
		return user
	}
	panic(e("Could not find user with id **%s**.", userid))
}

func (user configUser) valid() bool {
	return user.No > 0
}
