package repository

type User struct {
	username string
	password string
	balance  int
}

func (u *User) Username() string {
	return u.username
}

func (u *User) Password() string {
	return u.password
}

func (u *User) Balance() int {
	return u.balance
}
