package main

import (
	"fmt"
	"io"
	"net/http"

	"github.com/google/uuid"

	"github.com/bootdotdev/learn-file-storage-s3-golang-starter/internal/auth"
)

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

	thumb, err := get_thumbnail_from_request(r)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't get thumbnail", err)
		return
	}

	videoThumbnails[videoID] = thumb

	thumbURL := cfg.makeThumbURL(videoID)
	video.ThumbnailURL = &thumbURL

	err = cfg.db.UpdateVideo(video)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't update video url", err)
		return
	}

	respondWithJSON(w, http.StatusOK, video)
}

func get_thumbnail_from_request(r *http.Request) (thumbnail, error) {
	const maxMemory = 10 << 20
	err := r.ParseMultipartForm(maxMemory)
	if err != nil {
		return thumbnail{}, err
	}

	file, header, err := r.FormFile("thumbnail")
	if err != nil {
		return thumbnail{}, err
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return thumbnail{}, err
	}

	mediaType := header.Header.Get("Content-Type")

	thumb := thumbnail{
		mediaType: mediaType,
		data:      data,
	}

	return thumb, nil
}
