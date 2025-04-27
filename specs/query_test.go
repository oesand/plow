package specs

import (
	"reflect"
	"testing"
)

func TestParseQuery(t *testing.T) {
	tests := []struct {
		name     string
		query    string
		expected Query
	}{
		{
			name:  "Single key-value pair",
			query: "key=value",
			expected: Query{
				"key": "value",
			},
		},
		{
			name:  "Multiple key-value pairs",
			query: "key1=value1&key2=value2",
			expected: Query{
				"key1": "value1",
				"key2": "value2",
			},
		},
		{
			name:  "Key with multiple values",
			query: "key=value1&key=value2",
			expected: Query{
				"key": "value1",
			},
		},
		{
			name:     "Empty query string",
			query:    "",
			expected: Query{},
		},
		{
			name:  "Key without value",
			query: "key=",
			expected: Query{
				"key": "",
			},
		},
		{
			name:  "Query with special characters",
			query: "key=val%20ue&anotherKey=hello%20world",
			expected: Query{
				"key":        "val ue",
				"anotherKey": "hello world",
			},
		},
		{
			name:     "Invalid query with missing key",
			query:    "=value",
			expected: Query{},
		},
		{
			name:  "Query with spaces",
			query: "key=value with spaces",
			expected: Query{
				"key": "value with spaces",
			},
		},
		{
			name:  "Query with encoded special characters",
			query: "key=hello%2Cworld",
			expected: Query{
				"key": "hello,world",
			},
		},
		{
			name:  "Complex query with multiple params",
			query: "user=alice&age=30&hobbies=reading&hobbies=swimming",
			expected: Query{
				"user":    "alice",
				"age":     "30",
				"hobbies": "reading",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query := ParseQuery(tt.query)
			if !reflect.DeepEqual(query, tt.expected) {
				t.Errorf("ParseQuery() gotQuery = %v, want %v", query, tt.expected)
			}
		})
	}

}

func TestQuery_String(t *testing.T) {
	tests := []struct {
		name     string
		query    Query
		expected string
	}{
		{
			name:     "Single key-value pair",
			query:    Query{"key": "value"},
			expected: "key=value",
		},
		{
			name: "Multiple key-value pairs",
			query: Query{
				"key1": "value1",
				"key2": "value2",
			},
			expected: "key1=value1&key2=value2",
		},
		{
			name:     "Empty query",
			query:    Query{},
			expected: "",
		},
		{
			name:     "Key with empty value",
			query:    Query{"key": ""},
			expected: "key=",
		},
		{
			name: "Query with special characters",
			query: Query{
				"key":        "value with spaces",
				"anotherKey": "hello, world",
			},
			expected: "key=value+with+spaces&anotherKey=hello%2C+world",
		},
		{
			name: "Query with URL-encoded values",
			query: Query{
				"key":        "value%20with%20spaces",
				"anotherKey": "hello%2Cworld",
			},
			expected: "key=value%2520with%2520spaces&anotherKey=hello%252Cworld",
		},
		{
			name: "Query with empty string values",
			query: Query{
				"key1": "",
				"key2": "",
			},
			expected: "key1=&key2=",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.query.String(); got != tt.expected {
				t.Errorf("String() = %v, want %v", got, tt.expected)
			}
		})
	}
}
