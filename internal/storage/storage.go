package storage

import "errors"

var (
	ErrGrnotfound = errors.New("graph not found")
	ErrGrisExist  = errors.New("graph exists")
)
