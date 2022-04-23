package tshirt

import (
	"errors"
	"fmt"
)

type Size int

const (
	S Size = iota
	M
	L
	XL
	XXL
	XXXL
	XXXXL
)

var toID = map[string]Size{
	"S":     S,
	"M":     M,
	"L":     L,
	"XL":    XL,
	"XXL":   XXL,
	"XXXL":  XXXL,
	"XXXXL": XXXXL,
}

func Parse(s string) (ret Size, err error) {
	ret, ok := toID[s]
	if !ok {
		err = errors.New(fmt.Sprintf("parse t_shirt size error, not supported size: %s", s))
	}
	return
}
