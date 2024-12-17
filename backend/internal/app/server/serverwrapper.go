package server

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"table-tennis/internal/model"
	psqlstore "table-tennis/internal/store/postgres"
	"time"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func Start() error {

	db, err := newDB()
	if err != nil {
		return err
	}
	err = godotenv.Load(".env")
	if err != nil {
		fmt.Println("error loading", err)
	}
	port := os.Getenv("PORT")

	if port == "" {
		log.Fatal("PORT VALUE NOT SET")
	}

	//migration stuff
	err = db.AutoMigrate(
		&model.User{},
		&model.Club{},
		&model.Tournament{},
		&model.Participant{},
		&model.Match{},
		&model.Result{},
		&model.Set{},
		&model.Team{},
		&model.PlayerJoinTeam{},
		&model.TeamMatch{},
		&model.Session{},
		&model.Blog{},
		&model.LoginUser{},
	)
	if err != nil {
		log.Fatal(err)
	}
	// ONLY USE ONCE, DON't USE IT MULTIPLE TIMES
	// inserDummyData(db)
	// Scarpes the XML of users and puts everything into the database
	// InsertRealDataToDB(db)

	// // // register users
	// participants := []string{
	// 	"Sten Andre KUKISPUU",
	// 	"Martin KRAIKO",
	// 	"Oliver LAMVOL",
	// 	"Sten-Mathias SAARMAN",
	// 	"Mihail SHATROV",
	// 	"Tristan PRIISALM",
	// 	"Sebastian ASU",
	// 	"Reio RÄHN",
	// 	"Simmo NIGUL",
	// 	"Kert TALUMETS",
	// 	"Ivan SYNIYTSIA",
	// 	"Mike KOLESOV",
	// 	"Jevgeni STROGOV",
	// 	"Hans Hugo JAHNSON",
	// 	"Maksim SHIROKIKH",
	// 	"Martin AARN",
	// 	"Kirill ZUBKOV",
	// 	"Daniels KOPANTŠUK",
	// 	"Ivan TŠERNOV",
	// 	"Mathias MAKKO",
	// 	"Frank Tomas TÜRI",
	// }
	// if err := createRegistration(participants, 4, db); err != nil {
	// 	return err
	// }

	// participantsWoman := []string{
	// 	"Reelica HANSON",
	// 	"Reet KULLERKUPP",
	// 	"Margarita GRINENKO",
	// 	"Valerie LONSKI",
	// 	"Annigrete SUIMETS",
	// 	"Kristina ANDREJEVA",
	// 	"Elizaveta VOSTSINA",
	// 	"Amanda HALLIK",
	// 	"Elis TÜRK",
	// 	"Sigrid KUKK",
	// 	"Vitalia REINOL",
	// 	"Arina LITVINOVA",
	// 	"Egle HIIUS",
	// 	"Marianne PEDAK",
	// 	"Julia ŠELIHH",
	// 	"Katrin ADAMSON",
	// 	"Angela LAIDINEN",
	// 	"Tomomi OSANAI",
	// 	"Sirli JAANIMÄGI",
	// 	"Lisanna PÕDER",
	// 	"Diana ANDREJEVA",
	// 	"Airi AVAMERI",
	// }
	// if err := createRegistration(participantsWoman, 1, db); err != nil {
	// 	return err
	// }

	store := psqlstore.New(db)

	srv := newServer(store)

	srv.logger.Printf("The server is running on the port %v", port)

	return http.ListenAndServe(port, srv)
}

// func uploadToS3(file multipart.File, fileName string) error {
// 	// Check for AWS credentials
// 	accessKey := os.Getenv(awsAccessKey)
// 	secretKey := os.Getenv(awsSecretKey)
// 	if accessKey == "" || secretKey == "" {
// 		return errors.New("AWS access key or secret key not set")
// 	}

// 	// Load the AWS config
// 	cfg, err := config.LoadDefaultConfig(context.TODO(),
// 		config.WithRegion(region),
// 		// config.WithCredentialsProvider(aws.NewStaticCredentialsProvider(accessKey, secretKey, "")),
// 		config.WithCredentialsProvider(aws.)
// 	)
// 	if err != nil {
// 		return fmt.Errorf("unable to load SDK config: %w", err)
// 	}

// 	// Create an S3 client
// 	s3Client := s3.NewFromConfig(cfg)

// 	// Prepare the upload parameters
// 	params := &putobject.PutObjectInput{
// 		Bucket: aws.String(bucketName),
// 		Key:    aws.String(fileName),
// 		Body:   file,
// 	}

// 	// Ensure to close the file after the upload
// 	defer file.Close()

// 	// Perform the upload
// 	_, err = s3Client.PutObject(context.TODO(), params)
// 	if err != nil {
// 		return fmt.Errorf("failed to upload to S3: %w", err)
// 	}

// 	return nil
// }

// func createRegistration(participants []string, touranmentId uint, db *gorm.DB) error {
// 	var allUsers []model.User
// 	if err := db.Model(&model.User{}).Find(&allUsers).Error; err != nil {
// 		return err
// 	}

