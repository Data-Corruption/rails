package utils

import (
	"bufio"
	"strings"
)

var Version string

type LineCallback func(index int, line string) error

// ForEachLine calls the callback function for each line in the given string.
func ForEachLine(s string, callback LineCallback) error {
	scanner := bufio.NewScanner(strings.NewReader(s))
	scanner.Split(bufio.ScanLines)
	for i := 0; scanner.Scan(); i++ {
		if err := callback(i, scanner.Text()); err != nil {
			return err
		}
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	return nil
}

// Contains checks if a slice contains a given element.
func Contains[T comparable](element T, slice []T) bool {
	for _, sliceElement := range slice {
		if element == sliceElement {
			return true
		}
	}
	return false
}

// ContainsAny checks if a slice contains any of the elements in another slice.
func ContainsAny[T comparable](elements []T, slice []T) bool {
	for _, element := range elements {
		for _, sliceElement := range slice {
			if element == sliceElement {
				return true
			}
		}
	}
	return false
}

// Behold... my unholy attack on god himself.
func Ternary[T any](condition bool, a, b T) T {
	if condition {
		return a
	}
	return b
}

// Demons possessed me, then using my mortal form, they wrote this abomination.
func Try(funcs ...func() error) error {
	for _, fn := range funcs {
		if err := fn(); err != nil {
			return err
		}
	}
	return nil
}
