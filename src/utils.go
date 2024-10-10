package main

import (
	"bufio"
	r "reflect"
	"strconv"
	"strings"
)

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

// Behold... my unholy attack on god himself.
func Ternary[T any](condition bool, a, b T) T {
	if condition {
		return a
	}
	return b
}

// Demons possessed me then wrote this against my will. Forgive me.
func Try(funcs ...func() error) error {
	for _, fn := range funcs {
		if err := fn(); err != nil {
			return err
		}
	}
	return nil
}

// NumberToString converts a number to a string with a given width and padding character. e.g. (5, 3, "0", 10) -> "005", (5, 3, "0", 2) -> "101"
func NumberToString(number interface{}, width uint16, paddingChar string, base int) string {
	var s string
	v := r.ValueOf(number)
	switch v.Kind() {
	case r.Int, r.Int8, r.Int16, r.Int32, r.Int64:
		s = strconv.FormatInt(v.Int(), base)
	case r.Uint, r.Uint8, r.Uint16, r.Uint32, r.Uint64:
		s = strconv.FormatUint(v.Uint(), base)
	default:
		return "invalid type"
	}
	for len(s) < int(width) {
		s = paddingChar + s
	}
	return s
}