// 	// counter := 0
// 	for _, participant := range participants {
// 		p := strings.SplitN(participant, " ", 2)
// 		if strings.Count(participant, " ") >= 2 {
// 			test := strings.Split(participant, " ")
// 			var output string
// 			for index, t := range test {
// 				if index == len(test)-1 {
// 					p = []string{output, t}
// 				} else {
// 					output += " " + t
// 				}
// 			}
// 		}
// 		for i, user := range allUsers {
// 			if strings.Trim(strings.ToLower(user.FirstName), " ") == strings.Trim(strings.ToLower(p[0]), " ") && strings.Trim(strings.ToLower(user.LastName), " ") == strings.Trim(strings.ToLower(p[1]), " ") {
// 				//create participant
// 				if err := db.Create(&model.Participant{
// 					TournamentID: touranmentId,
// 					PlayerID:     user.ID,
// 				}).Error; err != nil {
// 					fmt.Println("this is error", err)
// 					return err
// 				}
// 				break
// 			} else if i == len(allUsers)-1 {
// 				//else add user and then create participant
// 				wU := model.User{
// 					FirstName: strings.ToUpper(p[0]),
// 					LastName:  strings.ToUpper(p[1]),
// 				}
// 				if err := db.Create(&wU).Error; err != nil {
// 					fmt.Println("This is error", err)
// 					return err
// 				}
// 				//create participant
// 				if err := db.Create(&model.Participant{
// 					TournamentID: touranmentId,
// 					PlayerID:     wU.ID,
// 				}).Error; err != nil {
// 					fmt.Println("Tas is error", err)
// 					return err
// 				}
// 			}
// 		}
// 	}

// 	return nil
// }

func newDB() (*gorm.DB, error) {

	err := godotenv.Load(".env")
	if err != nil {
		fmt.Println("error loading", err)
	}

	dbUser := os.Getenv("POSTGRES_USER")
	dbPassword := os.Getenv("POSTGRES_PASSWORD")
	dbName := os.Getenv("POSTGRES_DB")
	dbHost := os.Getenv("POSTGRES_HOST")
	dbPort := os.Getenv("POSTGRES_PORT")

	if dbUser == "" || dbPassword == "" || dbName == "" || dbHost == "" || dbPort == "" {
		return nil, fmt.Errorf("database connection environment variables not set")
	}

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		dbHost, dbUser, dbPassword, dbName, dbPort)

	gormConfig := &gorm.Config{
		// Logger: logger.Default.LogMode(logger.Info), //logging for database
	}

	db, err := gorm.Open(postgres.Open(dsn), gormConfig)
	if err != nil {
		return nil, err
	}
	// db.SetMaxOpenConns(100)
	// sqlDB.SetMaxIdleConns(10)
	// sqlDB.SetConnMaxLifetime(30 * time.Minute)

	return db, nil
}

