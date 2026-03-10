package sanitizer

import (
	"testing"
)

type TestStruct struct {
	Email     string `json:"email" sanitize:"email"`
	Name      string `json:"name" sanitize:"xss"`
	TrimField string `json:"trim_field" sanitize:"trim"`
	Lowercase string `json:"lowercase" sanitize:"lower"`
	NoTag     string `json:"no_tag"`
	NonString int    `json:"non_string"`
}

// Nested struct types
type Address struct {
	Street  string `json:"street" sanitize:"xss"`
	City    string `json:"city" sanitize:"trim"`
	Country string `json:"country" sanitize:"lower"`
}

type PersonWithAddress struct {
	Name    string  `json:"name" sanitize:"xss"`
	Email   string  `json:"email" sanitize:"email"`
	Address Address `json:"address"`
}

type PersonWithPointer struct {
	Name    string   `json:"name" sanitize:"xss"`
	Email   string   `json:"email" sanitize:"email"`
	Address *Address `json:"address"`
}

type PersonWithSlice struct {
	Name     string   `json:"name" sanitize:"xss"`
	Tags     []string `json:"tags" sanitize:"trim"`
	Aliases  []string `json:"aliases" sanitize:"lower"`
	Comments []string `json:"comments" sanitize:"xss"`
}

type PersonWithPointerSlice struct {
	Name      string     `json:"name" sanitize:"xss"`
	Addresses []*Address `json:"addresses"`
}

type PersonWithMap struct {
	Name       string            `json:"name" sanitize:"xss"`
	Metadata   map[string]string `json:"metadata"`
	Attributes map[string]interface{} `json:"attributes"`
}

func TestSanitizeStruct(t *testing.T) {
	tests := []struct {
		name     string
		input    TestStruct
		expected TestStruct
	}{
		{
			name: "Normal fields",
			input: TestStruct{
				Email:     "test@example.com",
				Name:      "John Doe",
				TrimField: "  trim me  ",
				Lowercase: "UPPERCASE",
				NoTag:     "<b>no tag</b>",
				NonString: 42,
			},
			expected: TestStruct{
				Email:     "test@example.com",
				Name:      "John Doe",
				TrimField: "trim me",
				Lowercase: "uppercase",
				NoTag:     "<b>no tag</b>",
				NonString: 42,
			},
		},
		{
			name: "Fields with XSS",
			input: TestStruct{
				Email:     "<script>alert('xss')</script>test@example.com",
				Name:      "Hello <script>alert('xss')</script>World",
				TrimField: "  <b>trim</b> me  ",
				Lowercase: "UPPER<b>CASE</b>",
				NoTag:     "<b>no tag</b>",
				NonString: 42,
			},
			expected: TestStruct{
				Email:     "alert('xss')test@example.com",
				Name:      "Hello World",
				TrimField: "<b>trim</b> me",
				Lowercase: "upper<b>case</b>",
				NoTag:     "<b>no tag</b>",
				NonString: 42,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			SanitizeStruct(&tt.input)

			if tt.input.Email != tt.expected.Email {
				t.Errorf("Email = %q, want %q", tt.input.Email, tt.expected.Email)
			}
			if tt.input.Name != tt.expected.Name {
				t.Errorf("Name = %q, want %q", tt.input.Name, tt.expected.Name)
			}
			if tt.input.TrimField != tt.expected.TrimField {
				t.Errorf("TrimField = %q, want %q", tt.input.TrimField, tt.expected.TrimField)
			}
			if tt.input.Lowercase != tt.expected.Lowercase {
				t.Errorf("Lowercase = %q, want %q", tt.input.Lowercase, tt.expected.Lowercase)
			}
			if tt.input.NoTag != tt.expected.NoTag {
				t.Errorf("NoTag = %q, want %q", tt.input.NoTag, tt.expected.NoTag)
			}
			if tt.input.NonString != tt.expected.NonString {
				t.Errorf("NonString = %d, want %d", tt.input.NonString, tt.expected.NonString)
			}
		})
	}
}

