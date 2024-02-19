package models

import "errors"

var ConnectionErrors = [6]error{
	nil, // Connection Accepted (not an error)
	errors.New("connection Refused: unacceptable protocol version"),
	errors.New("connection Refused: identifier rejected"),
	errors.New("connection Refused: server unavailable"),
	errors.New("connection Refused: bad user name or password"),
	errors.New("connection Refused: not authorized"),
}
