package main

import (
	"fmt"

	"github.com/google/uuid"
)

func (cfg *apiConfig) makeThumbURL(videoID uuid.UUID) string {
	return fmt.Sprintf("http://localhost:%s/api/thumbnails/%s", cfg.port, videoID.String())
}
