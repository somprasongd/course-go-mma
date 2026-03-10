package sanitizer

import (
	"html"
	"regexp"
	"strings"
)

// SanitizeInput removes or escapes potentially harmful characters and strings
func SanitizeInput(input string) string {
	// Trim whitespace
	s := strings.TrimSpace(input)

	// Remove HTML tags first
	s = removeHTMLTags(s)

	// HTML escape to prevent XSS
	s = html.EscapeString(s)

	return s
}

// removeHTMLTags removes HTML tags from input
func removeHTMLTags(input string) string {
	// Remove script tags and their content
	re := regexp.MustCompile(`(?i)<script[^>]*>.*?</script>`)
	result := re.ReplaceAllString(input, "")

	// Remove iframe tags and their content
	re = regexp.MustCompile(`(?i)<iframe[^>]*>.*?</iframe>`)
	result = re.ReplaceAllString(result, "")

	// Remove object tags and their content
	re = regexp.MustCompile(`(?i)<object[^>]*>.*?</object>`)
	result = re.ReplaceAllString(result, "")

	// Remove embed tags and their content
	re = regexp.MustCompile(`(?i)<embed[^>]*>.*?</embed>`)
	result = re.ReplaceAllString(result, "")

	// Remove all other HTML tags
	re = regexp.MustCompile(`<[^>]*>`)
	result = re.ReplaceAllString(result, "")

	// Remove javascript: and vbscript: URLs
	re = regexp.MustCompile(`(?i)(javascript|vbscript):`)
	result = re.ReplaceAllString(result, "")

	// Remove event handlers
	re = regexp.MustCompile(`(?i)\s+on\w+\s*=`)
	result = re.ReplaceAllString(result, "")

	return result
}

// SanitizeEmail specifically sanitizes email addresses
func SanitizeEmail(email string) string {
	// Remove any HTML tags from email
	re := regexp.MustCompile(`<[^>]*>`)
	sanitized := re.ReplaceAllString(email, "")
	
	// Trim whitespace
	sanitized = strings.TrimSpace(sanitized)
	
	// Convert to lowercase
	sanitized = strings.ToLower(sanitized)
	
	return sanitized
}