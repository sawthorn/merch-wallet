package repository

type User struct {
	id       int
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

func WithID(id int) UserOpts {
	return func(u *User) {
		u.id = id
	}
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

func (u *User) ID() int {
	return u.id
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
