package psqlstore

import (
	"errors"
	"fmt"
	"table-tennis/internal/model"
	"table-tennis/internal/store/utils"

	"gorm.io/gorm"
)

type ClubRepository struct {
	store *Store
}

func (c *ClubRepository) CreateTeam(team model.Team) error {
	tx := c.store.Db.Begin()
	if tx.Error != nil {
		return tx.Error
	}
	if result := tx.Create(&team); result.Error != nil {
		tx.Rollback()
		return result.Error
	}

	tournament, err := c.store.Tournament().Get(team.TournamentID)
	if err != nil {
		return err
	}
	if tournament.State != "created" {
		return errors.New("turniir on juba alustatud")
	}
	//add players
	for _, player := range team.Players {
		//if player is not in database yet (ID == 0) then add it to database and create link
		var newPlayerJoin model.PlayerJoinTeam
		if player.ID == 0 {
			user := model.User{
				FirstName: player.FirstName,
				LastName:  player.LastName,
			}
			if err := tx.Create(&user).Error; err != nil {
				tx.Rollback()
				return err
			}
			newPlayerJoin = model.PlayerJoinTeam{
				PlayerID: user.ID,
				TeamID:   team.ID,
			}
		} else {
			newPlayerJoin = model.PlayerJoinTeam{
				PlayerID: player.ID,
				TeamID:   team.ID,
			}
		}
		if err := tx.Create(&newPlayerJoin).Error; err != nil {
			tx.Rollback()
			return err
		}
	}
	return tx.Commit().Error
}

func (c *ClubRepository) DeleteTeamMatch(team_match_id uint) error {
	var teamMatchToDelete model.TeamMatch
	result := c.store.Db.Model(&model.TeamMatch{}).Where("id = ?", team_match_id).Find(&teamMatchToDelete)

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return errors.New("mängu ei leitud")
	}

	//begin db transactions
	tx := c.store.Db.Begin()
	if tx.Error != nil {
		return tx.Error
	}
	//reset winner id head referee and table referee if neccessary
	if err := tx.Model(&model.Match{}).Where("id = ?", teamMatchToDelete.MatchID).
		Updates(map[string]interface{}{
			"head_referee":  "",
			"table_referee": "",
			"winner_id":     0,
		}).Error; err != nil {
		tx.Rollback()
		return err
	}
	//query all the matches which needs to be deleted + sets
	team_matches, err := c.GetTeamMatches(teamMatchToDelete.TournamentID, teamMatchToDelete.ID)
	if err != nil {
		return err
	}
	//delete all the sets
	for _, m := range team_matches {
		if err := tx.Unscoped().Where("match_id = ?", m.ID).Delete(&model.Set{}).Error; err != nil {
			tx.Rollback()
			return err
		}
	}
	//delete all the matches
	if err := tx.Unscoped().Where("team_match_id = ?", teamMatchToDelete.ID).Delete(&model.Match{}).Error; err != nil {
		tx.Rollback()
		return err
	}
	//delete team_match
	if err := tx.Unscoped().Where("id = ?", team_match_id).Delete(&model.TeamMatch{}).Error; err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit().Error
}

func (c *ClubRepository) DeleteTeam(team_id uint, tournament_id uint) error {
	var tournament model.Tournament
	if err := c.store.Db.First(&tournament, "id = ?", tournament_id).Error; err != nil {
		return err
	}

	if tournament.State != "created" {
		return fmt.Errorf("can not change teams after creating matches")
	}
	tx := c.store.Db.Begin()
	if tx.Error != nil {
		return tx.Error
	}

	//delete all the players in the team
	if err := tx.Unscoped().Where("team_id = ?", team_id).Delete(&model.PlayerJoinTeam{}).Error; err != nil {
		tx.Rollback()
		return err
	}
	//delete the team
	if err := tx.Unscoped().Where("id = ? AND tournament_id = ?", team_id, tournament_id).Delete(&model.Team{}).Error; err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit().Error
}

