package users

import "errors"

var ErrNotFound = errors.New("users: not found")
var ErrConflict = errors.New("users: conflict")
