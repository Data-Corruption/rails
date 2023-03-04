#pragma once

#include <unordered_map>
#include <stdexcept>

#include "typedefs.hpp"
#include "string_addons.hpp"

struct instruction {
	std::string name;
	uint_8 opd_code;
	std::string encoding_type;
};

const std::vector<instruction> instructions{
	{"ADD",  0,  "CAB"},
	{"ADDC", 1,  "CAB"},
	{"SUB",  2,  "CAB"},
	{"SWB",  3,  "CAB"},
	{"NAND", 4,  "CAB"},
	{"RSFT", 5,  "CA"},
	{"IMM",  6,  "C immediate"},
	{"LD",   7,  "CA"},
	{"LDIM", 8,  "C immediate"},
	{"ST",   9,  "AB"},
	{"STIM", 10, "immediate C"},
	{"BEQ",  11, "immediate C"},
	{"BGT",  12, "immediate C"},
	{"JMPL", 13, "CA"},
	{"IN",   14, "CA"},
	{"OUT",  15, "AB"},
};

class Assembler {

private:

	uint line_number;
	std::vector<std::vector<std::string>> lines;
	std::unordered_map<std::string, uint_8> tag_line_number_map;

	// parses a register, removes the 'r' from the front of the string
	uint_8 parse_reg(std::string& input) {
		input.erase(input.begin());
		uint_8 result = stoi(input);
		if (result > 15) {
			throw std::invalid_argument("ERROR Line number: " + std::to_string(line_number) + " register index out of range!");
		}
		return result;
	}

	// parses an immediate value, if the immediate is a tag, it will return the line number of the tag
	uint_8 parse_imm(std::string& input) {
		if (string_contains(input, ":"))
			return tag_line_number_map[input];
		uint_8 result = stoi(input);
		if (result > 255) {
			throw std::invalid_argument("ERROR Line number: " + std::to_string(line_number) + " immediate out of range!");
		}
		return result;
	}

public:

	// returns the assembled program
	void run(std::vector<std::vector<uint_8>>& program, const std::string& raw_file) {
		tag_line_number_map.clear();
		lines.clear();
		program.clear();

		std::vector<std::string> unparsed_lines = string_split(raw_file, { '\n' });

		// load tags, split lines
		line_number = 0;
		for (uint i = 0; i < unparsed_lines.size(); i++) {
			// split by tabs/spaces add to lines
			lines.push_back(string_split(unparsed_lines[i], { ' ', '	' }));
			// if there is a line tag, save the line number, remove tag
			if (string_contains(lines.back()[0], ":")) {
				tag_line_number_map[lines.back()[0]] = line_number;
				lines.back().erase(lines.back().begin());
			}
			line_number++;
		}

		// parse lines
		for (line_number = 0; line_number < lines.size(); line_number++) {
			std::vector<std::string> line = lines[line_number];

			// if line is empty or a comment skip it
			if ((line.size() == 0) || (line[0][0] == '#'))
				continue;

			program.emplace_back();

			// parse pseudo instructions
			if (line[0] == "NOP") {
				program.back() = { 0, 0, 0, 0 };
			} else if (line[0] == "MOV") {
				program.back() = { 0, parse_reg(line[1]), 0, parse_reg(line[2]) };
			} else if (line[0] == "JMP") {
				program.back() = { 11, parse_imm(line[1]), 15 };
			} else if (line[0] == "EXIT") {
				program.back() = { 13, 0, 0, 0 };
			}

			// parse instructions
			for (auto& instruction : instructions) {
				if (line[0] == instruction.name) {
					if (instruction.encoding_type == "CAB") {
						program.back() = { instruction.opd_code, parse_reg(line[2]), parse_reg(line[3]), parse_reg(line[1]) };
					} else if (instruction.encoding_type == "CA") {
						program.back() = { instruction.opd_code, parse_reg(line[2]), parse_reg(line[1]) };
					} else if (instruction.encoding_type == "AB") {
						program.back() = { instruction.opd_code, parse_reg(line[1]), parse_reg(line[2]) };
					} else if (instruction.encoding_type == "C immediate") {
						program.back() = { instruction.opd_code, parse_imm(line[2]), parse_reg(line[1]) };
					} else if (instruction.encoding_type == "immediate C") {
						program.back() = { instruction.opd_code, parse_imm(line[1]), parse_reg(line[2]) };
					}
				}
			}

			// if the program line is empty, it means that the instruction was not found and parsing failed
			if (program.back().size() == 0)
				throw std::invalid_argument("ERROR Line number: " + std::to_string(line_number) + " something went wrong here.");
				
		}

	}

};