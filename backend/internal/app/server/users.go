package server

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"table-tennis/internal/model"
	"table-tennis/pkg/validator"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

func (s *Server) createUser() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var requestBody struct {
			FirstName string `json:"first_name"`
			LastName  string `json:"last_name"`
			Password  string `json:"password"`
			Email     string `json:"email"`
		}
		if err := s.decode(r, &requestBody); err != nil {
			s.error(w, http.StatusUnprocessableEntity, err, "Ei õnnestunud dekodeerida")
			return
		}
		err := s.store.LoginUser().Create(model.LoginUser{
			FirstName: requestBody.FirstName,
			LastName:  requestBody.LastName,
			Password:  requestBody.Password,
			Email:     requestBody.Email,
			Role:      0,
		})
		if err != nil {
			s.error(w, http.StatusUnprocessableEntity, err, "Kasutaja loomine ebaõnnestus")
		}

		s.respond(w, http.StatusCreated, Response{
			Message: "Mängija edukalt lisatud",
			Data:    nil,
		})
	}
}

func (s *Server) getUser() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := r.PathValue("id")
		iuserID, err := strconv.Atoi(userID)
		if err != nil {
			s.error(w, http.StatusUnprocessableEntity, err, "Ebakorrektne ID")
			return
		}
		if iuserID < 0 {
			s.error(w, http.StatusUnprocessableEntity, errors.New("ebakorrektne ID"), "Negatiivse IDga kasutajad puuduvad")
			return
		}
		user, err := s.store.User().Get(uint(iuserID))
		if err != nil {
			s.error(w, http.StatusUnprocessableEntity, err, "Kasutajat ei leitud")
			return
		}
		s.respond(w, http.StatusOK, Response{
			Message: "Kasutaja leitud",
			Data:    user,
		})
	}
}

func (s *Server) getAllUsers() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, err := s.store.User().GetAll()
		if err != nil {
			s.error(w, http.StatusUnprocessableEntity, err, "Kasutajaid ei letitud")
			return
		}
		s.respond(w, 200, Response{
			Message: "Kasutajad edukalt leitud",
			Data:    user,
		})
	}
}

func (s *Server) getUserMatches() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user_id := r.PathValue("id")
		user_id_int, err := strconv.Atoi(user_id)
		if err != nil {
			s.error(w, http.StatusUnprocessableEntity, err, "ID peab olema number")
			return
		}
		matches, err := s.store.User().GetMatches(uint(user_id_int))
		if err != nil {
			s.error(w, http.StatusUnprocessableEntity, err, "Kasutaja mänge ei leitud")
			return
		}
		s.respond(w, http.StatusOK, Response{
			Message: "Mängud edukalt leitud",
			Data:    matches,
			Error:   nil,
		})
	}
}

func (s *Server) userLogin() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var loginDetails struct {
			Login    string `json:"login" validate:"required"`
			Password string `json:"password" validate:"required"`
		}

		if err := s.decode(r, &loginDetails); err != nil {
			s.error(w, http.StatusUnauthorized, err, "Ei õnnestunud dekodeerida")
			return
		}

		if err := validator.Validate(loginDetails); err != nil {
			s.error(w, http.StatusUnprocessableEntity, err, "Volitamata")
			return
		}
		user, err := s.store.LoginUser().GetByLogin(loginDetails.Login)
		if err != nil {
			if err == sql.ErrNoRows {
				s.error(w, http.StatusUnprocessableEntity, errors.New(""), "Volitamata")
				return
			}
			s.error(w, http.StatusUnprocessableEntity, err, "Volitamata")
			return
		}
		fmt.Println("loginDetails", loginDetails.Password)
		fmt.Printf("%+v\n", user)
		if !user.ComparePassword(loginDetails.Password) {
			s.error(w, http.StatusUnauthorized, errors.New("vale parool või kasutajanimi"), "volitamata")
			return
		}

		accesstoken_id := uuid.New().String()
		accessToken, err := NewAccessToken(accesstoken_id, user.ID, time.Now().Add(15*time.Minute))
		if err != nil {
			s.error(w, http.StatusUnprocessableEntity, err, "volitamata")
			return
		}

		refreshToken, err := NewRefreshToken(accesstoken_id, time.Now().Add(24*7*time.Hour))
		if err != nil {
			s.error(w, http.StatusUnprocessableEntity, err, "volitamata")
			return
		}

		session := model.Session{
			AcessID:   accesstoken_id,
			UserID:    user.ID,
			CreatedAT: time.Now().Add(24 * 7 * time.Hour),
		}

		oldSession, err := s.store.Session().CheckByUserId(user.ID)
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				_, err = s.store.Session().Create(&session)
				if err != nil {
					s.error(w, http.StatusUnprocessableEntity, err, "volitamata")
					return
				}
				http.SetCookie(w, accessToken)
				http.SetCookie(w, refreshToken)
				user.Sanitize()
				s.respond(w, http.StatusOK, Response{Data: user})
				return
			}
			s.error(w, http.StatusUnprocessableEntity, err, "volitamata")
			return
		}
		_, err = s.store.Session().Update(oldSession.AcessID, session)
		if err != nil {
			s.error(w, http.StatusUnprocessableEntity, err, "volitamata")
			return
		}

		http.SetCookie(w, accessToken)
		http.SetCookie(w, refreshToken)
		user.Sanitize()
		s.respond(w, http.StatusOK, Response{Data: user})
	}
}

func (s *Server) checkSession() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := r.Context().Value(ctxUserID).(uint)
		if !ok {
			s.error(w, http.StatusUnauthorized, errors.New("volitamata"), "Ligipääs on keelatud")
			return
		}
		user, err := s.store.LoginUser().Get(uint(userID))
		if err != nil {
			s.error(w, http.StatusUnprocessableEntity, err, "Volitamata")
			return
		}
		user.Sanitize()
		s.respond(w, http.StatusOK, Response{
			Message: "Kasutaja edukalt leitud",
			Data:    user,
			Error:   nil,
		})
	}
}

func (s *Server) userLogout() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := r.Context().Value(ctxUserID).(uint)
		if !ok {
			s.error(w, http.StatusUnauthorized, errors.New("volitamata"), "Volitamata")
			return
		}
		err := s.store.Session().DeleteByUser(userID)
		if err != nil {
			s.error(w, http.StatusUnprocessableEntity, err, "Välja logimine ebaõnnestus")
			return
		}
		http.SetCookie(w, DeleteAccessToken())
		http.SetCookie(w, DeleteRefreshToken())
		s.respond(w, http.StatusOK, Response{
			Message: "Edukalt väljalogitud",
			Data:    nil,
			Error:   nil,
		})
	}
}
