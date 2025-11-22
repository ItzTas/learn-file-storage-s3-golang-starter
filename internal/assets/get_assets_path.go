package assets

import (
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/bootdotdev/learn-file-storage-s3-golang-starter/internal/utils"
)

func GetAssetsPath(contentType string) (string, error) {
	contentTypeToExt := map[string]string{
		"video/mp4": "mp4",
		"image/png": "png",
	}

	ext, ok := contentTypeToExt[contentType]
	if !ok {
		return getAssetsPathFromAllowed(contentType)
	}

	return generateHexFromExtension(ext)
}

func getAssetsPathFromAllowed(contentType string) (string, error) {
	contentType = strings.TrimPrefix(contentType, ".")
	allowedContentTypes := map[string]struct{}{
		"mp4": {},
		"png": {},
	}
	if _, ok := allowedContentTypes[contentType]; !ok {
		return "", fmt.Errorf("invalid content type, content types: %s", allowedContentTypes)
	}
	return generateHexFromExtension(contentType)
}

func generateHexFromExtension(ext string) (string, error) {
	randBytes, err := utils.GenerateRandomBytes(16)
	if err != nil {
		return "", err
	}
	hexStr := hex.EncodeToString(randBytes)
	return fmt.Sprintf("%s.%s", hexStr, ext), nil
}
