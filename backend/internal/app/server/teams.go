package server

import (
	"errors"
	"net/http"
	"strconv"
	"table-tennis/internal/model"
	"table-tennis/pkg/validator"
)

func (s *Server) createTeam() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var requestBody model.Team
		if err := s.decode(r, &requestBody); err != nil {
			s.error(w, http.StatusUnprocessableEntity, err, "Ei õnnestunud dekodeerida")
			return
		}

		tournamentID := r.PathValue("id")
		itournamentID, err := strconv.Atoi(tournamentID)
		if err != nil {
			s.error(w, http.StatusUnprocessableEntity, err, "Vigane ID")
			return
		}

		if itournamentID < 0 {
			s.error(w, http.StatusUnprocessableEntity, errors.New("vigane id"), "Turniiril ei saa olla negatiivne ID")
			return
		}

		requestBody.TournamentID = uint(itournamentID)

		if err := s.store.Club().CreateTeam(requestBody); err != nil {
			s.error(w, http.StatusUnprocessableEntity, err, "Tiimi loomisel tekkis viga")
			return
		}

		s.respond(w, http.StatusOK, Response{
			Data:    nil,
			Message: "Tiim loodi edukalt",
			Error:   nil,
		})
	}
}

func (s *Server) updateTeam() http.HandlerFunc {
	var requestBody model.Team
	return func(w http.ResponseWriter, r *http.Request) {
		if err := s.decode(r, &requestBody); err != nil {
			s.error(w, http.StatusUnprocessableEntity, err, "Ei õnnestunud dekodeerida")
			return
		}

		tournamentID := r.PathValue("id")
		itournamentID, err := strconv.Atoi(tournamentID)
		if err != nil {
			s.error(w, http.StatusUnprocessableEntity, err, "Vigane ID")
			return
		}

		if itournamentID < 0 {
			s.error(w, http.StatusUnprocessableEntity, errors.New("invalid id"), "Turniiril ei saa olla negatiivne ID")
			return
		}

		requestBody.TournamentID = uint(itournamentID)
		matchID := r.PathValue("team_id")
		iMatchID, err := strconv.Atoi(matchID)
		if err != nil {
			s.error(w, http.StatusUnprocessableEntity, err, "Vigane ID")
			return
		}

		requestBody.ID = uint(iMatchID)

		if err := s.store.Club().UpdateTeam(requestBody); err != nil {
			s.error(w, http.StatusUnprocessableEntity, err, "Tiimi uuendamisel tekkis tõrge")
			return
		}

		s.respond(w, http.StatusOK, Response{
			Data:    nil,
			Message: "Tiimi uuendatud edukalt",
			Error:   nil,
		})
	}
}

func (s *Server) getTeams() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tournamentID := r.PathValue("id")
		itournamentID, err := strconv.Atoi(tournamentID)
		if err != nil {
			s.error(w, http.StatusUnprocessableEntity, err, "Vigane ID")
			return
		}

		if itournamentID < 0 {
			s.error(w, http.StatusUnprocessableEntity, errors.New("vigane id"), "Turniiril ei saa olla negatiivne ID")
			return
		}

		teams, err := s.store.Club().GetTournamentTeams(uint(itournamentID))
		if err != nil {
			s.error(w, http.StatusUnprocessableEntity, err, "Tiime ei suudetud leida")
			return
		}

		s.respond(w, http.StatusOK, Response{
			Message: "Edukalt leiti kõik tiimid",
			Data:    teams,
			Error:   nil,
		})

	}
}

func (s *Server) getTeam() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		matchID := r.PathValue("team_id")
		iMatchID, err := strconv.Atoi(matchID)
		if err != nil {
			s.error(w, http.StatusUnprocessableEntity, err, "Vigane ID")
			return
		}
		//call a function
		team, err := s.store.Club().GetTournamentTeam(uint(iMatchID))
		if err != nil {
			s.error(w, http.StatusUnprocessableEntity, err, "Vigane ID")
			return
		}

		s.respond(w, http.StatusOK, Response{
			Message: "Edukalt leiti kõik tiimi mängijad",
			Data:    team,
			Error:   nil,
		})

	}
}

func (s *Server) getTeamMatch() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		matchID := r.PathValue("match_id")
		iMatchID, err := strconv.Atoi(matchID)
		if err != nil {
			s.error(w, http.StatusUnprocessableEntity, err, "Vigane ID")
			return
		}
		team, err := s.store.Club().GetTeamMatch(uint(iMatchID))
		if err != nil {
			s.error(w, http.StatusUnprocessableEntity, err, "Tiimi ei leitud")
			return
		}

		s.respond(w, http.StatusOK, Response{
			Message: "Edukalt leiti tiim",
			Data:    team,
			Error:   nil,
		})
	}
}

