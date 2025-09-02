package helpers

func IsAdmin(userId string, admins []string) bool {
	for _, adminId := range admins {
		if userId == adminId {
			return true
		}
	}
	return false
}
