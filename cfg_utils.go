package main

import (
	"encoding/base64"
	"fmt"
)

func (cfg *apiConfig) makeThumbURL(thumb thumbnail) string {
	data := base64.StdEncoding.EncodeToString(thumb.data)
	return fmt.Sprintf("data:%s;base64,%s", thumb.mediaType, data)
}
