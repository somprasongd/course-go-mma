package main

import (
	"fmt"
	"go-mma/shared/common/sanitizer"
)

type User struct {
	Email     string `json:"email" sanitize:"email"`
	Name      string `json:"name" sanitize:"xss"`
	Bio       string `json:"bio" sanitize:"trim"`
	Username  string `json:"username" sanitize:"lower"`
	Age       int    `json:"age"` // No sanitize tag, will be ignored
	Protected string `json:"protected"` // No sanitize tag, will be ignored
}

func main() {
	// Example with XSS attacks
	user := User{
		Email:     "<script>alert('xss')</script>user@example.com",
		Name:      "John <script>alert('hack')</script> Doe",
		Bio:       "   Software Engineer   ",
		Username:  "JOHNDOE",
		Age:       30,
		Protected: "<b>This won't be sanitized</b>",
	}

	fmt.Println("Before sanitization:")
	fmt.Printf("  Email: %s\n", user.Email)
	fmt.Printf("  Name: %s\n", user.Name)
	fmt.Printf("  Bio: '%s'\n", user.Bio)
	fmt.Printf("  Username: %s\n", user.Username)
	fmt.Printf("  Age: %d\n", user.Age)
	fmt.Printf("  Protected: %s\n", user.Protected)

	// Apply struct-based sanitization
	sanitizer.SanitizeStruct(&user)

	fmt.Println("\nAfter sanitization:")
	fmt.Printf("  Email: %s\n", user.Email)
	fmt.Printf("  Name: %s\n", user.Name)
	fmt.Printf("  Bio: '%s'\n", user.Bio)
	fmt.Printf("  Username: %s\n", user.Username)
	fmt.Printf("  Age: %d\n", user.Age)
	fmt.Printf("  Protected: %s\n", user.Protected)
}