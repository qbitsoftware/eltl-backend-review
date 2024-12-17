package server

import (
	"errors"
	"net/http"
	"strconv"
	"table-tennis/internal/model"
	"table-tennis/pkg/validator"
	"time"

	"gorm.io/gorm"
)

func (s *Server) getTournament() http.HandlerFunc {
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
		_, err = r.Cookie("at")
		isLoggedIn := err == nil

		t, err := s.store.Tournament().GetWithPrivate(uint(itournamentID), isLoggedIn)
		if err != nil {
			s.error(w, http.StatusUnprocessableEntity, err, "Turniiri ei leitud")
			return
		}
		s.respond(w, http.StatusOK, Response{
			Message: "Turniir leiti edukalt",
			Data:    t,
			Error:   nil,
		})
	}
}

func (s *Server) getTimeTable() http.HandlerFunc {
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
		matches, err := s.store.Tournament().GetTimeTable(uint(itournamentID))
		if err != nil {
			s.error(w, http.StatusUnprocessableEntity, err, "Ajakava leidmisel tekkis viga")
			return
		}
		s.respond(w, http.StatusOK, Response{
			Message: "Ajakava leiti edukalt",
			Data:    matches,
			Error:   nil,
		})
	}
}

func (s *Server) getAllTournaments() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_, err := r.Cookie("at")
		isLoggedIn := err == nil
		t, err := s.store.Tournament().GetAll(isLoggedIn)
		if err != nil {
			s.error(w, http.StatusUnprocessableEntity, err, "Hetkel ei ole ühtegi turniiri")
			return
		}
		s.respond(w, http.StatusOK, Response{
			Message: "Turniirid leiti edukalt",
			Data:    t,
			Error:   nil,
		})
	}
}

func (s *Server) getProtocols() http.HandlerFunc {
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

		matches, err := s.store.Tournament().GetProtocols(uint(itournamentID))
		if err != nil {
			s.error(w, http.StatusUnprocessableEntity, err, "Turniiri protokollide leidmisel oli viga")
			return
		}

		s.respond(w, http.StatusOK, Response{
			Message: "Edukalt leiti turniiri protokollid",
			Data:    matches,
			Error:   nil,
		})
	}
}

func (s *Server) updateTournament() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var input struct {
			ID        uint   `json:"ID"`
			Name      string `json:"name"`
			StartDate string `json:"start_date"`
			EndDate   string `json:"end_date"`
			Type      string `json:"type"`
			State     string `json:"state"`
			Private   bool   `json:"private"`
		}
		if err := s.decode(r, &input); err != nil {
			s.error(w, http.StatusUnprocessableEntity, err, "Ei õnnestunud dekodeerida")
			return
		}

		sD, err := time.Parse("2006-01-02 15:04", input.StartDate)
		if err != nil {
			s.error(w, http.StatusUnprocessableEntity, err, "Turniiri ajaformaat on vale")
			return
		}

		eD, err := time.Parse("2006-01-02 15:04", input.EndDate)
		if err != nil {
			s.error(w, http.StatusUnprocessableEntity, err, "Turniiri ajaformaat on vale")
			return
		}

		t := model.Tournament{
			Model: gorm.Model{
				ID: input.ID,
			},
			Name:      input.Name,
			StartDate: &sD,
			EndDate:   &eD,
			Type:      input.Type,
			State:     input.State,
			Private:   input.Private,
		}

		if err := s.store.Tournament().Update(t); err != nil {
			s.error(w, http.StatusUnprocessableEntity, err, "Turniiri uuendamisel tekkis viga")
			return
		}
		s.respond(w, http.StatusOK, Response{
			Message: "Turniiri uuendati edukalt",
			Data:    t,
			Error:   nil,
		})
	}
}

