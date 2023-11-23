package main

import (
	r "reflect"
	"strconv"
	"math"
	"bufio"
  "os"
	"io/ioutil"
  "strings"
)

type LineCallback func(line string) error

func ForEachLine(s string, callback LineCallback) error {
	scanner := bufio.NewScanner(strings.NewReader(s))
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		if err := callback(scanner.Text()); err != nil { return err }
	}
	if err := scanner.Err(); err != nil { return err }
	return nil
}

func WriteStringToFile(path, data string) error {
	file, err := os.Create(path)
	if err != nil { return err }
	defer file.Close()
	_, err = file.WriteString(data)
	if err != nil { return err }
	return nil
}
func WriteBytesToFile(path string, data []byte) error {
	file, err := os.Create(path)
	if err != nil { return err }
	defer file.Close()
	_, err = file.Write(data)
	if err != nil { return err }
	return nil
}

func readFile(path string) ([]byte, error) {
	file, err := os.Open(path)
	if err != nil { return nil, err }
	defer file.Close()
	b, err := ioutil.ReadAll(file)
	if err != nil { return nil, err }
	return b, nil
}
func ReadBytesFromFile(path string) ([]byte, error) {
	return readFile(path)
}
func ReadStringFromFile(path string) (string, error) {
	bytes, err := readFile(path)
	if err != nil { return "", err }
	return string(bytes), nil
}

func NumberToBool(number interface{}) bool {
	v := r.ValueOf(number)
	switch v.Kind() {
		case r.Int, r.Int8, r.Int16, r.Int32, r.Int64: return v.Int() != 0
		case r.Uint, r.Uint8, r.Uint16, r.Uint32, r.Uint64: return v.Uint() != 0
		default: return false
	}
}

func BoolToNumber(b bool) uint16 {
	if b { return 1 }
	return 0
}

func NumberToString(number interface{}, width uint16, paddingChar string, base int) string {
	var s string
	v := r.ValueOf(number)
	switch v.Kind() {
		case r.Int, r.Int8, r.Int16, r.Int32, r.Int64: s = strconv.FormatInt(v.Int(), base)
		case r.Uint, r.Uint8, r.Uint16, r.Uint32, r.Uint64: s = strconv.FormatUint(v.Uint(), base)
		default: return "invalid type"
	}
	for len(s) < int(width) {
		s = paddingChar + s
	}
	return s
}

// CenterGravityAdjust adjusts a value based on its distance from the center of a range
func CenterGravityAdjust(value, min, max, aggressiveness float64) float64 {
	if min >= max { panic("CenterGravityAdjust: min must be less than max") }
	center := (min + max) / 2
	distance := value - center
	return value - distance*math.Pow(math.Abs(distance)/(max-min), aggressiveness)
}