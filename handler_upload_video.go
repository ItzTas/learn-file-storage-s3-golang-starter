package main

import (
	"mime"
	"mime/multipart"
	"net/http"
	"os"

	"github.com/bootdotdev/learn-file-storage-s3-golang-starter/internal/auth"
	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerUploadVideo(w http.ResponseWriter, r *http.Request) {
	const uploadLimit = 1 << 30
	videoIDString := r.PathValue("videoID")
	videoID, err := uuid.Parse(videoIDString)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not parse video id", err)
		return
	}
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Could not get bearer token", err)
		return
	}

	userID, err := auth.ValidateJWT(token)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Could not validade user", err)
		return
	}

	video, err := cfg.db.GetVideo(videoID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not get video", err)
		return
	}
	if video.UserID != userID {
		respondWithError(w, http.StatusUnauthorized, "Not authorized to update this video", nil)
		return
	}
	_, newVideo, err := getVideoFromRequest(r)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not get new video", err)
		return
	}
	defer newVideo.Close()
	os.CreateTemp("", "tubely-upload.mp4")
}

func getVideoFromRequest(r *http.Request) (string, multipart.File, error) {
	const maxMemory = 10 << 20
	err := r.ParseMultipartForm(maxMemory)
	if err != nil {
		return "", nil, err
	}

	file, header, err := r.FormFile("video")
	if err != nil {
		return "", nil, err
	}

	mediaType := header.Header.Get("Content-Type")
	mt, _, err := mime.ParseMediaType(mediaType)
	if err != nil {
		return "", nil, err
	}
	allowedMediaTypes := map[string]struct{}{
		"video/mp4": {},
		"video/mpv": {},
	}
	if _, ok := allowedMediaTypes[mt]; !ok {
		return "", nil, unallowedMediaType
	}

	return mt, file, nil
}
