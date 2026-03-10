package sanitizer

import (
	"testing"
)

func TestSanitizeEmail(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Normal email",
			input:    "test@example.com",
			expected: "test@example.com",
		},
		{
			name:     "Email with script tag",
			input:    "<script>alert('xss')</script>test@example.com",
			expected: "alert('xss')test@example.com",
		},
		{
			name:     "Email with HTML tags",
			input:    "<b>test</b>@example.com",
			expected: "test@example.com",
		},
		{
			name:     "Email with whitespace",
			input:    "  test@example.com  ",
			expected: "test@example.com",
		},
		{
			name:     "Email with mixed case",
			input:    "Test@Example.Com",
			expected: "test@example.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SanitizeEmail(tt.input)
			if result != tt.expected {
				t.Errorf("SanitizeEmail(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestSanitizeInput(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Normal text",
			input:    "Hello World",
			expected: "Hello World",
		},
		{
			name:     "Text with script tag",
			input:    "Hello <script>alert('xss')</script>World",
			expected: "Hello World",
		},
		{
			name:     "Text with iframe tag",
			input:    "Hello <iframe src='javascript:alert(1)'></iframe>World",
			expected: "Hello World",
		},
		{
			name:     "Text with javascript URL",
			input:    "Click here: javascript:alert('xss')",
			expected: "Click here: alert(&#39;xss&#39;)",
		},
		{
			name:     "Text with event handler",
			input:    "<img src='test.jpg' onload='alert(1)'>",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SanitizeInput(tt.input)
			if result != tt.expected {
				t.Errorf("SanitizeInput(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}