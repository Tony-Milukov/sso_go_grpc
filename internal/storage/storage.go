package storage

import "errors"

var (
	ErrUserExists        = errors.New("user with that email or username already exists")
	ErrAuth              = errors.New("authentication Failed; Login or Password is incorrect")
	ErrNoPermission      = errors.New("terminated; No permission for this operation")
	ErrRoleExists        = errors.New("role with that name already exists")
	ErrRoleNotExists     = errors.New("this role do not exist")
	ErrUserNotExists     = errors.New("user with that email or username does not exist")
	ErrEmptyValue        = errors.New("impossible to update with empty value")
	ErrNoDelete          = errors.New("there was no data to delete")
	ErrUserAndRoleIvalid = errors.New("error on checking user id or role ids")
)
