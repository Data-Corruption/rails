package main

import (
	"fmt"
	"strings"
	"strconv"
)

// ---- types ----

type encodingType int
const (
	AB encodingType = iota
	CA
	CAB
	C_IMM
	IMM_C
)

type instruction struct {
	opcode uint8;
	encoding encodingType;
}

// ---- variables ----

var instructions = map[string]instruction{
	"ADD":  {0, CAB},
	"ADDC": {1, CAB},
	"SUB":  {2, CAB},
	"SWB":  {3, CAB},
	"NAND": {4, CAB},
	"RSFT": {5, CA},
	"IMM":  {6, C_IMM},
	"LD":   {7, CA},
	"LDIM": {8, C_IMM},
	"ST":   {9, AB},
	"STIM": {10, IMM_C},
	"BEQ":  {11, IMM_C},
	"BGT":  {12, IMM_C},
	"JMPL": {13, CA},
	"IN":   {14, CA},
	"OUT":  {15, AB},
}

var line_tag_map = map[string]uint16{}

// ---- functions ----

func isLineCommentOrEmpty(line string) bool {
	if len(line) < 2 { return true }
	if line[0] == '#' || (line[0] == '/' && line[1] == '/') { return true }
	return false
}

// parses a register, removes the 'r' from the front of the string if present and returns the number
func parseReg(input string, src_line_number uint16) (uint8, error) {
	// make sure string isn't empty
	if len(input) == 0 { 
		return 0, fmt.Errorf("Line %d: Error parsing register", src_line_number)
	}

	// remove 'r' from the front of the string if present
	if input[0] == 'r' { input = input[1:] }

	// parse register
	reg, err := strconv.ParseUint(input, 10, 8)
	if err != nil {
		return 0, fmt.Errorf("Line %d: Failed to parse register: %s", src_line_number, input)
	}
	// make sure it's in range
	if reg > 15 {
		return 0, fmt.Errorf("Line %d: Register out of range: %s", src_line_number, input)
	}
	return uint8(reg), nil
}

// parses an immediate value, if the immediate is a tag, it will return the line number of the tag
func parseImm(input string, src_line_number uint16) (uint8, error) {
	// make sure string isn't empty
	if len(input) == 0 {
		return 0, fmt.Errorf("Line %d: Error parsing immediate", src_line_number)
	}

	// if the last character is a colon, it's a tag
	if input[len(input)-1] == ':' {
		// get line number from tag
		line_number, ok := line_tag_map[input]
		if !ok { return 0, fmt.Errorf("Line %d: Unknown line tag: %s", src_line_number, input) }
		return uint8(line_number), nil
	} else {
		// parse immediate
		imm, err := strconv.ParseUint(input, 10, 8)
		if err != nil { return 0, fmt.Errorf("Line %d: Failed to parse immediate: %s", src_line_number, input) }
		return uint8(imm), nil
	}
}

// converts pseudo instructions to real instructions if present, otherwise returns the original tokens
func convertPseudoIfPresent(tokens []string) []string {
	switch tokens[0] {
		case "NOP": return []string{"ADD", "r0", "r0", "r0"}
		case "MOV": return []string{"ADD", tokens[1], "r0", tokens[2]}
		case "JMP": return []string{"BEQ", tokens[1], "r15"}
		case "EXIT": return []string{"JMPL", "r0", "r0"}
		default: return tokens
	}
}

// encodes an instruction, result = [opcode 4bit | a 4bit | b 4bit | c 4bit]
func encode(opcode uint16, a uint16, b uint16, c uint16) uint16 {
	return (opcode << 12) | (a << 8) | (b << 4) | c
}
// encodes an imm instruction, result = [opcode 4bit | imm 8bit | c 4bit]
func encodeImm(opcode uint16, imm uint16, c uint16) uint16 {
	return (opcode << 12) | (imm << 4) | c
}

