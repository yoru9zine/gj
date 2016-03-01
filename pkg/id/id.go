package id

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"strings"

	"github.com/pborman/uuid"
)

var (
	ErrDuplicated = errors.New("duplicated")
	ErrNotFound   = errors.New("not found")
)

func New() string {
	i := uuid.NewUUID()
	return fmt.Sprintf("%x", sha256.Sum256(i))
}

func Search(group []string, prefix string) (string, error) {
	cands := []string{}
	for _, v := range group {
		if strings.HasPrefix(v, prefix) {
			cands = append(cands, v)
		}
	}
	if len(cands) == 1 {
		return cands[0], nil
	} else if len(cands) == 0 {
		return "", ErrNotFound
	} else {
		return "", ErrDuplicated
	}
}
