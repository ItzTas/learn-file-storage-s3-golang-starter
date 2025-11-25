package main

import (
	"context"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"os"
	"path"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"

	"github.com/bootdotdev/learn-file-storage-s3-golang-starter/internal/assets"
	"github.com/bootdotdev/learn-file-storage-s3-golang-starter/internal/auth"
	"github.com/bootdotdev/learn-file-storage-s3-golang-starter/internal/videos"
)

func (cfg *apiConfig) handlerUploadVideo(w http.ResponseWriter, r *http.Request) {
	const uploadLimit = 1 << 30
	r.Body = http.MaxBytesReader(w, r.Body, uploadLimit)

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

	userID, err := auth.ValidateJWT(token, cfg.jwtSecret)
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
	mt, newVideo, err := getVideoFromRequest(r)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not get new video", err)
		return
	}
	defer func() {
		err := newVideo.Close()
		if err != nil {
			fmt.Println("Error closing file", err)
		}
	}()

	tmpVideo, err := os.CreateTemp("", "tubely-upload.mp4")
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not create new temporary video", err)
		return
	}
	defer func() {
		err := tmpVideo.Close()
		if err != nil {
			fmt.Println("Error closing file", err)
		}
		err = os.Remove(tmpVideo.Name())
		if err != nil {
			fmt.Println("Error removing file", err)
		}
	}()

	_, err = io.Copy(tmpVideo, newVideo)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not copy new video to tmp file", err)
		return
	}

	_, err = tmpVideo.Seek(0, io.SeekStart)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not reset file pointer", err)
		return
	}

	processedVideoName, err := videos.ProcessVideoForFastStart(tmpVideo.Name())
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not process video for fast start", err)
		return
	}

	key, err := assets.GetAssetsPath(mt)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not generate asset path", err)
		return
	}

	aspectRatio, err := videos.GetVideoAspectRatio(processedVideoName)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not get video aspect ratio", err)
		return
	}

	prefix := "other"

	switch aspectRatio {
	case videos.Aspect169:
		prefix = "landscape"
	case videos.Aspect916:
		prefix = "portrait"
	}

	key = path.Join(prefix, key)

	processedVideo, err := os.Open(processedVideoName)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not open processed video", err)
		return
	}
	defer func() {
		err := processedVideo.Close()
		if err != nil {
			fmt.Println("Error opening processedVideo", err)
		}
	}()

	_, err = cfg.s3Client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket:      &cfg.s3Bucket,
		Key:         aws.String(key),
		Body:        processedVideo,
		ContentType: aws.String(mt),
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not store video object", err)
		return
	}

	newURL := fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", cfg.s3Bucket, cfg.s3Region, key)
	video.VideoURL = &newURL

	err = cfg.db.UpdateVideo(video)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not update video", err)
		return
	}

	respondWithJSON(w, http.StatusOK, video)
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

	return mediaType, file, nil
}
