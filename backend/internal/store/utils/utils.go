package utils

import (
	"math"
	"strconv"
	"strings"
	"table-tennis/internal/model"
)

func RoundUpToPowerOf2(n int) int {
	if n <= 0 {
		panic("Input must be a positive integer")
	}

	if n == 1 {
		return 2
	}

	if n&(n-1) == 0 {
		return n
	}

	mask := n
	mask |= mask >> 1
	mask |= mask >> 2
	mask |= mask >> 4
	mask |= mask >> 8
	mask |= mask >> 16

	return mask + 1
}

func CalcRoundsSingle(n int) int {
	//it only could be 2, 4, 8, 16, 32, 64 and so on
	n = RoundUpToPowerOf2(n)
	return int(math.Ceil(math.Log2(float64(n))))
}

// needs refactor later
func CalcRoundsDouble(n int) int {
	return CalcRoundsSingle(n) + 1
}

func CalculateMatchesPerRoundSingle(n int) map[int]int {
	totalRounds := CalcRoundsSingle(n)
	totalPlayers := RoundUpToPowerOf2(n)

	winnersMatchesPerRound := make(map[int]int)
	j := 2
	for i := 1; i <= totalRounds; i++ {
		winnersMatchesPerRound[i] = totalPlayers / j
		j *= 2
	}
	return winnersMatchesPerRound
}

func CalculateMatchesPerRoundDouble(n int) (map[int]int, map[int]int) {
	winnerRounds := CalculateMatchesPerRoundSingle(n)
	loserRoundCount := len(winnerRounds) + (len(winnerRounds) - 3)
	loserRounds := make(map[int]int)
	divider := 2
	for i := 1; i <= loserRoundCount; i++ {
		loserRounds[i] = (RoundUpToPowerOf2(n) / 2) / divider
		if i%2 == 0 {
			divider *= 2
		}
	}

	return winnerRounds, loserRounds
}

func CalculateMatchesPerRoundDoubleWeird(n int) (map[int]int, map[int]int) {
	winnerRounds, loserRounds := CalculateMatchesPerRoundDouble(n)
	//add 3-4 place match to winners table
	winnerRounds[len(winnerRounds)]++

	//remove 3-4 place player's match who are going to winners table
	delete(loserRounds, len(loserRounds)-1)
	loserRounds[len(loserRounds)] = loserRounds[len(loserRounds)+1]
	delete(loserRounds, len(loserRounds))

	return winnerRounds, loserRounds
}

func Branch(seed, level, limit int) [][2]int {
	var pairs [][2]int

	levelSum := int(math.Pow(2, float64(level))) + 1

	if limit == level+1 {
		pairs = append(pairs, [2]int{seed, levelSum - seed})
		return pairs
	}

	if seed%2 == 1 {
		pairs = append(pairs, Branch(seed, level+1, limit)...)
		pairs = append(pairs, Branch(levelSum-seed, level+1, limit)...)
	} else {
		pairs = append(pairs, Branch(levelSum-seed, level+1, limit)...)
		pairs = append(pairs, Branch(seed, level+1, limit)...)
	}

	return pairs
}

func ParseBracketString(bracket string) ([2]int, error) {
	if bracket == "" {
		return [2]int{}, nil
	}
	id := strings.Split(bracket, "-")
	first, err := strconv.Atoi(id[0])
	if err != nil {
		return [2]int{}, err
	}
	second, err := strconv.Atoi(id[1])
	if err != nil {
		return [2]int{}, err
	}

	return [2]int{first, second}, nil
}

func Difference(a, b []model.PlayerWithTeam) []model.PlayerWithTeam {
	m := make(map[uint]bool)
	for _, user := range b {
		m[user.PlayerID] = true
	}

	var diff []model.PlayerWithTeam
	for _, user := range a {
		if !m[user.PlayerID] {
			diff = append(diff, user)
		}
	}

	return diff
}

func Reverse(round [4][2]int) [4][2]int {
	var reversed [4][2]int
	for i, pair := range round {
		reversed[i] = [2]int{pair[1], pair[0]}
	}
	return reversed
}
