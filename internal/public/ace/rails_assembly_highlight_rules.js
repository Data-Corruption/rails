(function (ace) {
  var oop = ace.require("ace/lib/oop");
  var TextHighlightRules = ace.require("ace/mode/text_highlight_rules").TextHighlightRules;

  var RailsAssemblyHighlightRules = function () {
    this.$rules = {
      start: [
        // Error if a line's first token is not a comment, label, or opcode
        {
          token: "error.invalid",
          regex:
            "^\\s*(?!;|[\\w.]+:|\\b(?:add|addc|sub|swb|nand|rsft|imm|ld|ldim|st|stim|beq|bgt|jmpl|in|out|nop|mov|jmp|exit)\\b)\\S+",
          caseInsensitive: true,
        },
        {
          token: "keyword.control.assembly",
          regex: "\\b(?:add|addc|sub|swb|nand|rsft|imm|ld|ldim|st|stim|beq|bgt|jmpl|in|out|nop|mov|jmp|exit)\\b",
          caseInsensitive: true,
        },
        // Register arguments with range check (0-15)
        {
          regex: "\\br-?\\d+\\b",
          token: function (value) {
            var num = parseInt(value.substring(1), 10);
            if (num >= 0 && num <= 15) return "variable.parameter.register.assembly";
            else return "invalid.illegal.register";
          },
          caseInsensitive: true,
        },
        // Numeric values with range check (-128 to 255)
        {
          regex: "\\b-?\\d+\\b",
          token: function (value) {
            var num = parseInt(value, 10);
            if (num >= -128 && num <= 255) return "constant.numeric.assembly";
            else return "invalid.illegal.numeric";
          },
        },
        {
          token: "entity.name.function.assembly",
          regex: "^[\\w.]+:",
        },
        {
          token: "comment.assembly",
          regex: ";.*$",
        },
      ],
    };

    this.normalizeRules();
  };

  oop.inherits(RailsAssemblyHighlightRules, TextHighlightRules);

  ace.define(
    "ace/mode/rails_assembly_highlight_rules",
    ["require", "exports", "module"],
    function (require, exports, module) {
      exports.RailsAssemblyHighlightRules = RailsAssemblyHighlightRules;
    }
  );
})(ace);
