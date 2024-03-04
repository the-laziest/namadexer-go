package service

import (
	"encoding/hex"
	"strings"
)

func hexToBytes(hash string) ([]byte, error) {
	hash = strings.ToUpper(hash)
	result, err := hex.DecodeString(hash)
	if err != nil {
		return nil, ErrBadRequest
	}
	return result, nil
}
