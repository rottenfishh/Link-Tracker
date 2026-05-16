package model

import "errors"

var (
	ErrInvalidRequest     = errors.New("invalid request")
	ErrNotFound           = errors.New("resource not found")
	ErrLinkAlreadyTracked = errors.New("link already tracked")
	ErrInternalServer     = errors.New("internal server error")
)