func (c *ClubRepository) UpdateTeam(team model.Team) error {
	tx := c.store.Db.Begin()
	if tx.Error != nil {
		return tx.Error
	}

	//check if tournament matches have been created
	var tournament model.Tournament
	if err := c.store.Db.First(&tournament, "id = ?", team.TournamentID).Error; err != nil {
		return err
	}

	// if tournament.State != "created" {
	// 	return fmt.Errorf("can not change teams after creating matches")
	// }

	var existingPlayers []model.PlayerWithTeam
	if err := tx.Model(&model.PlayerJoinTeam{}).
		Joins("JOIN users ON users.id = player_join_teams.player_id").
		Where("player_join_teams.team_id = ?", team.ID).
		Select("player_join_teams.team_id, player_join_teams.player_id, player_join_teams.confirmation, player_join_teams.has_rating, users.first_name, users.last_name").
		Scan(&existingPlayers).Error; err != nil {
		tx.Rollback()
		return err
	}

	var newPlayers []model.PlayerWithTeam
	for _, player := range team.Players {
		newPlayers = append(newPlayers, model.PlayerWithTeam{
			PlayerID:     player.ID,
			FirstName:    player.FirstName,
			LastName:     player.LastName,
			HasRating:    player.HasRating,
			Confirmation: player.Confirmation,
		})
		if player.ID != 0 {
			if err := tx.Model(&model.PlayerJoinTeam{}).Where("player_id = ?", player.ID).Updates(
				map[string]interface{}{
					"confirmation": player.Confirmation,
				},
			).Error; err != nil {
				tx.Rollback()
				return err
			}
		}
	}

	playersToAdd := utils.Difference(newPlayers, existingPlayers)
	playersToRemove := utils.Difference(existingPlayers, newPlayers)
	for _, playerID := range playersToAdd {
		if playerID.PlayerID == 0 {
			user := model.User{
				FirstName: playerID.FirstName,
				LastName:  playerID.LastName,
				//HERE ARE OING
			}
			if err := tx.Create(&user).Error; err != nil {
				tx.Rollback()
				return err
			}

			newPlayerJoin := model.PlayerJoinTeam{
				PlayerID: user.ID,
				TeamID:   team.ID,
			}
			if err := tx.Create(&newPlayerJoin).Error; err != nil {
				tx.Rollback()
				return err
			}
		} else {
			newPlayerJoin := model.PlayerJoinTeam{
				PlayerID: playerID.PlayerID,
				TeamID:   team.ID,
			}
			if err := tx.Create(&newPlayerJoin).Error; err != nil {
				tx.Rollback()
				return err
			}
		}
	}

	if len(playersToRemove) > 0 {
		var idArr []uint
		for _, p := range playersToRemove {
			idArr = append(idArr, p.PlayerID)
		}
		if err := tx.Unscoped().Where("team_id = ? AND player_id IN ?", team.ID, idArr).
			Delete(&model.PlayerJoinTeam{}).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	if err := tx.Model(&model.Team{}).
		Where("id = ?", team.ID).
		Updates(map[string]interface{}{
			"name":       team.Name,
			"captain":    team.Captain,
			"team_order": team.TeamOrder,
		}).Error; err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

func (c *ClubRepository) UpdateTeams(teams []model.Team, tournament_id uint) error {
	tx := c.store.Db.Begin()
	if tx.Error != nil {
		return tx.Error
	}
	for _, team := range teams {
		team.TournamentID = tournament_id
		if err := c.UpdateTeam(team); err != nil {
			tx.Rollback()
			return err
		}
	}

	return nil
}

func (c *ClubRepository) GetTournamentTeams(tournamentID uint) ([]model.Team, error) {
	var teams []model.Team

	if err := c.store.Db.Model(&model.Team{}).
		Where("tournament_id = ?", tournamentID).
		Order("team_order ASC").
		Find(&teams).Error; err != nil {
		return nil, err
	}

	for i, team := range teams {
		var players []model.User

		if err := c.store.Db.Table("player_join_teams").
			Joins("inner join users on users.id = player_join_teams.player_id").
			Where("player_join_teams.team_id = ?", team.ID).
			Select("users.id, users.first_name, users.last_name, users.email, users.birth_date, users.eltl_id, users.club_id, users.has_rating, users.rating_points, users.nationality, users.placing_order, users.image_url, player_join_teams.confirmation").
			Order("CASE WHEN users.nationality = 'EST' THEN 1 ELSE 2 END, users.placing_order ASC").
			Scan(&players).Error; err != nil {
			return nil, err
		}
		teams[i].Players = players
	}
	// fmt.Printf("%+v", teams)
	return teams, nil
}

func (c *ClubRepository) GetTournamentTeamsRegroupted(tournamentID uint, tx *gorm.DB) ([]model.Team, error) {
	var teams []model.Team
	if tx == nil {
		tx = c.store.Db.Begin()
	}
	rows, err := tx.Raw(`
    SELECT 
        t.id,
        t.name,
        t.captain,
        t.tournament_id,
        t.team_order,
        SUM(         
            CASE
                WHEN m.winner_id = t.id THEN 2 
                WHEN (m.p1_id = t.id AND m.winner_id != 0) OR (m.p2_id = t.id AND m.winner_id != 0) THEN 1 
                ELSE 0                                                                                     
            END           
        ) AS total_points
    FROM                     
        matches m
    LEFT JOIN 
        teams t
    ON             
        t.id = m.p1_id OR t.id = m.p2_id
    WHERE                                   
        m.tournament_id = ? AND t.tournament_id = ?
    GROUP BY                                        
        t.id, t.name, t.captain, t.tournament_id, t.team_order
ORDER BY 
    total_points DESC, RANDOM();
	`, tournamentID, tournamentID).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var team model.Team
		var totalPoints int
		if err := rows.Scan(&team.ID, &team.Name, &team.Captain, &team.TournamentID, &team.TeamOrder, &totalPoints); err != nil {
			return nil, err
		}

		team.TotalPoints = totalPoints
		teams = append(teams, team)
	}

	return teams, tx.Commit().Error
}

func (c *ClubRepository) GetTournamentTeam(teamID uint) (*model.Team, error) {
	var players []model.User
	var team model.Team

	//query team
	if err := c.store.Db.Model(&model.Team{}).Where("id = ?", teamID).First(&team).Error; err != nil {
		return nil, err
	}

	if err := c.store.Db.Table("player_join_teams").
		Joins("inner join users on users.id = player_join_teams.player_id").
		Where("player_join_teams.team_id = ?", teamID).
		Select("users.id, users.first_name, users.last_name, users.email").
		Scan(&players).Error; err != nil {
		return nil, err
	}

	team.Players = players
	return &team, nil

}

func (c *ClubRepository) GetTeamMatch(matchID uint) (*model.TeamMatch, error) {
	var teamMatch model.TeamMatch
	if err := c.store.Db.Where("match_id = ?", matchID).First(&teamMatch).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("no team match found with matchID: %d", matchID)
		}
		return nil, err
	}
	return &teamMatch, nil
}

