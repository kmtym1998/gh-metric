package usecase

import "strings"

// ParseTargetUsers parses the comma-separated target users string
func ParseTargetUsers(targetUsers string) []string {
	if targetUsers == "" {
		return nil
	}

	users := strings.Split(targetUsers, ",")
	result := make([]string, 0, len(users))
	for _, user := range users {
		if trimmed := strings.TrimSpace(user); trimmed != "" {
			result = append(result, trimmed)
		}
	}

	return result
}

// IsBot determines if a username belongs to a bot
func IsBot(username string) bool {
	username = strings.ToLower(username)

	// Common bot patterns
	botPatterns := []string{
		"copilot",
		"bot",
		"github-actions",
		"dependabot",
		"renovate",
		"codecov",
		"sonarcloud",
		"snyk",
		"whitesource",
		"greenkeeper",
		"imgbot",
		"allcontributors",
	}

	for _, pattern := range botPatterns {
		if strings.Contains(username, pattern) {
			return true
		}
	}

	// Check for common bot naming patterns
	if strings.HasSuffix(username, "[bot]") {
		return true
	}

	return false
}
