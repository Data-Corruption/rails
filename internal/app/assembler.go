package app

import (
	"fmt"
	"strconv"
	"strings"

	"rails/internal/utils"
)

// ---- types -----------------------------------------------------------------

type encodingType int

const (
	AB encodingType = iota
	CA
	CAB
	C_IMM
	IMM_C
)

type instruction struct {
	opcode   uint8
	encoding encodingType
}

// ---- variables -------------------------------------------------------------

// instructions maps assembly instructions to their opcode and encoding type.
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

// ---- Public functions ------------------------------------------------------

// Assemble converts a string of assembly code to a program ROM.
// Returns the number of instructions assembled and an error if one occurred.
func Assemble(assembly string, programROM []uint16) (uint, error) {
	var tagMap = map[string]uint16{}
	var lineNumber = (uint16)(0) // actual line number (excludes empty lines and comments)

	// first pass: build line tag map
	if err := utils.ForEachLine(assembly, func(index int, line string) error {
		line = strings.TrimSpace(line)
		// skip lines that are empty or start with a comment
		if len(line) == 0 || strings.HasPrefix(line, "//") || strings.HasPrefix(line, "#") {
			return nil
		}
		// get first token
		tokens := strings.Fields(line)
		if len(tokens) == 0 || len(tokens[0]) == 0 {
			return nil
		}
		// if ends with a colon, it's a tag, add it to the map
		if tokens[0][len(tokens[0])-1] == ':' {
			tagMap[tokens[0]] = lineNumber
		}
		lineNumber++
		return nil
	}); err != nil {
		return 0, err
	}

	// second pass: assemble
	lineNumber = 0 // reset line number for second pass
	if err := utils.ForEachLine(assembly, func(index int, line string) error {
		// check if program is too long, trim whitespace, skip empty lines and comments
		if lineNumber > 255 {
			return fmt.Errorf("program is too long. Max length is 256 instructions")
		}
		line = strings.TrimSpace(line)
		if len(line) == 0 || strings.HasPrefix(line, "//") || strings.HasPrefix(line, "#") {
			return nil
		}

		// parse instruction
		encodedInstruction, err := parseInstruction(line, tagMap)
		if err != nil {
			return fmt.Errorf("line %d: %s", index+1, err.Error())
		}

		// write encoded instruction to program ROM and increment line number
		programROM[lineNumber] = encodedInstruction
		lineNumber++
		return nil
	}); err != nil {
		return 0, err
	}

	// clear remaining program ROM
	for i := lineNumber; i < 256; i++ {
		programROM[i] = 0
	}

	return uint(lineNumber), nil
}

// ---- Private functions -----------------------------------------------------

// convertPseudoInstruction converts pseudo instructions to real instructions.
func convertPseudoInstruction(tokens []string) {
	switch tokens[0] {
	case "NOP":
		copy(tokens, []string{"ADD", "r0", "r0", "r0"})
	case "MOV":
		copy(tokens, []string{"ADD", tokens[1], "r0", tokens[2]})
	case "JMP":
		copy(tokens, []string{"BEQ", tokens[1], "r15"})
	case "EXIT":
		copy(tokens, []string{"JMPL", "r0", "r0"})
	default:
		return
	}
}

// parseInstruction returns the binary encoded equivalent of an assembly instruction.
func parseInstruction(line string, tagMap map[string]uint16) (uint16, error) {
	// split line into tokens
	tokens := strings.Fields(line)
	if len(tokens) < 1 {
		return 0, fmt.Errorf("error parsing instruction")
	}

	// if first token is a tag, remove it
	if tokens[0][len(tokens[0])-1] == ':' {
		tokens = tokens[1:]
	}

	// if it's a pseudo instruction, convert it to a real instruction
	convertPseudoInstruction(tokens)

	// get instruction
	instruction, ok := instructions[tokens[0]]
	if !ok {
		return 0, fmt.Errorf("unknown instruction: %s", tokens[0])
	}

	// check for invalid number of arguments
	minArgCount := 3
	if instruction.encoding == CAB {
		minArgCount = 4
	}
	if len(tokens) < minArgCount {
		return 0, fmt.Errorf("not enough args, expected %d, got: %d", minArgCount-1, len(tokens)-1)
	}

	// parses a register, removes the 'r' from the front of the string if present and returns the number
	parseReg := func(out *uint8, input string) func() error {
		return func() error {

			// make sure string isn't empty
			if len(input) == 0 {
				return fmt.Errorf("error parsing register")
			}

			// remove 'r' from the front of the string if present
			if input[0] == 'r' {
				input = input[1:]
			}

			// parse number
			reg, err := strconv.ParseUint(input, 10, 8)
			if err != nil {
				return fmt.Errorf("failed to parse register: %s", input)
			}

			// make sure it's in range
			if reg > 15 {
				return fmt.Errorf("register out of range: %s", input)
			}

			*out = uint8(reg)
			return nil
		}
	}

	// parses an immediate value, if the immediate is a tag, it will return the line number of the tag
	parseImm := func(out *uint8, input string) func() error {
		return func() error {

			// make sure string isn't empty
			if len(input) == 0 {
				return fmt.Errorf("error parsing immediate")
			}

			// if the last character is a colon, it's a tag
			if input[len(input)-1] == ':' {
				// get line number using tag
				if line_number, ok := tagMap[input]; ok {
					*out = uint8(line_number)
					return nil
				}
				return fmt.Errorf("unknown line tag: %s", input)
			} else {
				// not a tag, parse number
				if imm, err := strconv.ParseUint(input, 10, 8); err == nil {
					*out = uint8(imm)
					return nil
				}
				return fmt.Errorf("failed to parse immediate: %s", input)
			}
		}
	}

	// parse instruction
	var a, b, c, imm uint8
	parse := func() error {
		switch instruction.encoding {
		case AB:
			return utils.Try(parseReg(&a, tokens[1]), parseReg(&b, tokens[2]))
		case CA:
			return utils.Try(parseReg(&c, tokens[1]), parseReg(&a, tokens[2]))
		case CAB:
			return utils.Try(parseReg(&c, tokens[1]), parseReg(&a, tokens[2]), parseReg(&b, tokens[3]))
		case C_IMM:
			return utils.Try(parseReg(&c, tokens[1]), parseImm(&imm, tokens[2]))
		case IMM_C:
			return utils.Try(parseImm(&imm, tokens[1]), parseReg(&c, tokens[2]))
		default:
			return fmt.Errorf("unknown encoding type: %d", instruction.encoding)
		}
	}
	if err := parse(); err != nil {
		return 0, err
	}

	// encode instruction
	if instruction.encoding == C_IMM || instruction.encoding == IMM_C {
		return ((uint16(instruction.opcode) << 12) | (uint16(imm) << 4) | uint16(c)), nil // [opcode 4bit | imm 8bit | c 4bit]
	} else {
		return ((uint16(instruction.opcode) << 12) | (uint16(a) << 8) | (uint16(b) << 4) | uint16(c)), nil // [opcode 4bit | a 4bit | b 4bit | c 4bit]
	}
}
