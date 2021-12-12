package domain

import "errors"

var ErrTransReceiver = errors.New("invalid transaction receiver")

var ErrTransSender = errors.New("invalid transaction sender")

var ErrTransSum = errors.New("minimum sum for transaction should be higher than 150")

var ErrInvalidHeader = errors.New("invalid authorization header")

var ErrInvalidToken = errors.New("invalid token")

var ErrExpiredToken = errors.New("expired token")

var ErrNoHeader = errors.New("authorization header not found")
