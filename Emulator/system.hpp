#pragma once

#include <string>
#include <vector>
#include <sstream>
#include <fstream>
#include <filesystem>
#include <iostream>
#include <stdexcept>

#ifdef _WIN32
#include <windows.h>
#endif

namespace my_system {
    namespace files {
            // returns true if the file or directory exists
			extern bool exists(const std::string& path) {
                return std::filesystem::exists(path);
            }
            // returns true if the path is a directory
            extern bool is_directory(const std::string& path) {
                return std::filesystem::is_directory(path);
            }
            // reads the file at the given path and stores the data in the given string
			extern void read(std::string& data, const std::string& path) {
                if (!exists(path))
                    throw std::invalid_argument("Attempting to read file that does not exist: " + path);
                try {
                    std::stringstream stream;
                    std::ifstream file(path, std::ios::binary);
                    stream << file.rdbuf();
                    file.close();
                    data = stream.str();
                }
                catch (std::ifstream::failure e) {
                    throw std::runtime_error("Failed to read file: " + path);
                }
            }
            // writes the given data to the file at the given path
			extern void write(const std::string& data, const std::string& path) {
                std::ofstream file(path, std::ios::binary);
                file << data;
            }
	};
    namespace console {
        // clears the console
        extern void clear() {
            #ifdef __linux__ 
            std::cout << "\x1B[2J\x1B[H";
            #elif _WIN32
            COORD topLeft = { 0, 0 };
            HANDLE console_handle = GetStdHandle(STD_OUTPUT_HANDLE);
            CONSOLE_SCREEN_BUFFER_INFO screen;
            DWORD written;
            GetConsoleScreenBufferInfo(console_handle, &screen);
            FillConsoleOutputCharacterA(console_handle, ' ', screen.dwSize.X * screen.dwSize.Y, topLeft, &written);
            FillConsoleOutputAttribute(console_handle, FOREGROUND_GREEN | FOREGROUND_RED | FOREGROUND_BLUE, screen.dwSize.X * screen.dwSize.Y, topLeft, &written);
            SetConsoleCursorPosition(console_handle, topLeft);
            #endif
        }
        // logs the given message to the console
        extern void log(const std::string& message) {
            std::cout << message;
        }
        // waits for the user to enter a string and returns it
        extern std::string get_input(const std::string& message) {
            std::string input;
            std::cout << message;
            std::getline(std::cin, input);
            return input;
        }
    };
}