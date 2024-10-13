(function (ace) {
  var oop = ace.require("ace/lib/oop");
  var TextMode = ace.require("ace/mode/text").Mode;
  var RailsAssemblyHighlightRules = ace.require("ace/mode/rails_assembly_highlight_rules").RailsAssemblyHighlightRules;

  var Mode = function () {
    this.HighlightRules = RailsAssemblyHighlightRules;
  };
  oop.inherits(Mode, TextMode);

  (function () {
    this.$id = "ace/mode/rails_assembly";
  }).call(Mode.prototype);

  ace.define("ace/mode/rails_assembly", ["require", "exports", "module"], function (require, exports, module) {
    exports.Mode = Mode;
  });
})(ace);
