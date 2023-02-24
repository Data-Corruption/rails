#include <string>
#include <vector>
#include <algorithm>

#include "typedefs.hpp"

// splits a string into segments when a delimiter is encountered
std::vector<std::string> string_split(const std::string& input, const std::vector<char> delimiters) {
    std::vector<std::string> result;
    std::string segment = "";
    for (const char& c : input) {
        // if c is a delimiter and segment is not empty add segment to output then clear segment. Else add c to segment
        if (std::find(delimiters.begin(), delimiters.end(), c) != delimiters.end()) {
            if (segment.size() != 0) {
                result.push_back(segment);
                segment = "";
            }
        } else {
            segment += c;
		}
	}
    // if segment is not empty add it to output
	if (segment.size() != 0) {
        result.push_back(segment);
    }
    return result;
}

// returns true if the target sub string is at input_index in input.
bool string_matches(const std::string& input, const uint input_index, const std::string& target) {
	for (uint x = 0; x < target.size(); x++)
		if (input[x + input_index] != target[x])
			return false;
	return true;
}

// returns true if target is in input
bool string_contains(const std::string& input, const std::string& target) {
	for (uint i = 0; i < input.size(); i++) {
		if ((target.size() + i) > input.size())
			return false;
		if (string_matches(input, i, target))
			return true;
	}
	return false;
}