func (c *ClubRepository) CreateTeamMatch(teamMatch model.TeamMatch) (*model.TeamMatch, error) {
	//create a teamMatch and make matches for each p v p match
	tx := c.store.Db.Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}

	var tournament model.Tournament
	if err := c.store.Db.Model(&model.Tournament{}).Where("id = ?", teamMatch.TournamentID).Find(&tournament).Error; err != nil {
		return nil, err
	}
	if tournament.State != "matches_created" {
		return nil, fmt.Errorf("wanted tournament state - %v have - %v", "matches_created", tournament.State)
	}

	if err := tx.Create(&teamMatch).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	matches := []model.Match{
		{TournamentID: teamMatch.TournamentID, P1ID: teamMatch.PlayerAID, P2ID: teamMatch.PlayerYID, Type: "1v1", TeamMatchID: teamMatch.ID, Identifier: 0},
		{TournamentID: teamMatch.TournamentID, P1ID: teamMatch.PlayerBID, P2ID: teamMatch.PlayerXID, Type: "1v1", TeamMatchID: teamMatch.ID, Identifier: 1},
		{TournamentID: teamMatch.TournamentID, P1ID: teamMatch.PlayerCID, P2ID: teamMatch.PlayerZID, Type: "1v1", TeamMatchID: teamMatch.ID, Identifier: 2},
		{TournamentID: teamMatch.TournamentID, P1ID: teamMatch.PlayerAID, P2ID: teamMatch.PlayerXID, Type: "1v1", TeamMatchID: teamMatch.ID, Identifier: 4},
		{TournamentID: teamMatch.TournamentID, P1ID: teamMatch.PlayerCID, P2ID: teamMatch.PlayerYID, Type: "1v1", TeamMatchID: teamMatch.ID, Identifier: 5},
		{TournamentID: teamMatch.TournamentID, P1ID: teamMatch.PlayerBID, P2ID: teamMatch.PlayerZID, Type: "1v1", TeamMatchID: teamMatch.ID, Identifier: 6},
	}

	// Save each match in the database
	for _, match := range matches {
		if err := tx.Create(&match).Error; err != nil {
			tx.Rollback()
			return nil, err
		}
		for i := 0; i < 5; i++ {
			if err := tx.Create(&model.Set{MatchID: match.ID, SetNumber: i + 1}).Error; err != nil {
				tx.Rollback()
				return nil, err
			}
		}
	}

	//check if we have conditions to create 2v2 match also
	if teamMatch.PlayerDID != 0 && teamMatch.PlayerEID != 0 && teamMatch.PlayerVID != 0 && teamMatch.PlayerWID != 0 {
		// Check if a 2v2 match already exists for this TeamMatch
		var existing2v2Match model.Match
		matchExists := tx.Where("team_match_id = ? AND type = ?", teamMatch.ID, "2v2").First(&existing2v2Match).Error == nil

		if matchExists {
			// Update the existing 2v2 match with new players
			fmt.Println("Updating 2v2 match with new players...")
			update2v2Fields := map[string]interface{}{
				"P1ID":   teamMatch.PlayerDID,
				"P1ID_2": teamMatch.PlayerEID,
				"P2ID":   teamMatch.PlayerVID,
				"P2ID_2": teamMatch.PlayerWID,
			}
			if err := tx.Model(&existing2v2Match).Updates(update2v2Fields).Error; err != nil {
				tx.Rollback()
				return nil, err
			}
		} else {
			// Create a new 2v2 match
			fmt.Println("Creating a new 2v2 match...")
			twoVTwoMatch := model.Match{
				TournamentID: teamMatch.TournamentID,
				P1ID:         teamMatch.PlayerDID, // Team 1: Player D and Player E
				P1ID_2:       teamMatch.PlayerEID,
				P2ID:         teamMatch.PlayerVID, // Team 2: Player V and Player W
				P2ID_2:       teamMatch.PlayerWID,
				Type:         "2v2", // Match type is 2v2
				TeamMatchID:  teamMatch.ID,
				Identifier:   3,
			}
			if err := tx.Create(&twoVTwoMatch).Error; err != nil {
				tx.Rollback()
				return nil, err
			}
			//create sets for matches
			for i := 0; i < 5; i++ {
				if err := tx.Create(&model.Set{MatchID: twoVTwoMatch.ID, SetNumber: i + 1}).Error; err != nil {
					tx.Rollback()
					return nil, err
				}
			}
		}
	}

	return &teamMatch, tx.Commit().Error
}

