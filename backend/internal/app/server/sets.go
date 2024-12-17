package server

import (
	"errors"
	"net/http"
	"strconv"
	"table-tennis/internal/model"
)

func (s *Server) createSet() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var set model.Set
		if err := s.decode(r, &set); err != nil {
			s.error(w, http.StatusUnprocessableEntity, err, "Ei õnnestunud dekodeerida")
			return
		}

		if err := s.store.Set().Create(set); err != nil {
			s.error(w, http.StatusUnprocessableEntity, err, "Setti koostamisel esines viga")
			return
		}

		s.respond(w, http.StatusOK, Response{
			Message: "Set loodi edukalt",
			Data:    set,
			Error:   nil,
		})
	}
}

func (s *Server) updateSet() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var set model.Set
		if err := s.decode(r, &set); err != nil {
			s.error(w, http.StatusUnprocessableEntity, err, "Ei õnnestunud dekodeerida")
			return
		}

		if err := s.store.Set().Update(set); err != nil {
			s.error(w, http.StatusUnprocessableEntity, err, "Seti uuendamisel tekkis tõrge")
			return
		}

		s.respond(w, http.StatusOK, Response{
			Message: "Seti uuendati edukalt",
			Data:    set,
			Error:   nil,
		})
	}
}

func (s *Server) createSetsMultiple() http.HandlerFunc {
	var responseBody struct {
		Match_id     uint `json:"id"`
		Team_1_score int  `json:"team_1_score"`
		Team_2_score int  `json:"team_2_score"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		if err := s.decode(r, &responseBody); err != nil {
			s.error(w, http.StatusUnprocessableEntity, err, "Ei õnnestunud dekodeerida")
			return
		}

		if err := s.store.Match().CreateSets(responseBody.Match_id, responseBody.Team_1_score, responseBody.Team_2_score); err != nil {
			s.error(w, http.StatusUnprocessableEntity, err, "Mängu settide loomisel tekkis viga")
			return
		}

		s.respond(w, http.StatusOK, Response{
			Message: "Mängule loodi edukalt setid",
			Data:    nil,
			Error:   nil,
		})

	}
}

func (s *Server) getSets() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		mID := r.PathValue("id")
		imID, err := strconv.Atoi(mID)
		var data struct {
			Sets    []model.Set  `json:"sets"`
			Players []model.User `json:"users"`
		}
		if err != nil {
			s.error(w, http.StatusUnprocessableEntity, err, "Vigane ID")
			return
		}
		if imID < 0 {
			s.error(w, http.StatusUnprocessableEntity, errors.New("vigane id"), "Turniiril ei saa olla negatiivne ID")
			return
		}
		sets, err := s.store.Match().GetMatchSets(uint(imID))
		if err != nil {
			s.error(w, http.StatusUnprocessableEntity, err, "Mängude leidmisel tekkis viga")
			return
		}

		players, err := s.store.User().GetPlayers(uint(imID))
		if err != nil {
			s.error(w, http.StatusUnprocessableEntity, err, "Mängijate leidmisel tekkis tõrge")
			return
		}

		data.Sets = sets
		data.Players = players
		s.respond(w, http.StatusOK, Response{
			Message: "Edukalt leiti setid",
			Error:   nil,
			Data:    data,
		})
	}
}