// parses an instruction, returns the binary representation of the instruction. sln is short for source line number
func parseInstruction(line string, sln uint16) (uint16, error) {
	// split line into tokens
	raw_tokens := strings.Fields(line)
	if len(raw_tokens) < 1 { return 0, fmt.Errorf("Line %d: Error parsing instruction", sln) }

	// trim whitespace from tokens
	for i := 0; i < len(raw_tokens); i++ {
		raw_tokens[i] = strings.TrimSpace(raw_tokens[i])
	}

	// if first token is a tag, remove it
	if raw_tokens[0][len(raw_tokens[0])-1] == ':' {
		raw_tokens = raw_tokens[1:]
	}

	// convert pseudo instructions to real instructions if present
	tokens := convertPseudoIfPresent(raw_tokens)

	// get instruction
	inst, ok := instructions[tokens[0]]
	if !ok { return 0, fmt.Errorf("Line %d: Unknown instruction: %s", sln, tokens[0]) }

	// check for invalid number of arguments
	minArgCount := 3
	if inst.encoding == CAB { minArgCount = 4 }
	if len(tokens) < minArgCount {
		return 0, fmt.Errorf("Line %d: Not enough args, expected %d, got: %d", sln, minArgCount-1, len(tokens)-1)
	}

	// parse instruction
	var a, b, c, imm uint8
	var err error
	parse := func() error {
    switch inst.encoding {
    case AB:
      if a, err = parseReg(tokens[1], sln); err != nil { return err }
      if b, err = parseReg(tokens[2], sln); err != nil { return err }
			break
		case CA:
			if c, err = parseReg(tokens[1], sln); err != nil { return err }
			if a, err = parseReg(tokens[2], sln); err != nil { return err }
			break
		case CAB:
			if c, err = parseReg(tokens[1], sln); err != nil { return err }
			if a, err = parseReg(tokens[2], sln); err != nil { return err }
			if b, err = parseReg(tokens[3], sln); err != nil { return err }
			break
		case C_IMM:
			if c, err = parseReg(tokens[1], sln); err != nil { return err }
			if imm, err = parseImm(tokens[2], sln); err != nil { return err }
			break
		case IMM_C:
			if imm, err = parseImm(tokens[1], sln); err != nil { return err }
			if c, err = parseReg(tokens[2], sln); err != nil { return err }
			break
		default:
			return fmt.Errorf("Line %d: Unknown encoding type: %d", sln, inst.encoding)
		}
		return nil
	}
	if err := parse(); err != nil {
		return 0, fmt.Errorf("Line %d:\nRaw: \"%s\"\n%d Tokens: %s\nError: %s", sln, line, len(tokens), tokens, err.Error())
	}

	// encode instruction
	if inst.encoding == C_IMM || inst.encoding == IMM_C {
		return encodeImm(uint16(inst.opcode), uint16(imm), uint16(c)), nil
	} else {
		return encode(uint16(inst.opcode), uint16(a), uint16(b), uint16(c)), nil
	}
}

// Returns string if error, nil if successful
func AssembleFile(path string, prom []uint16) (uint8, error) {
	// clear line tag map
	line_tag_map = map[string]uint16{}
	var line_number uint16 = 0

	// read file
	raw_file, err := ReadStringFromFile(path)
	if err != nil { return 0, err }
	if len(raw_file) == 0 { return 0, fmt.Errorf("File is empty") }

	// first pass: build line tag map
	ForEachLine(raw_file, func(line string) error {
		line = strings.TrimSpace(line)
		if isLineCommentOrEmpty(line) { return nil }
		// get up to the first space
		first_token := strings.Split(line, " ")[0]
		// if ends with a colon, it's a tag, add it to the map
		if first_token[len(first_token)-1] == ':' {
			line_tag_map[first_token] = line_number
		}
		line_number++
		return nil
	})

	// second pass: assemble
	line_number = 0
	var src_line_number uint16 = 0
	err = ForEachLine(raw_file, func(line string) error {
		// check if program is too long
		if line_number > 255 { return fmt.Errorf("Program is too long. Max length is 256 instructions") }

		// trim whitespace, skip empty lines and comments
		line = strings.TrimSpace(line)
		if isLineCommentOrEmpty(line) {
			src_line_number++
			return nil
		}

		// parse instruction
		encoded_instr, err := parseInstruction(line, src_line_number)
		if err != nil { return err }

		// write instruction to prom and increment line numbers
		prom[line_number] = encoded_instr
		src_line_number++
		line_number++
		return nil
	})

	// clear remaining prom
	for i := line_number; i < 256; i++ {
		prom[i] = 0
	}

	return uint8(line_number), err
}