func (c *ClubRepository) UpdateTeamMatch(teamMatch model.TeamMatch) error {
	var existingTeamMatch model.TeamMatch
	if err := c.store.Db.First(&existingTeamMatch, teamMatch.ID).Error; err != nil {
		return err
	}
	var matchInfo model.Match
	if err := c.store.Db.First(&matchInfo, teamMatch.MatchID).Error; err != nil {
		return err
	}
	//if match is finished, no longer able to update
	// if matchInfo.WinnerID != 0 {
	// 	return errors.New("match already finisehd")
	// }
	updateFields := map[string]interface{}{
		"WinnerID":  teamMatch.WinnerID,
		"PlayerAID": teamMatch.PlayerAID,
		"PlayerBID": teamMatch.PlayerBID,
		"PlayerCID": teamMatch.PlayerCID,
		"PlayerXID": teamMatch.PlayerXID,
		"PlayerYID": teamMatch.PlayerYID,
		"PlayerZID": teamMatch.PlayerZID,

		"PlayerVID": teamMatch.PlayerVID,
		"PlayerWID": teamMatch.PlayerWID,
		"PlayerEID": teamMatch.PlayerEID,
		"PlayerDID": teamMatch.PlayerDID,
		"CaptainA":  teamMatch.CaptainA,
		"CaptainB":  teamMatch.CaptainB,
		"Notes":     teamMatch.Notes,
	}

	// Begin transaction to update the TeamMatch and potentially create a 2v2 Match
	tx := c.store.Db.Begin()
	if tx.Error != nil {
		return tx.Error
	}
	if err := tx.Model(&teamMatch).Updates(updateFields).Error; err != nil {
		tx.Rollback()
		return err
	}

	if teamMatch.PlayerDID != 0 && teamMatch.PlayerEID != 0 && teamMatch.PlayerVID != 0 && teamMatch.PlayerWID != 0 {
		// Check if a 2v2 match already exists for this TeamMatch
		var existing2v2Match model.Match
		matchExists := tx.Where("team_match_id = ? AND type = ?", teamMatch.ID, "2v2").First(&existing2v2Match).Error == nil

		if matchExists {
			// Update the existing 2v2 match with new players
			fmt.Println("Updating 2v2 match with new players...")
			update2v2Fields := map[string]interface{}{
				"P1ID":   teamMatch.PlayerDID,
				"P1ID_2": teamMatch.PlayerEID,
				"P2ID":   teamMatch.PlayerVID,
				"P2ID_2": teamMatch.PlayerWID,
			}
			if err := tx.Model(&existing2v2Match).Updates(update2v2Fields).Error; err != nil {
				tx.Rollback()
				return err
			}
		} else {
			// Create a new 2v2 match
			fmt.Println("Creating a new 2v2 match...")
			twoVTwoMatch := model.Match{
				TournamentID: teamMatch.TournamentID,
				P1ID:         teamMatch.PlayerDID, // Team 1: Player D and Player E
				P1ID_2:       teamMatch.PlayerEID,
				P2ID:         teamMatch.PlayerVID, // Team 2: Player V and Player W
				P2ID_2:       teamMatch.PlayerWID,
				Type:         "2v2", // Match type is 2v2
				TeamMatchID:  teamMatch.ID,
				Identifier:   3,
			}
			if err := tx.Create(&twoVTwoMatch).Error; err != nil {
				tx.Rollback()
				return err
			}
			//create sets for matches
			for i := 0; i < 5; i++ {
				if err := tx.Create(&model.Set{MatchID: twoVTwoMatch.ID, SetNumber: i + 1}).Error; err != nil {
					tx.Rollback()
					return err
				}
			}
		}
	}

	return tx.Commit().Error
}

func (c *ClubRepository) GetTeamMatches(tournament_id, team_match_id uint) ([]model.Match, error) {
	var output []model.Match
	if err := c.store.Db.Model(&model.Match{}).Where("tournament_id = ? AND team_match_id = ?", tournament_id, team_match_id).Order("identifier ASC").Find(&output).Error; err != nil {
		return nil, err
	}
	return output, nil
}

func (c *ClubRepository) GetAllTeamMatches(tournament_id, team_id uint) ([]model.Match, error) {
	var output []model.Match
	if err := c.store.Db.Model(&model.Match{}).Where("tournament_id = ? AND (p1_id = ? OR p2_id = ?)", tournament_id, team_id, team_id).Find(&output).Error; err != nil {
		return nil, err
	}
	return output, nil
}
