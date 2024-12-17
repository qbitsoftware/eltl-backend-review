package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"table-tennis/internal/store"
	"table-tennis/pkg/router"
)

const (
	jwtKey                      = "JWT_KEY"
	awsRegion                   = "AWS_REGION"
	awsUploadBucket             = "AWS_UPLOAD_BUCKET"
	awsRetrieveBucket           = "AWS_RETRIEVE_BUCKET"
	awsAccessKey                = "AWS_ACCESS_KEY"
	awsSecretKey                = "AWS_SECRET_KEY"
	awsRetrieveBucketUrl        = "AWS_RETRIEVE_BUCKET_URL"
	awsUploadBucketUrl          = "AWS_UPLOAD_BUCKET_URL"
	ctxKeyRequestID      ctxKey = iota
	ctxUserID
)

type ctxKey int

type Response struct {
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
	Error   interface{} `json:"error"`
}

type Server struct {
	router *router.Router
	logger *log.Logger
	store  store.Store
	awsS3  *Bucket
}

func newServer(store store.Store) *Server {
	s := &Server{
		router: router.New(),
		logger: log.New(os.Stdout, "", 0),
		store:  store,
		awsS3:  newS3Client(),
	}

	configureRouter(s)

	return s
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}

func configureRouter(s *Server) {
	s.router.OPTION("/", s.corsQuickFix())

	s.router.Use(s.setRequestID, s.logRequest, s.CORSMiddleware)
	s.router.UseWithPrefix("auth", s.jwtMiddleware)
	//---------USER---------//
	s.router.POST("/api/v1/users", s.createUser())
	s.router.GET("/api/v1/users/{id}", s.getUser())
	s.router.GET("/api/v1/users", s.getAllUsers())
	s.router.GET("/api/v1/users/{id}/matches", s.getUserMatches())
	s.router.GET("/api/v1/auth/users/current", s.checkSession())
	//---------TOURNAMENTS-----------------//
	s.router.GET("/api/v1/tournaments", s.getAllTournaments())
	s.router.POST("/api/v1/auth/tournaments", s.createTournament())
	s.router.PATCH("/api/v1/auth/tournaments", s.updateTournament())
	s.router.DELETE("/api/v1/auth/tournaments/{id}", s.deleteTournament())
	s.router.GET("/api/v1/tournaments/{id}", s.getTournament())
	s.router.GET("/api/v1/tournaments/{id}/participants", s.getAllParticipants())
	s.router.POST("/api/v1/auth/tournaments/{id}/generate", s.generate())
	s.router.GET("/api/v1/tournaments/{id}/bracket", s.showBracket())
	s.router.GET("/api/v1/tournaments/{id}/groupbracket", s.showMeistriliiga())
	s.router.GET("/api/v1/tournaments/{id}/timetable", s.getTimeTable())
	s.router.GET("/api/v1/tournaments/{id}/protocols", s.getProtocols())
	//--------MATCHES-------------//
	s.router.POST("/api/v1/auth/tournaments/matches/set", s.createSet())
	s.router.PATCH("/api/v1/auth/tournaments/matches/set", s.updateSet())
	s.router.POST("/api/v1/auth/tournaments/matches/finish", s.finishMatch())
	s.router.POST("/api/v1/auth/ournaments/matches/set/multiple", s.createSetsMultiple())
	s.router.GET("/api/v1/tournaments/matches/{id}/set", s.getSets())
	s.router.GET("/api/v1/tournaments/calculate/test/{id}", s.calcTournament())
	s.router.GET("/api/v1/auth/tournaments/{id}/matches/regrouped", s.getRegroupedTeams())
	s.router.POST("/api/v1/auth/tournaments/create/regrouped", s.createRegroupedMatches())
	s.router.DELETE("/api/v1/auth/tournaments/{id}/regrouped", s.deleteRegroupedMatches())
	s.router.PATCH("/api/v1/auth/tournaments/matches/forfeit", s.updateForfeitMatch())
	//---------TEAMS------------//
	s.router.POST("/api/v1/auth/tournaments/{id}/team", s.createTeam())
	s.router.DELETE("/api/v1/auth/tournaments/{id}/teams/{team_id}", s.deleteTeam())
	s.router.POST("/api/v1/auth/tournaments/{id}/teams", s.updateTeams())
	s.router.PATCH("/api/v1/auth/tournaments/{id}/teams/{team_id}", s.updateTeam())
	s.router.GET("/api/v1/tournaments/{id}/teams", s.getTeams())
	s.router.GET("/api/v1/teams/{team_id}", s.getTeam())
	//---------TEAMS-MATCHES----------//
	s.router.GET("/api/v1/team_matches/{match_id}", s.getTeamMatch())
	s.router.POST("/api/v1/auth/team_matches", s.createTeamMatch())
	s.router.DELETE("/api/v1/auth/team_matches/{team_match_id}", s.deleteTeamMatch())
	s.router.PATCH("/api/v1/auth/team_matches/{id}", s.updateTeamMatch())
	s.router.POST("/api/v1/auth/tournaments/matches/create", s.createMatch())
	s.router.POST("/api/v1/auth/team_matches/clock", s.updateTime())
	s.router.POST("/api/v1/auth/matches/{id}/table", s.updateTable())
	s.router.GET("/api/v1/tournaments/{id}/matches/get/{team_match_id}/score", s.getTeamMatchScores())
	s.router.GET("/api/v1/tournaments/{id}/matches/get/{team_match_id}", s.getTeamMatches())
	// s.router.GET("/api/v1/put/users/to/db", s.scrapeUsers())
	//------------BLOGS-----------------//
	s.router.POST("/api/v1/auth/tournaments/{id}/blog", s.createBlog())
	s.router.GET("/api/v1/tournaments/{id}/blog", s.getBlogs())
	s.router.POST("/api/v1/login", s.userLogin())
	s.router.POST("/api/v1/auth/logout", s.userLogout())
	//-----------IMAGES-------------------//
	s.router.POST("/api/v1/auth/{tournament_id}/image/{day}/upload", s.imageUpload())
	s.router.POST("/api/v1/auth/image", s.imageRemove())
	s.router.GET("/api/v1/tournaments/{id}/images/{day}/get", s.getBucketImages())
}

// func (s *Server) scrapeUsers() http.HandlerFunc {
// 	return func(w http.ResponseWriter, r *http.Request) {

// 		if err := s.store.User().ScrapeUsers(); err != nil {
// s.error(w, http.StatusUnprocessableEntity, err, "Usereid ei saanud lauatennise liidu lehekyljeslt katte")
// 			return
// 		}
// 		s.respond(w, http.StatusOK, Response{
// 			Message: "Edukalt leiti kõik tiimi mängijad",
// 			Data:    "Successful",
// 			Error:   nil,
// 		})
// 	}
// }

func (s *Server) error(w http.ResponseWriter, code int, err error, msg string) {
	s.respond(w, code, Response{
		Message: msg,
		Data:    nil,
		Error:   err.Error(),
	})
}

func (s *Server) respond(w http.ResponseWriter, code int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	if data != nil {
		json.NewEncoder(w).Encode(data)
	}
}

func (s *Server) decode(r *http.Request, data interface{}) error {
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		return fmt.Errorf("decode json: %w", err)
	}
	return nil
}
