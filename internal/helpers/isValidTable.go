package helpers

func IsValidTable(tableName string, validNames map[string]string) bool {
	_, ok := validNames[tableName]
	return ok
}