func TestSanitizeStructWithNestedStruct(t *testing.T) {
	input := PersonWithAddress{
		Name:  "<script>alert('xss')</script>John",
		Email: "Test@Example.COM",
		Address: Address{
			Street:  "<b>Main St</b>",
			City:    "  Bangkok  ",
			Country: "THAILAND",
		},
	}

	SanitizeStruct(&input)

	if input.Name != "John" {
		t.Errorf("Name = %q, want %q", input.Name, "John")
	}
	if input.Email != "test@example.com" {
		t.Errorf("Email = %q, want %q", input.Email, "test@example.com")
	}
	if input.Address.Street != "Main St" {
		t.Errorf("Address.Street = %q, want %q", input.Address.Street, "Main St")
	}
	if input.Address.City != "Bangkok" {
		t.Errorf("Address.City = %q, want %q", input.Address.City, "Bangkok")
	}
	if input.Address.Country != "thailand" {
		t.Errorf("Address.Country = %q, want %q", input.Address.Country, "thailand")
	}
}

func TestSanitizeStructWithPointer(t *testing.T) {
	input := PersonWithPointer{
		Name:  "<script>alert('xss')</script>John",
		Email: "Test@Example.COM",
		Address: &Address{
			Street:  "<b>Main St</b>",
			City:    "  Bangkok  ",
			Country: "THAILAND",
		},
	}

	SanitizeStruct(&input)

	if input.Name != "John" {
		t.Errorf("Name = %q, want %q", input.Name, "John")
	}
	if input.Email != "test@example.com" {
		t.Errorf("Email = %q, want %q", input.Email, "test@example.com")
	}
	if input.Address == nil {
		t.Fatal("Address should not be nil")
	}
	if input.Address.Street != "Main St" {
		t.Errorf("Address.Street = %q, want %q", input.Address.Street, "Main St")
	}
	if input.Address.City != "Bangkok" {
		t.Errorf("Address.City = %q, want %q", input.Address.City, "Bangkok")
	}
	if input.Address.Country != "thailand" {
		t.Errorf("Address.Country = %q, want %q", input.Address.Country, "thailand")
	}
}

func TestSanitizeStructWithNilPointer(t *testing.T) {
	input := PersonWithPointer{
		Name:    "John",
		Email:   "test@example.com",
		Address: nil,
	}

	// Should not panic
	SanitizeStruct(&input)

	if input.Name != "John" {
		t.Errorf("Name = %q, want %q", input.Name, "John")
	}
	if input.Address != nil {
		t.Errorf("Address should be nil")
	}
}

func TestSanitizeStructWithSlice(t *testing.T) {
	input := PersonWithSlice{
		Name:     "<script>alert('xss')</script>John",
		Tags:     []string{"  tag1  ", "  tag2  ", "tag3"},
		Aliases:  []string{"ALIAS1", "ALIAS2"},
		Comments: []string{"<b>comment1</b>", "<script>alert('xss')</script>"},
	}

	SanitizeStruct(&input)

	if input.Name != "John" {
		t.Errorf("Name = %q, want %q", input.Name, "John")
	}
	if len(input.Tags) != 3 || input.Tags[0] != "tag1" || input.Tags[1] != "tag2" || input.Tags[2] != "tag3" {
		t.Errorf("Tags = %v, want [tag1 tag2 tag3]", input.Tags)
	}
	if len(input.Aliases) != 2 || input.Aliases[0] != "alias1" || input.Aliases[1] != "alias2" {
		t.Errorf("Aliases = %v, want [alias1 alias2]", input.Aliases)
	}
	if len(input.Comments) != 2 || input.Comments[0] != "comment1" || input.Comments[1] != "" {
		t.Errorf("Comments = %v, want [comment1 ]", input.Comments)
	}
}

