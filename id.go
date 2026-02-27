package main

import (
	"strings"

	"github.com/google/uuid"
)

func generateID(prefix string) string {
	id := uuid.New()
	short := strings.ReplaceAll(id.String(), "-", "")
	return prefix + "_" + short[:12]
}
