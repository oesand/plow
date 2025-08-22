package prm

import "fmt"

// Min creates a condition that validates a numeric value is greater than or equal to the minimum.
func Min[T NumericTypes](min T) Condition[T] {
	return &minCond[T]{min: min}
}

type minCond[T NumericTypes] struct {
	min T
}

func (c *minCond[T]) Validate(value T) error {
	if value < c.min {
		return fmt.Errorf("value must be >= %v", c.min)
	}
	return nil
}

// Max creates a condition that validates a numeric value is less than or equal to the maximum.
func Max[T NumericTypes](max T) Condition[T] {
	return &maxCond[T]{max: max}
}

type maxCond[T NumericTypes] struct {
	max T
}

func (c *maxCond[T]) Validate(value T) error {
	if value > c.max {
		return fmt.Errorf("value must be <= %v", c.max)
	}
	return nil
}
