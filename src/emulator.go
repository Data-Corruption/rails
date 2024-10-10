package main

import (
	"bytes"
	"encoding/binary"
	"sync"
)

const (
	OPCODE_MASK uint16 = 0xF000
	A_MASK      uint16 = 0x0F00
	B_MASK      uint16 = 0x00F0
	C_MASK      uint16 = 0x000F
	IMM_MASK    uint16 = 0x0FF0
	BYTE_BASK   uint16 = 0x00FF
	// 1101 0000 0000 0000 (exit instruction)
	EXIT_INSTRUCTION uint16 = 0xD000
)

type RailsState struct {
	Prom          [256]uint16 // program rom
	Ram           [256]uint8  // ram
	Regfile       [16]uint8   // general purpose registers
	InRegs        [16]uint8   // input registers
	OutRegs       [16]uint8   // output registers
	Pc            uint8       // program counter
	ProgramLength uint8       // useful for debugging and avoiding unnecessary work
	CarryFlag     bool        // carry flag
}

// all functions assume the mutex is locked by the caller with the exception of EvalUntil
type RailsEmulator struct {
	State      RailsState // cpu state
	Mutex      sync.Mutex // mutex for thread safety
	IsBusy     bool       // is the cpu currently evaluating via a goroutine?
	ShouldStop bool       // is something asking the cpu to stop evaluating?
}

// save state to file
func (e *RailsEmulator) SaveState(path string) error {
	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.LittleEndian, e.State) // encode
	if err != nil {
		return err
	}
	return WriteBytesToFile(path, buf.Bytes())
}

// load state from file
func (e *RailsEmulator) LoadState(path string) error {
	data, err := ReadBytesFromFile(path)
	if err != nil {
		return err
	}
	buf := bytes.NewReader(data)
	return binary.Read(buf, binary.LittleEndian, &e.State) // decode
}

// clears ram, registers, carry flag, and the program counter
func (e *RailsEmulator) Reset() {
	if e.IsBusy {
		panic("Attempted to reset CPU while it was busy")
	}
	e.State.Ram = [256]uint8{}
	e.State.Regfile = [16]uint8{}
	e.State.InRegs = [16]uint8{}
	e.State.OutRegs = [16]uint8{}
	e.State.CarryFlag = false
	e.State.Pc = 0
}

// returns the target instruction in the format "xxxx-xxxx-xxxx-xxxx"
func (e *RailsEmulator) InstructionToString(address uint8) string {
	s := NumberToString(e.State.Prom[address], 16, "0", 2)
	for i := 4; i < len(s); i += 5 {
		s = s[:i] + "-" + s[i:]
	}
	return s
}

// assumes the mutex is locked by the caller
// handles pc incrementing if needed
func (e *RailsEmulator) executeInstruction() {
	opcode := (e.State.Prom[e.State.Pc] & opcode_mask) >> 12
	a := uint8((e.State.Prom[e.State.Pc] & a_mask) >> 8)
	b := uint8((e.State.Prom[e.State.Pc] & b_mask) >> 4)
	c := uint8(e.State.Prom[e.State.Pc] & c_mask)
	imm := uint8((e.State.Prom[e.State.Pc] & imm_mask) >> 4)

	switch opcode {
	case 0: // ADD
		result := uint16(e.State.Regfile[a]) + uint16(e.State.Regfile[b])
		if result > 255 {
			e.State.CarryFlag = true
		} else {
			e.State.CarryFlag = false
		}
		e.State.Regfile[c] = uint8(result & byteMask)
		break
	case 1: // ADDC
		result := uint16(e.State.Regfile[a]) + uint16(e.State.Regfile[b]) + Ternary[uint16](e.State.CarryFlag, 1, 0)
		if result > 255 {
			e.State.CarryFlag = true
		} else {
			e.State.CarryFlag = false
		}
		e.State.Regfile[c] = uint8(result & byteMask)
		break
	case 2: // SUB
		result := uint16(e.State.Regfile[a]) - uint16(e.State.Regfile[b])
		if result > 255 {
			e.State.CarryFlag = true
		} else {
			e.State.CarryFlag = false
		}
		e.State.Regfile[c] = uint8(result & byteMask)
		break
	case 3: // SWB
		result := uint16(e.State.Regfile[b]) - uint16(e.State.Regfile[a]) - Ternary[uint16](e.State.CarryFlag, 1, 0)
		if result > 255 {
			e.State.CarryFlag = true
		} else {
			e.State.CarryFlag = false
		}
		e.State.Regfile[c] = uint8(result & byteMask)
		break
	case 4:
		e.State.Regfile[c] = ^(e.State.Regfile[a] & e.State.Regfile[b])
		break // NAND
	case 5:
		e.State.Regfile[c] = e.State.Regfile[a] >> 1
		break // RSFT
	case 6:
		e.State.Regfile[c] = imm
		break // IMM
	case 7:
		e.State.Regfile[c] = e.State.Ram[e.State.Regfile[a]]
		break // LD
	case 8:
		e.State.Regfile[c] = e.State.Ram[imm]
		break // LDIM
	case 9:
		e.State.Ram[e.State.Regfile[a]] = e.State.Regfile[b]
		break // ST
	case 10:
		e.State.Ram[imm] = e.State.Regfile[c]
		break // STIM
	case 11:
		if e.State.Regfile[15] == e.State.Regfile[c] {
			e.State.Pc = imm
			e.State.CarryFlag = false
			return
		} // BEQ
	case 12:
		if e.State.Regfile[15] > e.State.Regfile[c] {
			e.State.Pc = imm
			e.State.CarryFlag = false
			return
		} // BGT
	case 13: // JMPL
		e.State.Regfile[c] = e.State.Pc + 1
		e.State.Regfile[0] = 0
		e.State.Pc = e.State.Regfile[a]
		e.State.CarryFlag = false
		return
	case 14:
		e.State.Regfile[c] = e.State.InRegs[a]
		break // IN
	case 15:
		e.State.OutRegs[a] = e.State.Regfile[b]
		break // OUT
	default:
		panic("Invalid opcode")
	}
	e.State.Regfile[0] = 0 // register 0 is always 0
	e.State.Pc++
}

// evaluate an instruction
func (e *RailsEmulator) Eval() {
	if e.IsBusy {
		panic("CPU is already busy")
	}
	e.executeInstruction()
}

type StopType int

const (
	IO StopType = iota
	EXIT
)

// evaluate instructions until either: an IO or EXIT instruction is encountered.
// only function that doesn't assume the mutex is locked by the caller
func (e *RailsEmulator) EvalUntil(stopType StopType) {
	e.Mutex.Lock()
	if e.IsBusy {
		panic("CPU is already busy")
	}
	e.IsBusy = true
	e.Mutex.Unlock()

	exitLoop := func() {
		e.ShouldStop = false
		e.IsBusy = false
		e.Mutex.Unlock()
	}

	for {
		e.Mutex.Lock()
		if e.ShouldStop {
			exitLoop()
			break
		}

		// if instruction is the target stop type then stop
		opcode := (e.State.Prom[e.State.Pc] & opcode_mask) >> 12
		if stopType == IO {
			if opcode == 14 || opcode == 15 {
				exitLoop()
				break
			}
		} else if stopType == EXIT {
			if e.State.Prom[e.State.Pc] == exitInstr {
				exitLoop()
				break
			}
		}

		// execute instruction
		e.executeInstruction()
		e.Mutex.Unlock()
	}
}