func TestSanitizeStructWithPointerSlice(t *testing.T) {
	input := PersonWithPointerSlice{
		Name: "<script>alert('xss')</script>John",
		Addresses: []*Address{
			{
				Street:  "<b>Street 1</b>",
				City:    "  City1  ",
				Country: "COUNTRY1",
			},
			{
				Street:  "<i>Street 2</i>",
				City:    "  City2  ",
				Country: "COUNTRY2",
			},
		},
	}

	SanitizeStruct(&input)

	if input.Name != "John" {
		t.Errorf("Name = %q, want %q", input.Name, "John")
	}
	if len(input.Addresses) != 2 {
		t.Fatalf("Expected 2 addresses, got %d", len(input.Addresses))
	}
	if input.Addresses[0].Street != "Street 1" {
		t.Errorf("Addresses[0].Street = %q, want %q", input.Addresses[0].Street, "Street 1")
	}
	if input.Addresses[0].City != "City1" {
		t.Errorf("Addresses[0].City = %q, want %q", input.Addresses[0].City, "City1")
	}
	if input.Addresses[0].Country != "country1" {
		t.Errorf("Addresses[0].Country = %q, want %q", input.Addresses[0].Country, "country1")
	}
	if input.Addresses[1].Street != "Street 2" {
		t.Errorf("Addresses[1].Street = %q, want %q", input.Addresses[1].Street, "Street 2")
	}
}

func TestSanitizeStructWithMap(t *testing.T) {
	input := PersonWithMap{
		Name: "<script>alert('xss')</script>John",
		Metadata: map[string]string{
			"key1": "<b>value1</b>",
			"key2": "<script>alert('xss')</script>",
		},
		Attributes: map[string]interface{}{
			"description": "<b>description</b>",
			"notes":       "<script>alert('xss')</script>note",
		},
	}

	SanitizeStruct(&input)

	if input.Name != "John" {
		t.Errorf("Name = %q, want %q", input.Name, "John")
	}
	if input.Metadata["key1"] != "value1" {
		t.Errorf("Metadata['key1'] = %q, want %q", input.Metadata["key1"], "value1")
	}
	if input.Metadata["key2"] != "" {
		t.Errorf("Metadata['key2'] = %q, want %q", input.Metadata["key2"], "")
	}
	if input.Attributes["description"] != "description" {
		t.Errorf("Attributes['description'] = %q, want %q", input.Attributes["description"], "description")
	}
	if input.Attributes["notes"] != "note" {
		t.Errorf("Attributes['notes'] = %q, want %q", input.Attributes["notes"], "note")
	}
}

func TestSanitizeStructWithNil(t *testing.T) {
	// Should not panic
	SanitizeStruct(nil)

	// Should not panic
	var s *TestStruct
	SanitizeStruct(s)

	// Should not panic
	var i int
	SanitizeStruct(i)

	// Should not panic with nil map
	input := PersonWithMap{
		Name:     "John",
		Metadata: nil,
	}
	SanitizeStruct(&input)
	if input.Name != "John" {
		t.Errorf("Name = %q, want %q", input.Name, "John")
	}

	// Should not panic with nil slice
	input2 := PersonWithSlice{
		Name:     "John",
		Tags:     nil,
		Aliases:  nil,
		Comments: nil,
	}
	SanitizeStruct(&input2)
	if input2.Name != "John" {
		t.Errorf("Name = %q, want %q", input2.Name, "John")
	}
}

func TestSanitizeStructWithEmptySlice(t *testing.T) {
	input := PersonWithSlice{
		Name:     "John",
		Tags:     []string{},
		Aliases:  []string{},
		Comments: []string{},
	}

	// Should not panic
	SanitizeStruct(&input)

	if input.Name != "John" {
		t.Errorf("Name = %q, want %q", input.Name, "John")
	}
	if len(input.Tags) != 0 {
		t.Errorf("Tags should be empty, got %v", input.Tags)
	}
}
