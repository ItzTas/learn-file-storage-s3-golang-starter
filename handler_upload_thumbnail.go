package main

import (
	"errors"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"

	"github.com/google/uuid"

	"github.com/bootdotdev/learn-file-storage-s3-golang-starter/internal/auth"
)

var unallowedMediaType = errors.New("Media type not allowed")

func (cfg *apiConfig) handlerUploadThumbnail(w http.ResponseWriter, r *http.Request) {
	videoIDString := r.PathValue("videoID")
	videoID, err := uuid.Parse(videoIDString)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid ID", err)
		return
	}

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't find JWT", err)
		return
	}

	userID, err := auth.ValidateJWT(token, cfg.jwtSecret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't validate JWT", err)
		return
	}

	fmt.Println("uploading thumbnail for video", videoID, "by user", userID)

	video, err := cfg.db.GetVideo(videoID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't get video metadata", err)
		return
	}

	if video.UserID != userID {
		respondWithError(w, http.StatusUnauthorized, "Invalid user id", nil)
		return
	}

	mediaType, thumbFile, err := getThumbnailFromRequest(r)
	if err != nil {
		if errors.Is(err, unallowedMediaType) {
			respondWithError(w, http.StatusInternalServerError, fmt.Sprintf("Couldn't get thumbnail: %s", err.Error()), err)
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Couldn't get thumbnail", err)
		return
	}
	defer thumbFile.Close()

	exts, err := mime.ExtensionsByType(mediaType)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't get media type", err)
		return
	}
	mediaTypeExt := exts[0]
	thumbPath := filepath.Join(cfg.assetsRoot, fmt.Sprintf("%s%s", videoID.String(), mediaTypeExt))
	if _, err := os.Stat(thumbPath); err == nil {
		if err := os.Remove(thumbPath); err != nil {
			respondWithError(w, http.StatusInternalServerError, "Couldn't remove existing thumb file", err)
			return
		}
	} else if !os.IsNotExist(err) {
		respondWithError(w, http.StatusInternalServerError, "Error checking existing thumb file", err)
		return
	}

	localThumbFile, err := os.Create(thumbPath)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't create thumb file", err)
		return
	}
	defer localThumbFile.Close()

	_, err = io.Copy(localThumbFile, thumbFile)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't write to thumb file", err)
		return
	}
	thumbURL := fmt.Sprintf("http://localhost:%s/assets/%s%s", cfg.port, videoID.String(), mediaTypeExt)
	video.ThumbnailURL = &thumbURL

	if err = cfg.db.UpdateVideo(video); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't update video url", err)
		return
	}

	respondWithJSON(w, http.StatusOK, video)
}

func getThumbnailFromRequest(r *http.Request) (string, multipart.File, error) {
	const maxMemory = 10 << 20
	err := r.ParseMultipartForm(maxMemory)
	if err != nil {
		return "", nil, err
	}

	file, header, err := r.FormFile("thumbnail")
	if err != nil {
		return "", nil, err
	}

	mediaType := header.Header.Get("Content-Type")
	mt, _, err := mime.ParseMediaType(mediaType)
	if err != nil {
		return "", nil, err
	}
	allowedMediaTypes := map[string]struct{}{
		"image/jpeg": {},
		"image/png":  {},
	}
	if _, ok := allowedMediaTypes[mt]; !ok {
		return "", nil, unallowedMediaType
	}

	return mediaType, file, nil
}