func (s *Server) createTournament() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var input struct {
			Name      string `json:"name" validate:"required"`
			StartDate string `json:"start_date" validate:"required"`
			EndDate   string `json:"end_date" validate:"required"`
			Type      string `json:"type" validate:"required|contains:meistriliiga"`
			State     string `json:"state"`
			Private   bool   `json:"private"`
		}
		if err := s.decode(r, &input); err != nil {
			s.error(w, http.StatusUnprocessableEntity, err, "Ei õnnestunud dekodeerida")
			return
		}
		if err := validator.Validate(&input); err != nil {
			s.error(w, http.StatusUnprocessableEntity, err, "Sisend on vale. Kontrollige sisendit ja proovige uuesti")
			return
		}

		sD, err := time.Parse("2006-01-02 15:04", input.StartDate)
		if err != nil {
			s.error(w, http.StatusUnprocessableEntity, err, "Turniiri ajaformaat on vale")
			return
		}

		eD, err := time.Parse("2006-01-02 15:04", input.EndDate)
		if err != nil {
			s.error(w, http.StatusUnprocessableEntity, err, "Turniiri ajaformaat on vale")
			return
		}

		t := model.Tournament{
			Name:      input.Name,
			StartDate: &sD,
			EndDate:   &eD,
			Type:      input.Type,
			State:     input.State,
			Private:   input.Private,
		}

		if err := s.store.Tournament().Create(t); err != nil {
			s.error(w, http.StatusUnprocessableEntity, err, "Turniiri loomisel tekkis viga")
			return
		}
		s.respond(w, http.StatusOK, Response{
			Message: "Turniir loodi edukalt",
			Data:    nil,
			Error:   nil,
		})
	}
}

func (s *Server) getAllParticipants() http.HandlerFunc {
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
		u, err := s.store.Tournament().GetParticipants(uint(itournamentID))
		if err != nil {
			s.error(w, http.StatusUnprocessableEntity, err, "Turniiri mängijate leidmisel tekkis viga")
			return
		}
		s.respond(w, http.StatusOK, Response{
			Message: "Registreerinud mängijad leiti edukalt",
			Data:    u,
			Error:   nil,
		})
	}
}

func (s *Server) generate() http.HandlerFunc {
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
		err = s.store.Tournament().Generate(uint(itournamentID))
		if err != nil {
			s.error(w, http.StatusUnprocessableEntity, err, "Turniiri mängude genereerimisel läks midagi valesti")
			return
		}
		s.respond(w, http.StatusOK, Response{
			Message: "Turniiri mängud loodi edukalt",
			Data:    nil,
			Error:   nil,
		})
	}
}

func (s *Server) showBracket() http.HandlerFunc {
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
		t, err := s.store.Tournament().ShowTable(uint(itournamentID))
		if err != nil {
			s.error(w, http.StatusUnprocessableEntity, err, "Turniiri edetabeli koostamisel, läks midagi valesti")
			return
		}
		s.respond(w, http.StatusOK, Response{
			Message: "Edukalt koostati turniiri edetabel",
			Data:    t,
			Error:   nil,
		})
	}
}

func (s *Server) calcTournament() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		iTournamentID, err := strconv.Atoi(id)
		if err != nil {
			s.error(w, http.StatusUnprocessableEntity, err, "Vigane ID")
			return
		}
		if err := s.store.Tournament().CalculateRating(uint(iTournamentID)); err != nil {
			s.error(w, http.StatusUnprocessableEntity, err, "Turniiri reitingu arvutamisel tekkis viga")
			return
		}
		s.respond(w, http.StatusOK, Response{
			Message: "Reitingud on edukalt uuendatud",
			Data:    nil,
			Error:   nil,
		})
	}
}

func (s *Server) deleteTournament() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		iTournamentID, err := strconv.Atoi(id)
		if err != nil {
			s.error(w, http.StatusUnprocessableEntity, err, "Vigane ID")
			return
		}
		if err := s.store.Tournament().Delete(uint(iTournamentID)); err != nil {
			s.error(w, http.StatusUnprocessableEntity, err, "Turniiri kustutamisel tekkis viga")
			return
		}
		s.respond(w, http.StatusOK, Response{
			Message: "Turniir on edukalt kustutatud",
			Data:    nil,
			Error:   nil,
		})
	}
}
func (s *Server) showMeistriliiga() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		iTournamentID, err := strconv.Atoi(id)
		if err != nil {
			s.error(w, http.StatusUnprocessableEntity, err, "Vigane ID")
			return
		}

		table, err := s.store.Tournament().GetMeistriliigaTabel(uint(iTournamentID))
		if err != nil {
			s.error(w, http.StatusUnprocessableEntity, err, "Turniiri edetabeli koostamine ebaõnnestus")
			return
		}
		s.respond(w, http.StatusOK, Response{
			Message: "Edukalt koostati turniiri edetabel",
			Data:    table,
			Error:   nil,
		})
	}
}
