package prm

import (
	"fmt"
	"testing"
)

func TestMinCondition(t *testing.T) {
	tests := []struct {
		name     string
		min      int
		value    int
		expected error
	}{
		{"value equals min", 5, 5, nil},
		{"value greater than min", 5, 10, nil},
		{"value less than min", 5, 3, fmt.Errorf("value must be >= 5")},
		{"value equals min (zero)", 0, 0, nil},
		{"value greater than min (zero)", 0, 1, nil},
		{"value less than min (zero)", 0, -1, fmt.Errorf("value must be >= 0")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			condition := Min(tt.min)
			result := condition.Validate(tt.value)

			if tt.expected == nil {
				if result != nil {
					t.Errorf("expected no error, got %v", result)
				}
			} else {
				if result == nil {
					t.Errorf("expected error %v, got nil", tt.expected)
				} else if result.Error() != tt.expected.Error() {
					t.Errorf("expected error %v, got %v", tt.expected, result)
				}
			}
		})
	}
}

func TestMaxCondition(t *testing.T) {
	tests := []struct {
		name     string
		max      int
		value    int
		expected error
	}{
		{"value equals max", 10, 10, nil},
		{"value less than max", 10, 5, nil},
		{"value greater than max", 10, 15, fmt.Errorf("value must be <= 10")},
		{"value equals max (zero)", 0, 0, nil},
		{"value less than max (zero)", 0, -1, nil},
		{"value greater than max (zero)", 0, 1, fmt.Errorf("value must be <= 0")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			condition := Max(tt.max)
			result := condition.Validate(tt.value)

			if tt.expected == nil {
				if result != nil {
					t.Errorf("expected no error, got %v", result)
				}
			} else {
				if result == nil {
					t.Errorf("expected error %v, got nil", tt.expected)
				} else if result.Error() != tt.expected.Error() {
					t.Errorf("expected error %v, got %v", tt.expected, result)
				}
			}
		})
	}
}

func TestMinConditionWithDifferentTypes(t *testing.T) {
	// Test with int8
	t.Run("int8", func(t *testing.T) {
		condition := Min[int8](5)
		if err := condition.Validate(int8(3)); err == nil {
			t.Error("expected error for value less than min")
		}
		if err := condition.Validate(int8(7)); err != nil {
			t.Error("expected no error for value greater than min")
		}
	})

	// Test with uint
	t.Run("uint", func(t *testing.T) {
		condition := Min[uint](10)
		if err := condition.Validate(uint(5)); err == nil {
			t.Error("expected error for value less than min")
		}
		if err := condition.Validate(uint(15)); err != nil {
			t.Error("expected no error for value greater than min")
		}
	})

	// Test with float64
	t.Run("float64", func(t *testing.T) {
		condition := Min[float64](3.14)
		if err := condition.Validate(2.71); err == nil {
			t.Error("expected error for value less than min")
		}
		if err := condition.Validate(3.14); err != nil {
			t.Error("expected no error for value equal to min")
		}
		if err := condition.Validate(4.0); err != nil {
			t.Error("expected no error for value greater than min")
		}
	})
}

func TestMaxConditionWithDifferentTypes(t *testing.T) {
	// Test with int16
	t.Run("int16", func(t *testing.T) {
		condition := Max[int16](100)
		if err := condition.Validate(int16(150)); err == nil {
			t.Error("expected error for value greater than max")
		}
		if err := condition.Validate(int16(50)); err != nil {
			t.Error("expected no error for value less than max")
		}
	})

	// Test with uint32
	t.Run("uint32", func(t *testing.T) {
		condition := Max[uint32](1000)
		if err := condition.Validate(uint32(1500)); err == nil {
			t.Error("expected error for value greater than max")
		}
		if err := condition.Validate(uint32(500)); err != nil {
			t.Error("expected no error for value less than max")
		}
	})

	// Test with float32
	t.Run("float32", func(t *testing.T) {
		condition := Max[float32](2.5)
		if err := condition.Validate(float32(3.0)); err == nil {
			t.Error("expected error for value greater than max")
		}
		if err := condition.Validate(float32(2.5)); err != nil {
			t.Error("expected no error for value equal to max")
		}
		if err := condition.Validate(float32(1.0)); err != nil {
			t.Error("expected no error for value less than max")
		}
	})
}

func TestMinMaxEdgeCases(t *testing.T) {
	t.Run("min with negative values", func(t *testing.T) {
		condition := Min[int](-10)
		if err := condition.Validate(-15); err == nil {
			t.Error("expected error for value less than negative min")
		}
		if err := condition.Validate(-5); err != nil {
			t.Error("expected no error for value greater than negative min")
		}
	})

	t.Run("max with negative values", func(t *testing.T) {
		condition := Max[int](-5)
		if err := condition.Validate(0); err == nil {
			t.Error("expected error for value greater than negative max")
		}
		if err := condition.Validate(-10); err != nil {
			t.Error("expected no error for value less than negative max")
		}
	})

	t.Run("min with large numbers", func(t *testing.T) {
		condition := Min[int64](9223372036854775807)
		if err := condition.Validate(int64(9223372036854775806)); err == nil {
			t.Error("expected error for value less than max int64")
		}
		if err := condition.Validate(int64(9223372036854775807)); err != nil {
			t.Error("expected no error for value equal to max int64")
		}
	})
}
