package storage

import "errors"

var (
	Errnotfoundresp = errors.New("username have not roots orgs")
	ErrGrisExist    = errors.New("not this organization")
)
