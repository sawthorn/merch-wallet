package repository

type User struct {
	username string
	password string
	balance  int
}

type UserOpts func(*User)

func NewUser(opts ...UserOpts) *User {
	u := &User{}
	for _, opt := range opts {
		opt(u)
	}

	return u
}

func WithUsername(username string) UserOpts {
	return func(u *User) {
		u.username = username
	}
}

func WithPassword(password string) UserOpts {
	return func(u *User) {
		u.password = password
	}
}

func WithBalance(balance uint) UserOpts {
	return func(u *User) {
		u.balance = int(balance)
	}
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
