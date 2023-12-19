package storage

import "errors"

var (
	ErrUserExists    = errors.New("User with that email or username already exists")
	ErrAuth          = errors.New("Authentication Failed; Login or Password is incorrect")
	ErrNoPermission  = errors.New("Terminated; No permission for this operation")
	ErrRoleExists    = errors.New("Role with that name already exists")
	ErrUserNotExists = errors.New("User with that email or username does not exist")
)
