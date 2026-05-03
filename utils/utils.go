package utils

func Truncate(s string, maxLen int) string {
	if len(s) > maxLen {
		return s[:maxLen-1] + "…"
	}
	return s
}
