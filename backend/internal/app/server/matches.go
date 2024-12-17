package server

import (
	"errors"
	"net/http"
	"strconv"
	"table-tennis/internal/model"
	"table-tennis/pkg/validator"
)

func (s *Server) finishMatch() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var match model.Match
		if err := s.decode(r, &match); err != nil {
			s.error(w, http.StatusUnprocessableEntity, err, "Ei õnnestunud dekodeerida")
			return
		}

		if err := s.store.Tournament().FinishMatch(&match, nil); err != nil {
			s.error(w, http.StatusUnprocessableEntity, err, "Mängu ei saa lõpetada. Sisestatud info on puudulik või mäng on juba lõpetatud")
			return
		}

		s.respond(w, http.StatusOK, Response{
			Message: "Mäng on edukalt lõpetatud!",
			Error:   nil,
			Data:    nil,
		})
	}
}

func (s *Server) getRegroupedTeams() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tournament_id := r.PathValue("id")
		iTournament_id, err := strconv.Atoi(tournament_id)
		if err != nil {
			s.error(w, http.StatusUnprocessableEntity, err, "Vigane ID")
			return
		}

		teams, err := s.store.Tournament().GetRegroupedMatches(uint(iTournament_id))
		if err != nil {
			s.error(w, http.StatusUnprocessableEntity, err, "Hetkel ei saa kuvada tiimide järjestust")
			return
		}

		s.respond(w, http.StatusOK, Response{
			Message: "Tiimide järjestus on edukalt loodud",
			Error:   nil,
			Data:    teams,
		})
	}
}

func (s *Server) deleteRegroupedMatches() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tournament_id := r.PathValue("id")
		iTournamentId, err := strconv.Atoi(tournament_id)
		if err != nil {
			s.error(w, http.StatusUnprocessableEntity, err, "Vigane ID")
			return
		}

		if err := s.store.Tournament().DeleteRegroupedMatches(uint(iTournamentId)); err != nil {
			s.error(w, http.StatusUnprocessableEntity, err, "Regrupeeritud mängude kustutamine ebaõnnestus")
			return
		}

		s.respond(w, http.StatusOK, Response{
			Message: "Regrupeering on edukalt kustutatud",
			Error:   nil,
			Data:    "Regrupeering on edukalt kustutatud",
		})
	}
}

