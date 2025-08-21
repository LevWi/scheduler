package common

import "errors"

var ErrNotFound = errors.New("not found")
var ErrUnauthorized = errors.New("unauthorized")
var ErrNotAllowed = errors.New("not allowed")

var ErrInternal = errors.New("internal error")
var ErrInvalidArgument = errors.New("invalid argument")
