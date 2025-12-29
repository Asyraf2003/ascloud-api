package ports

import "errors"

var ErrAccountEmailTaken = errors.New("account email already exists")
