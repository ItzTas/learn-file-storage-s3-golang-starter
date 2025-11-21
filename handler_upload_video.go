package main

import (
	"net/http"

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
}
