package common

import "errors"

var ErrNotFound = errors.New("not found")
var ErrUnauthorized = errors.New("unauthorized")

var ErrInternal = errors.New("internal error")
var ErrInvalidArgument = errors.New("invalid argument")
