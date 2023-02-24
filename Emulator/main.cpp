#include <iostream>
#include <stdexcept>
#include <string>
#include <vector>

#include "typedefs.hpp"
#include "system.hpp"
#include "assembler.hpp"
#include "emulator.hpp"

Assembler ass;
Emulator emu;

int main(int argc, char** argv) {

	std::string unassembled_program;
	std::vector<std::vector<uint_8>> program;

	try {

		// print art
		std::cout << art << std::endl;

		// get path to assembly file
		if (argc < 2)
			throw std::invalid_argument("missing command line argument(s) expected: 'rails <path to assembly file>'");
		std::string path = argv[1];

		// check if path is valid
		if (!my_system::files::exists(path) || my_system::files::is_directory(path))
			throw std::invalid_argument("argument is not a valid file path: \"" + path + "\"");

		// assemble program
		my_system::files::read(unassembled_program, path);
		ass.run(program, unassembled_program);
		
		// run program
		emu.run(program);

	} catch (const std::exception& e) {
		std::cerr << "ERROR: " << e.what() << std::endl;
		exit(EXIT_FAILURE);
	}

}