package server

import (
	"context"
	"net/http"
	"os"
)

func (s *Server) imageUpload() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tournament_id := r.PathValue("tournament_id")
		day_number := r.PathValue("day")

		err := r.ParseMultipartForm(20 << 20)
		if err != nil {
			s.error(w, http.StatusBadRequest, err, "Midagi läks valesti")
			return
		}

		//tournament images
		fileHeaders := r.MultipartForm.File["images"]

		for _, fileHeader := range fileHeaders {
			// Open uploaded file
			file, err := fileHeader.Open()
			if err != nil {
				s.error(w, http.StatusBadRequest, err, "Midagi läks valesti")
				return
			}
			defer file.Close()

			// Check file content type
			contentType := fileHeader.Header.Get("Content-Type")
			if contentType != "image/jpeg" && contentType != "image/png" && contentType != "image/gif" {
				s.error(w, http.StatusBadRequest, err, "Midagi läks valesti")
				return
			}
			if err := s.awsS3.UploadFile(context.Background(), os.Getenv(awsUploadBucket), "gallery/"+tournament_id+"/"+day_number+"/"+fileHeader.Filename, file); err != nil {
				s.error(w, http.StatusBadRequest, err, "Midagi läks valesti")
				return
			}
		}

		s.respond(w, http.StatusOK, Response{
			Message: "Successfully uploaded to S3",
			Data:    nil,
			Error:   nil,
		})
	}
}

func (s *Server) getBucketImages() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tournament_id := r.PathValue("id")
		day_number := r.PathValue("day")
		files, err := s.awsS3.ListObjects(context.Background(), "gallery/thumbnails/"+tournament_id+"/"+day_number+"/")
		if err != nil {
			s.error(w, http.StatusUnprocessableEntity, err, "something went wront")
			return
		}
		var output []string
		for _, f := range files {
			output = append(output, os.Getenv(awsRetrieveBucketUrl)+*f.Key)
		}
		s.respond(w, http.StatusOK, Response{
			Message: "Successful",
			Data:    output,
			Error:   nil,
		})
	}
}

func (s *Server) imageRemove() http.HandlerFunc {
	var requestBody struct {
		ImageKey string `json:"image_key"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		if err := s.decode(r, &requestBody); err != nil {
			s.error(w, http.StatusUnprocessableEntity, err, "Midagi läks valesti")
			return
		}
		b, err := s.awsS3.DeleteObject(context.Background(), requestBody.ImageKey, "", false)
		if err != nil {
			s.error(w, http.StatusBadRequest, err, "Pildi kustutamisel tekkis viga")
			return
		}

		s.respond(w, http.StatusOK, Response{
			Message: "Pilt kustutati edukalt",
			Data:    b,
			Error:   nil,
		})
	}
}
