package domain

import "errors"

// ErrExists - object with these credentials already exists.
var ErrExists = errors.New("object already exists")

// ErrNotFound - object with these credentials does not exist.
var ErrNotFound = errors.New("object not exists")

var ErrNotCreated = errors.New("could not create object")

var ErrNoHeader = errors.New("authorization header not found")

var ErrInvalidHeader = errors.New("invalid authorization header")

var ErrInvalidToken = errors.New("invalid token")

var ErrExpiredToken = errors.New("expired token")

var ErrTokenNotCreated = errors.New("failed to create token")
