package consts

import (
	"errors"
	"time"
)

const (
	DB_TIMEOUT    = 10 * time.Second
	CTX_TIMEOUT   = 10 * time.Second
	READ_TIMEOUT  = 10 * time.Second
	WRITE_TIMEOUT = 30 * time.Second
	IDLE_TIMEOUT  = 1 * time.Minute
)

var (
	NOT_FOUND = errors.New("not found")
)
