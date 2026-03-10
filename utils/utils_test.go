package utils

import (
	"testing"
	"time"
)

func TestToSafeFilename(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"normal_title", "normal_title"},
		{"title/with/slashes", "title_with_slashes"},
		{"illegal<>:\"|?*", "illegal_______"},
		{"  spaces  ", "  spaces  "},
	}

	for _, tt := range tests {
		result := ToSafeFilename(tt.input)
		if result != tt.expected {
			t.Errorf("ToSafeFilename(%s) = %s; want %s", tt.input, result, tt.expected)
		}
	}
}

func TestUniqueStrings(t *testing.T) {
	input := []string{"a", "b", "a", "c", "b", "d"}
	expected := []string{"a", "b", "c", "d"}

	result := UniqueStrings(input)
	if len(result) != len(expected) {
		t.Errorf("Expected length %d, got %d", len(expected), len(result))
	}

	for i, v := range result {
		if v != expected[i] {
			t.Errorf("At index %d: expected %s, got %s", i, expected[i], v)
		}
	}
}

func TestGetExecutionTime(t *testing.T) {
	// 这是一个简单的冒烟测试，确保不崩溃且返回字符串
	start := time.Now()
	end := start.Add(time.Second * 5)
	result := GetExecutionTime(start, end)
	if result == "" {
		t.Error("GetExecutionTime returned empty string")
	}
}