func (s *Server) createRegroupedMatches() http.HandlerFunc {
	var requestBody struct {
		Matches      []model.Team `json:"teams" validate:"required"`
		TournamentID int          `json:"tournament_id" validate:"required"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		if err := s.decode(r, &requestBody); err != nil {
			s.error(w, http.StatusUnprocessableEntity, err, "Ei õnnestunud dekodeerida")
			return
		}

		if err := validator.Validate(&requestBody); err != nil {
			s.error(w, http.StatusUnprocessableEntity, err, "Valideerimine ebaõnnestus")
			return
		}
		//check if actually all the teams have according tournament_id
		for _, team := range requestBody.Matches {
			if team.TournamentID != uint(requestBody.TournamentID) {
				s.error(w, http.StatusUnprocessableEntity, errors.New("tiimi info on ebakorrektne"), "Tiimide info on ebakorrektne")
				return
			}
		}

		if err := s.store.Tournament().CreateRegroupedMatches(uint(requestBody.TournamentID), requestBody.Matches); err != nil {
			s.error(w, http.StatusUnprocessableEntity, err, "Midagi läks viltu. Värskendage veebilehte ja proovige uuesti")
			return
		}

		s.respond(w, http.StatusOK, Response{
			Message: "Regrupeering on edukalt läbi viidud",
			Error:   nil,
			Data:    nil,
		})
	}
}

func (s *Server) createMatch() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var match model.Match
		if err := s.decode(r, &match); err != nil {
			s.error(w, http.StatusUnprocessableEntity, err, "Ei õnnestunud dekodeerida")
			return
		}

		if err := s.store.Match().CreateMatch(match); err != nil {
			s.error(w, http.StatusUnprocessableEntity, err, "Mängu ei õnnestunud luua. Kontrollige andmeid ja proovige uuesti")
			return
		}

		s.respond(w, http.StatusOK, Response{
			Message: "Mäng loodi edukalt",
			Error:   nil,
			Data:    nil,
		})
	}
}

func (s *Server) updateTable() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var requestBody struct {
			Table int `json:"table"`
		}
		if err := s.decode(r, &requestBody); err != nil {
			s.error(w, http.StatusUnprocessableEntity, err, "Ei õnnestunud dekodeerida")
			return
		}

		matchID := r.PathValue("id")
		iMatchID, err := strconv.Atoi(matchID)
		if err != nil {
			s.error(w, http.StatusUnprocessableEntity, err, "Vigane ID")
			return
		}
		if err := s.store.Match().UpdateTable(uint(iMatchID), requestBody.Table); err != nil {
			s.error(w, http.StatusUnprocessableEntity, err, "Mängu laua numbri vahetamisel tekkis tõrge. Värskendage veebilehte ja proovige uuesti")
			return
		}
		s.respond(w, http.StatusOK, Response{
			Message: "Mängu laua number uuendati",
			Error:   nil,
			Data:    nil,
		})
	}
}

func (s *Server) getTeamMatches() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		team_match_id := r.PathValue("team_match_id")
		iTeamMatchID, err := strconv.Atoi(team_match_id)
		if err != nil {
			s.error(w, http.StatusUnprocessableEntity, err, "Vigane ID")
			return
		}
		_ = iTeamMatchID
		tournament_id := r.PathValue("id")
		iTournamentID, err := strconv.Atoi(tournament_id)
		if err != nil {
			s.error(w, http.StatusUnprocessableEntity, err, "Vigane ID")
			return
		}

		matches, err := s.store.Club().GetTeamMatches(uint(iTournamentID), uint(iTeamMatchID))
		if err != nil {
			s.error(w, http.StatusUnprocessableEntity, err, "Ei leidnud tiimimängu")
			return
		}

		//Call function to get team_matches
		s.respond(w, http.StatusOK, Response{
			Message: "Edukalt männgud leitud",
			Data:    matches,
			Error:   nil,
		})

	}
}

func (s *Server) getTeamMatchScores() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		type matchScore struct {
			Team1Score uint `json:"team_1_score"`
			Team2Score uint `json:"team_2_score"`
		}

		var data struct {
			TotalScore []matchScore `json:"match_score"`
		}

		teamMatchID := r.PathValue("team_match_id")
		iTeamMatchID, err := strconv.Atoi(teamMatchID)
		if err != nil {
			s.error(w, http.StatusUnprocessableEntity, err, "Vigane ID")
			return
		}

		tournamentID := r.PathValue("id")
		iTournamentID, err := strconv.Atoi(tournamentID)
		if err != nil {
			s.error(w, http.StatusUnprocessableEntity, err, "Vigane ID")
			return
		}

		matches, err := s.store.Club().GetTeamMatches(uint(iTournamentID), uint(iTeamMatchID))
		if err != nil {
			s.error(w, http.StatusUnprocessableEntity, err, "Mänge ei leitud")
			return
		}

		var team1TotalScore uint
		var team2TotalScore uint

		for _, match := range matches {
			sets, err := s.store.Match().GetMatchSets(match.ID)
			if err != nil {
				s.error(w, http.StatusUnprocessableEntity, err, "Mängule ei leitud sette")
				return
			}

			var team1SetWins uint
			var team2SetWins uint

			for _, set := range sets {
				team1Points := set.Team1Score
				team2Points := set.Team2Score

				if team1Points >= 11 && (team1Points-team2Points) >= 2 {
					team1SetWins++
				} else if team2Points >= 11 && (team2Points-team1Points) >= 2 {
					team2SetWins++
				}
			}

			if team1SetWins >= 3 {
				team1TotalScore++
			} else if team2SetWins >= 3 {
				team2TotalScore++
			}

			data.TotalScore = append(data.TotalScore, matchScore{
				Team1Score: team1TotalScore,
				Team2Score: team2TotalScore,
			})
		}

		s.respond(w, http.StatusOK, Response{
			Message: "Mängu skoorid leiti edukalt",
			Data:    data,
			Error:   nil,
		})
	}
}

func (s *Server) updateForfeitMatch() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		type requestBody struct {
			MatchID        int  `json:"match_id"`
			WinnerID       int  `json:"winner_id"`
			IsForfeitMatch bool `json:"is_forfeit_match"`
		}
		var rb requestBody
		if err := s.decode(r, &rb); err != nil {
			s.error(w, http.StatusUnprocessableEntity, err, "Vigane Info")
			return
		}
		if err := s.store.Match().UpdateForfeithMatch(uint(rb.MatchID), uint(rb.WinnerID), rb.IsForfeitMatch); err != nil {
			s.error(w, http.StatusUnprocessableEntity, err, "Loobumisvõidu loomisel tekkis viga")
			return
		}

		s.respond(w, http.StatusOK, Response{
			Message: "Loobumisvõit sisestati edukalt",
			Data:    nil,
			Error:   nil,
		})
	}
}
