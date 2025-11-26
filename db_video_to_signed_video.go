package main

import (
	"strings"
	"time"

	"github.com/bootdotdev/learn-file-storage-s3-golang-starter/internal/database"
	"github.com/bootdotdev/learn-file-storage-s3-golang-starter/internal/s3internal"
)

func (cfg *apiConfig) dbVideoToSignedVideo(video database.Video) (database.Video, error) {
	splits := strings.Split(*video.VideoURL, ",")
	bucket, key := splits[0], splits[1]

	expireTime := 1 * time.Minute
	url, err := s3internal.GeneratePresignedURL(cfg.s3Client, bucket, key, expireTime)
	if err != nil {
		return database.Video{}, err
	}
	video.VideoURL = &url
	return video, nil
}
