package main

import (
  "fmt"
  "os"
)

func main() {
  // if no args, start gui
  if len(os.Args) == 1 {
    StartGui()
    return
  }
  // if args, run command
  if os.Args[1] == "schem" {
    // todo: generate schematic
    fmt.Println("wip")
  } else {
    printHelp()
  }
}

func printHelp() {
  helpMessage := `
Usage: rails command [arguments]
Commands:
  Name - schem
    Description - Optional cli command for generating schematics in addition to the gui
    Usage - rails schem [path/to/file_name.rails] [path/to/output_file_name.schem]

For more information, visit [https://github.com/Data-Corruption/Rails]
`
  fmt.Print(helpMessage)
}
