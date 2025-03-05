package main

import (
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/bootdotdev/learn-file-storage-s3-golang-starter/internal/auth"
	"github.com/bootdotdev/learn-file-storage-s3-golang-starter/internal/database"
	"github.com/google/uuid"
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

	const maxMemory = 10 << 20
	r.ParseMultipartForm(maxMemory)

	// "thumbnail" should match the HTML form input name
	file, header, err := r.FormFile("thumbnail")
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Unable to parse form file", err)
		return
	}

	fileHeader := header.Header.Get("Content-Type")

	fileData, err := io.ReadAll(file)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Unable to parse form file", err)
		return
	}

	defer file.Close()

	videosDb, err := cfg.db.GetVideo(videoID)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "User is not the owner of the video", err)
		return
	}

	newVideoTN := thumbnail{
		data:      fileData,
		mediaType: fileHeader,
	}

	videoThumbnails[videosDb.ID] = newVideoTN
	tnURL := "http://localhost:" + cfg.port + "/api/thumbnails/" + videoID.String()

	dbVideoUpload := database.Video{
		ID:           videosDb.ID,
		CreatedAt:    videosDb.CreatedAt,
		UpdatedAt:    time.Now(),
		ThumbnailURL: &tnURL,
		VideoURL:     videosDb.VideoURL,
		CreateVideoParams: database.CreateVideoParams{
			Title:       videosDb.Title,
			Description: videosDb.Description,
			UserID:      userID,
		},
	}

	err = cfg.db.UpdateVideo(dbVideoUpload)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Contact Admin", err)
		return
	}

	respondWithJSON(w, http.StatusOK, dbVideoUpload)
}