func inserDummyData(db *gorm.DB) {
	// 	//-------------TOURNAMENT--------------------//
	tournament_3 := time.Date(2024, time.May, 11, 13, 0, 0, 0, time.UTC)
	tournament_3_end := tournament_3.Add(24 * time.Hour)
	tournament_2 := time.Date(2024, time.August, 24, 0, 0, 0, 0, time.UTC)
	tournament_2_end := tournament_2.Add(24 * time.Hour)
	tournament_4 := time.Date(2024, time.October, 24, 0, 0, 0, 0, time.UTC)
	tournament_4_end := tournament_4.Add(24 * time.Hour * 6)

	single_elim_tournament := &model.Tournament{
		Name:      "Single elim tournmanet",
		StartDate: &tournament_2,
		EndDate:   &tournament_2_end,
		Type:      "single_elimination",
		State:     "created",
	}

	double_elim_weird_tournament := &model.Tournament{
		Name:      "Double elim weird",
		StartDate: &tournament_3,
		EndDate:   &tournament_3_end,
		Type:      "double_elimination_weird",
		State:     "created",
	}

	double_elim_tournament := &model.Tournament{
		Name:      "Double elim normal",
		StartDate: &tournament_4,
		EndDate:   &tournament_4_end,
		Type:      "double_elimination",
		State:     "created",
	}

	if err := db.Create(single_elim_tournament).Error; err != nil {
		return
	}
	if err := db.Create(double_elim_weird_tournament).Error; err != nil {
		return
	}
	if err := db.Create(double_elim_tournament).Error; err != nil {
		return
	}

	//add participants
	db.Create(&model.Participant{PlayerID: 695, TournamentID: single_elim_tournament.ID})
	db.Create(&model.Participant{PlayerID: 698, TournamentID: single_elim_tournament.ID})
	db.Create(&model.Participant{PlayerID: 699, TournamentID: single_elim_tournament.ID})
	db.Create(&model.Participant{PlayerID: 700, TournamentID: single_elim_tournament.ID})
	db.Create(&model.Participant{PlayerID: 701, TournamentID: single_elim_tournament.ID})
	db.Create(&model.Participant{PlayerID: 702, TournamentID: single_elim_tournament.ID})
	db.Create(&model.Participant{PlayerID: 703, TournamentID: single_elim_tournament.ID})
	db.Create(&model.Participant{PlayerID: 704, TournamentID: single_elim_tournament.ID})
	db.Create(&model.Participant{PlayerID: 705, TournamentID: single_elim_tournament.ID})
	db.Create(&model.Participant{PlayerID: 706, TournamentID: single_elim_tournament.ID})
	db.Create(&model.Participant{PlayerID: 707, TournamentID: single_elim_tournament.ID})

	db.Create(&model.Participant{PlayerID: 708, TournamentID: double_elim_tournament.ID})
	db.Create(&model.Participant{PlayerID: 709, TournamentID: double_elim_tournament.ID})
	db.Create(&model.Participant{PlayerID: 710, TournamentID: double_elim_tournament.ID})
	db.Create(&model.Participant{PlayerID: 711, TournamentID: double_elim_tournament.ID})
	db.Create(&model.Participant{PlayerID: 712, TournamentID: double_elim_tournament.ID})
	db.Create(&model.Participant{PlayerID: 713, TournamentID: double_elim_tournament.ID})
	db.Create(&model.Participant{PlayerID: 714, TournamentID: double_elim_tournament.ID})
	db.Create(&model.Participant{PlayerID: 715, TournamentID: double_elim_tournament.ID})
	db.Create(&model.Participant{PlayerID: 716, TournamentID: double_elim_tournament.ID})
	db.Create(&model.Participant{PlayerID: 717, TournamentID: double_elim_tournament.ID})
	db.Create(&model.Participant{PlayerID: 718, TournamentID: double_elim_tournament.ID})

	db.Create(&model.Participant{PlayerID: 719, TournamentID: double_elim_weird_tournament.ID})
	db.Create(&model.Participant{PlayerID: 720, TournamentID: double_elim_weird_tournament.ID})
	db.Create(&model.Participant{PlayerID: 721, TournamentID: double_elim_weird_tournament.ID})
	db.Create(&model.Participant{PlayerID: 722, TournamentID: double_elim_weird_tournament.ID})
	db.Create(&model.Participant{PlayerID: 723, TournamentID: double_elim_weird_tournament.ID})
	db.Create(&model.Participant{PlayerID: 724, TournamentID: double_elim_weird_tournament.ID})
	db.Create(&model.Participant{PlayerID: 725, TournamentID: double_elim_weird_tournament.ID})
	db.Create(&model.Participant{PlayerID: 726, TournamentID: double_elim_weird_tournament.ID})
	db.Create(&model.Participant{PlayerID: 727, TournamentID: double_elim_weird_tournament.ID})
	db.Create(&model.Participant{PlayerID: 728, TournamentID: double_elim_weird_tournament.ID})
	db.Create(&model.Participant{PlayerID: 729, TournamentID: double_elim_weird_tournament.ID})

	db.Create(&model.Participant{PlayerID: 740, TournamentID: double_elim_weird_tournament.ID})
	db.Create(&model.Participant{PlayerID: 741, TournamentID: double_elim_weird_tournament.ID})
	db.Create(&model.Participant{PlayerID: 742, TournamentID: double_elim_weird_tournament.ID})
	db.Create(&model.Participant{PlayerID: 743, TournamentID: double_elim_weird_tournament.ID})
	db.Create(&model.Participant{PlayerID: 744, TournamentID: double_elim_weird_tournament.ID})
	db.Create(&model.Participant{PlayerID: 745, TournamentID: double_elim_weird_tournament.ID})
	db.Create(&model.Participant{PlayerID: 746, TournamentID: double_elim_weird_tournament.ID})
	db.Create(&model.Participant{PlayerID: 747, TournamentID: double_elim_weird_tournament.ID})
	db.Create(&model.Participant{PlayerID: 748, TournamentID: double_elim_weird_tournament.ID})
	db.Create(&model.Participant{PlayerID: 749, TournamentID: double_elim_weird_tournament.ID})
	db.Create(&model.Participant{PlayerID: 750, TournamentID: double_elim_weird_tournament.ID})
	db.Create(&model.Participant{PlayerID: 740, TournamentID: double_elim_weird_tournament.ID})
	db.Create(&model.Participant{PlayerID: 741, TournamentID: double_elim_weird_tournament.ID})
	db.Create(&model.Participant{PlayerID: 742, TournamentID: double_elim_weird_tournament.ID})
	db.Create(&model.Participant{PlayerID: 743, TournamentID: double_elim_weird_tournament.ID})
	db.Create(&model.Participant{PlayerID: 744, TournamentID: double_elim_weird_tournament.ID})
	db.Create(&model.Participant{PlayerID: 745, TournamentID: double_elim_weird_tournament.ID})
	db.Create(&model.Participant{PlayerID: 746, TournamentID: double_elim_weird_tournament.ID})
	db.Create(&model.Participant{PlayerID: 747, TournamentID: double_elim_weird_tournament.ID})
	db.Create(&model.Participant{PlayerID: 748, TournamentID: double_elim_weird_tournament.ID})
	db.Create(&model.Participant{PlayerID: 749, TournamentID: double_elim_weird_tournament.ID})
	db.Create(&model.Participant{PlayerID: 750, TournamentID: double_elim_weird_tournament.ID})

}