func (s *Server) updateTime() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var requestBody []model.ChangeClock
		if err := s.decode(r, &requestBody); err != nil {
			s.error(w, http.StatusUnprocessableEntity, err, "Ei õnnestunud dekodeerida")
			return
		}

		if err := s.store.Tournament().ChangeMatchTime(requestBody); err != nil {
			s.error(w, http.StatusUnprocessableEntity, err, "Aja muutmine ebaõnnestus")
			return
		}

		s.respond(w, http.StatusOK, Response{
			Message: "Edukalt muudeti aega",
			Data:    nil,
			Error:   nil,
		})

	}
}

func (s *Server) createTeamMatch() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var requestBody model.TeamMatch
		if err := s.decode(r, &requestBody); err != nil {
			s.error(w, http.StatusUnprocessableEntity, err, "Ei õnnestunud dekodeerida")
			return
		}

		if err := validator.Validate(requestBody); err != nil {
			s.error(w, http.StatusUnprocessableEntity, err, "Vale info!")
			return
		}

		response, err := s.store.Club().CreateTeamMatch(requestBody)
		if err != nil {
			s.error(w, http.StatusUnprocessableEntity, err, "Tiimi mängu loomisel tekkis viga")
			return
		}
		s.respond(w, http.StatusOK, Response{
			Message: "Tiimi mäng loodi edukalt",
			Data:    response,
			Error:   nil,
		})
	}
}

func (s *Server) deleteTeamMatch() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		team_match_id := r.PathValue("team_match_id")
		i_team_match_id, err := strconv.Atoi(team_match_id)
		if err != nil {
			s.error(w, http.StatusUnprocessableEntity, err, "Vigane ID")
			return
		}
		if err := s.store.Club().DeleteTeamMatch(uint(i_team_match_id)); err != nil {
			s.error(w, http.StatusUnprocessableEntity, err, "Viga mängu resettimisel")
			return
		}
		s.respond(w, http.StatusOK, Response{
			Message: "Mäng kustutati edukalt",
			Data:    nil,
			Error:   nil,
		})
	
	}
}

func (s *Server) updateTeamMatch() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var requestBody model.TeamMatch
		if err := s.decode(r, &requestBody); err != nil {
			s.error(w, http.StatusUnprocessableEntity, err, "Ei õnnestunud dekodeerida")
			return
		}

		id := r.PathValue("id")
		iteamMatch, err := strconv.Atoi(id)
		if err != nil {
			s.error(w, http.StatusUnprocessableEntity, err, "Vigane ID")
			return
		}

		requestBody.ID = uint(iteamMatch)

		if err := s.store.Club().UpdateTeamMatch(requestBody); err != nil {
			s.error(w, http.StatusUnprocessableEntity, err, "Tiimi mängu uuendamisel tekkis viga")
			return
		}
		s.respond(w, http.StatusOK, Response{
			Message: "Edukalt uuendati tiimi mängu",
			Data:    nil,
			Error:   nil,
		})
	}
}

func (s *Server) deleteTeam() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tournamentID := r.PathValue("id")
		itournamentID, err := strconv.Atoi(tournamentID)
		if err != nil {
			s.error(w, http.StatusUnprocessableEntity, err, "Vigane ID")
			return
		}

		if itournamentID < 0 {
			s.error(w, http.StatusUnprocessableEntity, errors.New("vigane id"), "Turniiril ei saa olla negatiivne ID")
			return
		}

		matchID := r.PathValue("team_id")
		iMatchID, err := strconv.Atoi(matchID)
		if err != nil {
			s.error(w, http.StatusUnprocessableEntity, err, "Vigane ID")
			return
		}

		if err := s.store.Club().DeleteTeam(uint(iMatchID), uint(itournamentID)); err != nil {
			s.error(w, http.StatusUnprocessableEntity, err, "Tiimi kustutamisel tekkis tõrge")
			return
		}

		s.respond(w, http.StatusOK, Response{
			Message: "Tiimi kustutamine läks edukalt",
			Data:    nil,
			Error:   nil,
		})

	}
}

func (s *Server) updateTeams() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var requestBody []model.Team
		if err := s.decode(r, &requestBody); err != nil {
			s.error(w, http.StatusUnprocessableEntity, err, "Ei õnnestunud dekodeerida")
			return
		}

		tournamentID := r.PathValue("id")
		itournamentID, err := strconv.Atoi(tournamentID)
		if err != nil {
			s.error(w, http.StatusUnprocessableEntity, err, "Vigane ID")
			return
		}

		if itournamentID < 0 {
			s.error(w, http.StatusUnprocessableEntity, errors.New("vigane id"), "Turniiril ei saa olla negatiivne ID")
			return
		}

		if err := s.store.Club().UpdateTeams(requestBody, uint(itournamentID)); err != nil {
			s.error(w, http.StatusUnprocessableEntity, err, "Tiimide uuendamisel tekkis tõrge")
			return
		}

		s.respond(w, http.StatusOK, Response{
			Message: "Edukalt muudeti turniiride järjestus",
			Data:    nil,
			Error:   nil,
		})
	}
}
