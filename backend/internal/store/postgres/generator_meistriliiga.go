package psqlstore

import (
	"errors"
	"fmt"
	"table-tennis/internal/model"
	"time"

	"gorm.io/gorm"
)

var MeistriLiigaMatches = [7][4][2]int{
	{{1, 2}, {3, 8}, {4, 7}, {5, 6}},
	{{1, 6}, {5, 7}, {4, 8}, {2, 3}},
	{{1, 4}, {3, 5}, {2, 6}, {7, 8}},
	{{1, 5}, {4, 6}, {3, 7}, {2, 8}},
	{{1, 3}, {2, 4}, {5, 8}, {6, 7}},
	{{1, 7}, {6, 8}, {2, 5}, {3, 4}},
	{{1, 8}, {2, 7}, {3, 6}, {4, 5}},
}

type Generator_Meistriliiga struct {
	store *Store
	Generator_Base
}

func CreateGeneratorMeistriliiga(database *Store) *Generator_Meistriliiga {
	return &Generator_Meistriliiga{
		store: database,
	}
}

func (g *Generator_Meistriliiga) CreateMatches(tournament *model.Tournament, tx *gorm.DB) error {
	// GENERATING TIMETABLE INFORMATION
	//LOGIC -->
	// Query all the teams in order
	// Make the first matches until (gameday[2] (third gameday)) / REGROUPING
	//game days first one is day, second is how many games

	all_teams, err := g.store.Club().GetTournamentTeams(tournament.ID)
	if err != nil {
		return err
	}

	// fmt.Println("get all teams", all_teams)

	//Add check to test if there are exactly 8 teams THINK SOMETHING IF LESS MOST PROBABLY NO
	if len(all_teams) != 8 {
		return errors.New("only 8 teams allowed")
	}

	if tournament.State == "created" {
		//for development
		tournament.State = "started"
		if err := tx.Model(&model.Tournament{}).Where("id = ?", tournament.ID).Updates(&tournament).Error; err != nil {
			return err
		}
	}

	if tournament.State != "started" {
		return errors.New("turniiril ei ole õige staatus")
	}

	gameDay := 0
	roundNumber := 1
	for i, round := range MeistriLiigaMatches {
		if i == 0 || i == 3 || i == 6 {
			gameDay++
		}
		for index, match := range round {
			team1, err := FindTeam(match[0]-1, all_teams)
			if err != nil {
				return err
			}
			team2, err := FindTeam(match[1]-1, all_teams)
			if err != nil {
				return err
			}
			year, month, day := tournament.StartDate.Date()
			var start_date time.Time
			if roundNumber == 4 || roundNumber == 7 {
				start_date = time.Date(year, month, day+gameDay-1, 10+((roundNumber/gameDay-3)*3), 0, 0, 0, tournament.StartDate.Location())
			} else {
				start_date = time.Date(year, month, day+gameDay-1, 10+((roundNumber/gameDay-2)*3), 0, 0, 0, tournament.StartDate.Location())
			}

			matchToCreate := model.Match{
				TournamentID: tournament.ID,
				P1ID:         team1.ID,
				P2ID:         team2.ID,
				Type:         "voor",
				CurrentRound: roundNumber, //Vooru nr
				Identifier:   gameDay,     //Game day number
				StartDate:    start_date,
				Table:        index + 1,
			}
			// fmt.Println(matchToCreate.Type, matchToCreate.CurrentRound, ": ", team1.Name, " VS ", team2.Name)
			if err := tx.Create(&matchToCreate).Error; err != nil {
				return err
			}
		}
		roundNumber++
	}

	if err := g.UpdateTournamentState(tournament, tx); err != nil {
		return err
	}
	return nil
}

func (g *Generator_Meistriliiga) CreateBrackets(tournament *model.Tournament, tx *gorm.DB) error {
	return nil
}

func (g *Generator_Meistriliiga) LinkMatches(tournament *model.Tournament, tx *gorm.DB) error {
	return nil
}

func (g *Generator_Meistriliiga) AssignPlayers(tournament *model.Tournament, tx *gorm.DB) error {
	return nil
}

func FindTeam(order int, teams []model.Team) (*model.Team, error) {
	for _, t := range teams {
		if t.TeamOrder == order {
			return &t, nil
		}
	}
	return nil, fmt.Errorf("no team found")
}

//FOR LATER PURPOSESE
// func generateRounds(participants []int) [][][2]int {
// 	n := len(participants)
// 	numRounds := n - 1                    // Total rounds for round-robin scheduling
// 	rounds := make([][][2]int, numRounds) // Store rounds of pairs

// 	// Iterate over rounds
// 	for round := 0; round < numRounds; round++ {
// 		var roundPairs [][2]int

// 		// Create pairs by pairing first element with the last
// 		roundPairs = append(roundPairs, [2]int{participants[0], participants[len(participants)-1]})

// 		// Pair the rest of the participants
// 		for i := 1; i < n/2; i++ {
// 			roundPairs = append(roundPairs, [2]int{participants[i], participants[n-i-1]})
// 		}

// 		// Store the pairs for the round
// 		rounds[round] = roundPairs

// 		// Rotate participants but keep the first one fixed
// 		participants = append(participants[:1], append([]int{participants[n-1]}, participants[1:n-1]...)...)
// 	}

// 	return rounds
// }
