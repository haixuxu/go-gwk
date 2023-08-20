package utils

import (
	"github.com/google/uuid"
	"strings"
)

func GetUUID() string {
	idstr := uuid.NewString()
	return strings.ReplaceAll(idstr, "-", "")
}
