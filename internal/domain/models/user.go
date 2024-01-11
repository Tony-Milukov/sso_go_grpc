package models

type User struct {
	Email    string
	Password string
	UserId   uint64
	Username string
	Roles    []*Role
}
