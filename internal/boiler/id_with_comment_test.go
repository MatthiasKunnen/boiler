package boiler_test

import (
	"github.com/MatthiasKunnen/boiler/internal/boiler"
	"github.com/go-json-experiment/json"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCommentedId_Marshal(t *testing.T) {
	tests := []struct {
		name     string
		input    boiler.IdWithComment
		expected string
	}{
		{
			name:     "with comment",
			input:    boiler.IdWithComment{Id: 123, Comment: "test comment"},
			expected: `[123,"test comment"]`,
		},
		{
			name:     "empty comment",
			input:    boiler.IdWithComment{Id: 456, Comment: ""},
			expected: `[456,""]`,
		},
		{
			name:     "zero id",
			input:    boiler.IdWithComment{Id: 0, Comment: "zero id"},
			expected: `[0,"zero id"]`,
		},
		{
			name:     "large id",
			input:    boiler.IdWithComment{Id: 18446744073709551615, Comment: "max uint64"},
			expected: `[18446744073709551615,"max uint64"]`,
		},
		{
			name:     "special characters in comment",
			input:    boiler.IdWithComment{Id: 789, Comment: "comment with \"quotes\" and \n newlines"},
			expected: `[789,"comment with \"quotes\" and \n newlines"]`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual, err := json.Marshal(tt.input)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, string(actual))
		})
	}
}

func TestCommentedId_Unmarshal(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected boiler.IdWithComment
		wantErr  bool
	}{
		{
			name:     "full array format",
			input:    `[123,"test comment"]`,
			expected: boiler.IdWithComment{Id: 123, Comment: "test comment"},
			wantErr:  false,
		},
		{
			name:     "array with just id",
			input:    `[456]`,
			expected: boiler.IdWithComment{Id: 456, Comment: ""},
			wantErr:  false,
		},
		{
			name:     "direct uint64",
			input:    `789`,
			expected: boiler.IdWithComment{Id: 789, Comment: ""},
			wantErr:  false,
		},
		{
			name:     "zero id in array",
			input:    `[0,"zero"]`,
			expected: boiler.IdWithComment{Id: 0, Comment: "zero"},
			wantErr:  false,
		},
		{
			name:     "large id",
			input:    `[18446744073709551615,"max uint64"]`,
			expected: boiler.IdWithComment{Id: 18446744073709551615, Comment: "max uint64"},
			wantErr:  false,
		},
		{
			name:     "empty comment in array",
			input:    `[123,""]`,
			expected: boiler.IdWithComment{Id: 123, Comment: ""},
			wantErr:  false,
		},
		{
			name:     "special characters in comment",
			input:    `[789,"comment with \"quotes\" and \n newlines"]`,
			expected: boiler.IdWithComment{Id: 789, Comment: "comment with \"quotes\" and \n newlines"},
			wantErr:  false,
		},
		{
			name:     "negative number",
			input:    `[-123,"comment"]`,
			expected: boiler.IdWithComment{Id: 0, Comment: "comment"},
		},
		{
			name:     "floating point number",
			input:    `[123.5,"comment"]`,
			expected: boiler.IdWithComment{Id: 123, Comment: "comment"},
		},
		// Error cases
		{
			name:    "empty array",
			input:   `[]`,
			wantErr: true,
		},
		{
			name:    "string as first element",
			input:   `["not a number","comment"]`,
			wantErr: true,
		},
		{
			name:    "number as second element",
			input:   `[123,456]`,
			wantErr: true,
		},
		{
			name:    "string input",
			input:   `"not a number"`,
			wantErr: true,
		},
		{
			name:    "object input",
			input:   `{"id":123,"comment":"test"}`,
			wantErr: true,
		},
		{
			name:    "null input",
			input:   `null`,
			wantErr: true,
		},
		{
			name:    "boolean input",
			input:   `true`,
			wantErr: true,
		},
		{
			name:    "malformed json",
			input:   `[123,`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var result boiler.IdWithComment
			err := json.Unmarshal([]byte(tt.input), &result)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			if err != nil {
				assert.NoError(t, err)
			}

			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCommentedId_RoundTrip(t *testing.T) {
	tests := []boiler.IdWithComment{
		{Id: 0, Comment: ""},
		{Id: 1, Comment: "one"},
		{Id: 123, Comment: "test comment"},
		{Id: 456, Comment: ""},
		{Id: 18446744073709551615, Comment: "max uint64"},
		{Id: 789, Comment: "special chars: \"\n\t\\"},
	}

	for _, original := range tests {
		t.Run("", func(t *testing.T) {
			data, err := json.Marshal(original)
			assert.NoError(t, err)

			var actual boiler.IdWithComment
			err = json.Unmarshal(data, &actual)
			assert.NoError(t, err)
			assert.Equal(t, original, actual)
		})
	}
}

func TestCommentedId_UnmarshalIntoExistingNoMerge(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected boiler.IdWithComment
		wantErr  bool
	}{
		{
			name:     "full array format",
			input:    `[123,"new"]`,
			expected: boiler.IdWithComment{Id: 123, Comment: "new"},
		},
		{
			name:     "array without string format",
			input:    `[123]`,
			expected: boiler.IdWithComment{Id: 123, Comment: ""},
		},
		{
			name:     "ID only format",
			input:    `123`,
			expected: boiler.IdWithComment{Id: 123, Comment: ""},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := boiler.IdWithComment{Id: 999, Comment: "existing"}
			err := json.Unmarshal([]byte(tt.input), &actual)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, actual)
		})
	}
}

// Benchmark tests
func BenchmarkCommentedId_Marshal(b *testing.B) {
	c := boiler.IdWithComment{Id: 123456789, Comment: "benchmark comment"}

	for b.Loop() {
		_, err := json.Marshal(c)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkCommentedId_Unmarshal(b *testing.B) {
	data := []byte(`[123456789,"benchmark comment"]`)

	for b.Loop() {
		var c boiler.IdWithComment
		err := json.Unmarshal(data, &c)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkCommentedId_UnmarshalDirect(b *testing.B) {
	data := []byte(`123456789`)

	for b.Loop() {
		var c boiler.IdWithComment
		err := json.Unmarshal(data, &c)
		if err != nil {
			b.Fatal(err)
		}
	}
}
