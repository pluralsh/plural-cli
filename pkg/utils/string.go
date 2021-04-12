package utils

func Pluralize(one, many string, count int) string {
	if count == 1 {
		return one
	}
	return many
}