package server

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"os"
	"table-tennis/internal/model"
	"table-tennis/pkg/jwttoken"
	"time"
	"github.com/google/uuid"
)

func (s *Server) setRequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := uuid.New().String()
		w.Header().Set("X-Request-ID", id)
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), ctxKeyRequestID, id)))
	})
}

func (s *Server) logRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rw := &responseWriter{w, http.StatusOK}
		if r.Method == http.MethodOptions {
			next.ServeHTTP(rw, r)
			return
		}

		s.logger.Printf("started %s %s ----- remote_addr:%s request_id:%s",
			r.Method,
			r.RequestURI,
			r.RemoteAddr,
			r.Context().Value(ctxKeyRequestID),
		)
		start := time.Now()
		next.ServeHTTP(rw, r)
		s.logger.Printf("completed in %s with %d %s ----- remote_addr:%s  request_id:%s",
			time.Since(start),
			rw.code,
			http.StatusText(rw.code),
			r.RemoteAddr,
			r.Context().Value(ctxKeyRequestID),
		)
	})
}

func (s *Server) CORSMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers dynamically based on the request's Origin header
		origin := r.Header.Get("Origin")
		if origin != "" {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Credentials", "true")
		}

		// Allow only specific methods for actual requests
		if r.Method == http.MethodOptions {
			w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Access-Control-Allow-Headers, Authorization, X-Requested-With")
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (s *Server) jwtMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		accessToken, err := r.Cookie("at")
		if err != nil {
			refreshToken, err := r.Cookie("rt")
			if err != nil {
				s.error(w, http.StatusUnauthorized, err, "Unauthorized")
				return
			}
			alg := jwttoken.HmacSha256(os.Getenv(jwtKey))
			claims, err := alg.DecodeAndValidate(refreshToken.Value)
			if err != nil {
				s.error(w, http.StatusUnauthorized, err, "Unauthorized")
				return
			}
			id, err := claims.Get("at_id")
			if err != nil {
				s.error(w, http.StatusUnauthorized, err, "Unauthorized")
				return
			}

			session, err := s.store.Session().Check(id.(string))
			if err != nil {
				http.SetCookie(w, DeleteAccessToken())
				http.SetCookie(w, DeleteRefreshToken())
				if err == sql.ErrNoRows {
					s.error(w, http.StatusUnauthorized, errors.New("no valid session"), "Unauthorized")
					return
				}
				s.error(w, http.StatusUnauthorized, err, "Unauthorized")
				return
			}
			accessToken_id := uuid.New().String()
			atToken, err := NewAccessToken(accessToken_id, session.UserID, time.Now().Add(15*time.Minute))
			if err != nil {
				s.error(w, http.StatusUnprocessableEntity, err, "Unauthorized")
				return
			}
			rtToken, err := NewRefreshToken(accessToken_id, time.Now().Add(24*7*time.Hour))
			if err != nil {
				s.error(w, http.StatusUnprocessableEntity, err, "Unauthorized")
				return
			}
			_, err = s.store.Session().Update(session.AcessID, model.Session{AcessID: accessToken_id, UserID: session.UserID, CreatedAT: time.Now().Add(24 * 7 * time.Hour)})
			if err != nil {
				s.error(w, http.StatusUnprocessableEntity, err, "Unauthorized")
				return
			}

			http.SetCookie(w, atToken)
			http.SetCookie(w, rtToken)
			next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), ctxUserID, session.UserID)))
		} else {

			alg := jwttoken.HmacSha256(os.Getenv(jwtKey))
			claims, err := alg.DecodeAndValidate(accessToken.Value)
			if err != nil {
				s.error(w, http.StatusUnauthorized, err, "Unauthorized")
				return
			}

			id, err := claims.Get("user_id")
			if err != nil {
				s.error(w, http.StatusUnauthorized, err, "Unauthorized")
				return
			}

			float, ok := id.(float64)
			if !ok {
				s.error(w, http.StatusUnauthorized, err, "Invalid user id")
				return
			}
			next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), ctxUserID, uint(float))))

		}
	})
}

func (s *Server) corsQuickFix() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin != "" {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Credentials", "true")
		}
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE, PATCH")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Access-Control-Allow-Headers, Authorization, X-Requested-With")
		w.WriteHeader(http.StatusOK)
	}
}
