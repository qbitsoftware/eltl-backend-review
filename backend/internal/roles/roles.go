package roles

type Role int

const (
	User Role = iota
	MediaUser
	TournamentManager
	Admin
)

func (r Role) String() string {
	switch r {
	case User:
		return "User"
	case MediaUser:
		return "Media user"
	case TournamentManager:
		return "Tournament manager"
	case Admin:
		return "Admin"
	default:
		return "Unknown"
	}
}
