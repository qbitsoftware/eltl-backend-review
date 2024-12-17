package server

import (
	"errors"
	"net/http"
	"strconv"
	"table-tennis/internal/model"
)

func (s *Server) createBlog() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var requestBody model.Blog
		if err := s.decode(r, &requestBody); err != nil {
			s.error(w, http.StatusUnprocessableEntity, err, "Vale formaat")
			return
		}
		tournamentID := r.PathValue("id")
		iTournamentID, err := strconv.Atoi(tournamentID)
		if err != nil {
			s.error(w, http.StatusUnprocessableEntity, err, "Vigane ID")
			return
		}
		if iTournamentID < 0 {
			s.error(w, http.StatusUnprocessableEntity, errors.New("turniiril ei saa olla negatiivne ID"), "Vigane ID")
		}
		userID, ok := r.Context().Value(ctxUserID).(uint)
		if !ok {
			s.error(w, http.StatusUnauthorized, errors.New("volitamata"), "Ligipääs on keelatud")
			return
		}
		requestBody.TournamentID = uint(iTournamentID)
		requestBody.AuthorID = userID
		if err := s.store.User().CreateBlog(requestBody); err != nil {
			s.error(w, http.StatusUnprocessableEntity, err, "Blogi koostamisel tekkis viga")
			return
		}
		s.respond(w, http.StatusOK, Response{
			Message: "Edukalt loodud blogi",
			Data:    nil,
			Error:   nil,
		})
	}
}

func (s *Server) getBlogs() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tournamentID := r.PathValue("id")
		iTournamentID, err := strconv.Atoi(tournamentID)
		if err != nil {
			s.error(w, http.StatusUnprocessableEntity, err, "Vigane ID")
			return
		}

		blogs, err := s.store.User().GetTournamentBlog(uint(iTournamentID))
		if err != nil {
			s.error(w, http.StatusUnprocessableEntity, err, "Blogide saamisel teekis viga")
			return
		}

		s.respond(w, http.StatusOK, Response{
			Message: "Edukalt leitud blogid",
			Data:    blogs,
			Error:   nil,
		})
	}
}
