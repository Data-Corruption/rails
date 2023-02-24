#pragma once

#include <vector>

#include "typedefs.hpp"
#include "art.hpp"
#include "system.hpp"

class ArithmeticLogicUnit {
public:
    bool carry_flag = false;

    uint_8 add(uint_8 a, uint_8 b) {
        uint16_t result = a + b;
        carry_flag = result > 255;
        return (uint_8)result;
    }
	uint_8 addc(uint_8 a, uint_8 b) {
        uint16_t result = a + b + carry_flag;
        carry_flag = result > 255;
        return (uint_8)result;
    }
	uint_8 sub(uint_8 a, uint_8 b) {
        uint16_t result = a - b;
        carry_flag = result > 255;
        return (uint_8)result;
    }
	uint_8 swb(uint_8 a, uint_8 b) {
        uint16_t result = a - b - carry_flag;
        carry_flag = result > 255;
        return (uint_8)result;
    }
    uint_8 nand(uint_8 a, uint_8 b) {
        return ~(a & b);
    }
};

class Emulator {
public:

    ArithmeticLogicUnit alu;
    uint_8 program_counter = 0;
    std::vector<uint_8> registers;
	std::vector<uint_8> io_registers;
    std::vector<uint_8> ram;

    void print_registers() {
        my_system::console::log("Registers \n");
        for (uint i = 0; i < registers.size(); i++) {
            my_system::console::log("r" + std::to_string(i) + ":" + std::to_string(registers[i]) + " ");
            if (i % 4 == 3)
                my_system::console::log("\n");
        }
        my_system::console::log("\n");
    }

    void print_io_registers() {
        my_system::console::log("IO Registers \n");
        for (uint i = 0; i < io_registers.size(); i++) {
            my_system::console::log("r" + std::to_string(i) + ":" + std::to_string(io_registers[i]) + " ");
            if (i % 4 == 3)
                my_system::console::log("\n");
        }
        my_system::console::log("\n");
    }

    void run(const std::vector<std::vector<uint_8>>& program) {
        registers.resize(16);
        io_registers.resize(16);
        ram.resize(256);

        bool running = true;
        bool modified_pc = false;
        program_counter = 0;

        while (running) {
            std::vector<uint_8> instruction = program[program_counter];
            switch (instruction[0]) {
            case 0: // ADD
                registers[instruction[3]] = alu.add(registers[instruction[1]], registers[instruction[2]]);
                break;
            case 1: // ADDC
                registers[instruction[3]] = alu.addc(registers[instruction[1]], registers[instruction[2]]);
                break;
            case 2: // SUB
                registers[instruction[3]] = alu.sub(registers[instruction[1]], registers[instruction[2]]);
                break;
            case 3: // SWB
                registers[instruction[3]] = alu.swb(registers[instruction[1]], registers[instruction[2]]);
                break;
            case 4: // NAND
                registers[instruction[3]] = alu.nand(registers[instruction[1]], registers[instruction[2]]);
                break;
            case 5: // RSFT
                registers[instruction[3]] = registers[instruction[1]] >> 1;
                break;
            case 6: // IMM
                registers[instruction[2]] = instruction[1];
                break;
            case 7: // LD 
                registers[instruction[3]] = ram[registers[instruction[1]]];
                break;
            case 8: // LDIM
                registers[instruction[2]] = ram[instruction[1]];
                break;
            case 9: // ST
                ram[registers[instruction[1]]] = registers[instruction[2]];
                break;
            case 10: // STIM
                ram[instruction[1]] = registers[instruction[2]];
                break;
            case 11: // BEQ
                if (registers[15] == registers[instruction[2]]) {
                    program_counter = instruction[1];
                    modified_pc = true;
                }
                break;
            case 12: // BGT
                if (registers[15] > registers[instruction[2]]) {
                    program_counter = instruction[1];
                    modified_pc = true;
                }
                break;
            case 13: // JMPL
                if ((instruction[1] == 0) && (instruction[2] == 0) && (instruction[3] == 0)){
                    running = false;
                    break;
                }
                registers[instruction[3]] = program_counter + 1;
                program_counter = instruction[1];
                modified_pc = true;
                break;
            case 14: // IN 
                my_system::console::clear();
                my_system::console::log(art);
                my_system::console::log("Program Counter: " + std::to_string(program_counter) + "\n");
                print_registers();
                print_io_registers();
                registers[instruction[2]] = std::stoi(my_system::console::get_input("Program reading io register: " + std::to_string(instruction[1]) + ", enter value 0-255: "));
                break;
            case 15: // OUT
                io_registers[instruction[1]] = registers[instruction[2]];
                break;
            default:
                std::cout << "Invalid instruction: " << instruction[0] << std::endl;
                return;
            }
            if (!modified_pc)
                program_counter++;
            modified_pc = false;
        }

        my_system::console::clear();
        my_system::console::log(art);
        print_registers();
        print_io_registers();

        my_system::console::log("Program finished!");

    }

};