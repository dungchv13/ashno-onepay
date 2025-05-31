package uuid

import (
	"strings"

	"github.com/google/uuid"
)

func NewNoDash() string {
	u4 := uuid.New()
	return strings.ReplaceAll(u4.String(), "-", "")
}

func IsValidUUID(u string) bool {
	_, err := uuid.Parse(u)
	return err == nil
}
