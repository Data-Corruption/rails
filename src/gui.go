package main

import (
	"fmt"
	"time"
  "image/color"
	imgui "github.com/AllenDang/imgui-go"
  g "github.com/AllenDang/giu"
	dialog "github.com/sqweek/dialog"
)

const MIN_SCALE = 0.5;
const MAX_SCALE = 1.2;
const scalarAggressiveness = 1.2;

var (
	cpu RailsEmulator = RailsEmulator{}
	// gui stuff
	wnd *g.MasterWindow
	font *g.FontInfo
	fontScale float32 = 0.8
	scalar float32 // yes the way i'm doing this is kinda jank but it works so *dab*
	viewAsUnsigned bool = false
	io_input_data int32 = 0
	io_input_address int32 = 0
	loadedProgramName string = ""
)

func updateScalar() {
	scalar = float32(CenterGravityAdjust(float64(fontScale), MIN_SCALE - 0.6, MAX_SCALE + 1.3, scalarAggressiveness))
}

// force gui to update every 1/30th of a second
func updater() {
	ticker := time.NewTicker(time.Second / 30)
	for {
		g.Update()
		<-ticker.C
	}
}

// ---- GUI Callbacks ----

func zoomIn() {
	fontScale += 0.1
	if fontScale > MAX_SCALE { fontScale = MAX_SCALE }
	updateScalar()
	imgui.IO.SetFontGlobalScale(g.Context.IO(), fontScale)
}
func zoomOut() {
	fontScale -= 0.1
	if fontScale < MIN_SCALE { fontScale = MIN_SCALE }
	updateScalar()
	imgui.IO.SetFontGlobalScale(g.Context.IO(), fontScale)
}

func loadProgram() {
	// get file path
	if cpu.IsBusy { return }
	file, err := dialog.File().Filter("Rails Program", "rails").Load()
	if err != nil { return }
	cpu.Reset()
	cpu.State.ProgramLength, err = AssembleFile(file, cpu.State.Prom[:])
	if err != nil {
		fmt.Println("Error loading program: " + err.Error())
		cpu.Reset()
		return
	}
	loadedProgramName = fmt.Sprintf("%s (%d instructions)", file, cpu.State.ProgramLength)
}

func printProgram() {
	if (cpu.State.ProgramLength == 0) || (loadedProgramName == "") {
		fmt.Println("No program loaded");
		return
	}
	fmt.Printf("Printing %s ...\n", loadedProgramName)
	for i := 0; i < int(cpu.State.ProgramLength); i++ {
		fmt.Printf("%s: %s\n", NumberToString(uint64(i), 3, " ", 10), cpu.InstructionToString(uint8(i)))
	}
}

func saveSnapshot() {
	path, err := dialog.File().Filter("Rails Snapshot binary", "bin").SetStartFile("snapshot.bin").Save()
	if err != nil { return }
	// if file doesn't end with .bin, add it
	if path[len(path)-4:] != ".bin" { path += ".bin" }
	err = cpu.SaveState(path)
	if err != nil {
		fmt.Println("Error saving snapshot: " + err.Error())
		return
	}
}
func loadSnapshot() {
	path, err := dialog.File().Filter("Rails Snapshot binary", "bin").Load()
	if err != nil { return }
	err = cpu.LoadState(path)
	if err != nil {
		fmt.Println("Error loading snapshot: " + err.Error())
		return
	}
}

func saveProgramAsSchematic() {
	fmt.Println("wip - saveProgramAsSchematic")
}

func enterIoData() {
	if io_input_address < 0 || io_input_address > 15 {
		fmt.Println("Invalid IO address. Must be between 0 and 15")
		return
	}
	// clamp test number to -128 to 127
	if io_input_data > 127 {
		io_input_data = 127
	} else if io_input_data < -128 {
		io_input_data = -128
	}
	cpu.State.InRegs[io_input_address] = uint8(io_input_data)
}
func programComboChanged() {
	fmt.Println("wip - programComboChanged")
}

func eval() {
	if cpu.IsBusy { return }
	cpu.Eval()
}
func evalUntil(stopType StopType) {
	if cpu.State.ProgramLength == 0 {
		fmt.Println("No program appears to be loaded, or it is empty")
		return
	}
	if cpu.IsBusy { return }
	cpu.Mutex.Unlock()
	go cpu.EvalUntil(stopType)
	cpu.Mutex.Lock()
}
func evalUntilIO() { evalUntil(IO) }
func evalUntilEXIT() { evalUntil(EXIT) }
func reset() { cpu.Reset() }
func stop() { cpu.ShouldStop = true }

// ---- GUI Widget Builders ----

type CellContentFunc func(offset int) g.Widget

func buildTable(rows int, cols int, contentFunc CellContentFunc) []*g.TableRowWidget {
  tableRows := make([]*g.TableRowWidget, rows)
  for i := 0; i < rows; i++ {
    var cells []g.Widget
    for j := 0; j < cols; j++ {
      offset := i*cols + j
      cells = append(cells, contentFunc(offset))
    }
    tableRows[i] = g.TableRow(cells...)
  }
  return tableRows
}

func arrayValueToString(index uint64, array []uint8, index_out_width uint16) string {
	a := NumberToString(index, index_out_width, "0", 10)
	if (viewAsUnsigned) {
		return fmt.Sprintf("%s: %s", a, NumberToString(array[index], 3, "0", 10))
	} else {
		return fmt.Sprintf("%s: %s", a, NumberToString(int8(array[index]), 4, " ", 10))
	}
}

// ---- GUI Loop ----

func loop() {
	cpu.Mutex.Lock()

  g.PushColorWindowBg(color.RGBA{R: 0, G: 0, B: 20, A: 255})
	g.SingleWindowWithMenuBar().Layout(
		g.MenuBar().Layout(
			g.Menu("Program").Layout(
				g.MenuItem("Open").OnClick(loadProgram),
				g.MenuItem("Save Snapshot").OnClick(saveSnapshot),
				g.MenuItem("Load Snapshot").OnClick(loadSnapshot),
				g.MenuItem("Save As Schematic").OnClick(saveProgramAsSchematic),
				g.MenuItem("Print To Console").OnClick(printProgram),
			),
			g.Menu("Settings").Layout(
				g.Row(
					g.Label("Zoom Level:"),
					g.Button("-").OnClick(zoomOut),
					g.Button("+").OnClick(zoomIn),
				),
				g.Row(g.Label("View data as unsigned:"), g.Checkbox("", &viewAsUnsigned)),
			),
		),
		g.TreeNode("IO Ports").Flags(g.TreeNodeFlagsCollapsingHeader).Layout(
			g.Label("Input Registers"),
			g.Row(
				g.Label("Address:"),
				g.InputInt(&io_input_address).Size(100),
				g.Label("Value:"),
				g.InputInt(&io_input_data).Size(100),
				g.Button("Enter").OnClick(enterIoData),
			),
			g.Table().FastMode(true).
				Rows(buildTable(4, 4, func(offset int) g.Widget {
					return g.Label(arrayValueToString(uint64(offset), cpu.State.InRegs[:], 2))
				})...).Size(500, 120 * scalar),
			g.Label("Output Registers"),
			g.Table().FastMode(true).
				Rows(buildTable(4, 4, func(offset int) g.Widget {
					return g.Label(arrayValueToString(uint64(offset), cpu.State.OutRegs[:], 2))
				})...).Size(500, 120 * scalar),
		),
		g.TreeNode("Regfile").Flags(g.TreeNodeFlagsCollapsingHeader).Layout(
			g.Table().FastMode(true).
				Rows(buildTable(4, 4, func(offset int) g.Widget {
					return g.Label(arrayValueToString(uint64(offset), cpu.State.Regfile[:], 2))
				})...).Size(500, 120 * scalar),
		),
		g.TreeNode("RAM").Flags(g.TreeNodeFlagsCollapsingHeader).Layout(
			g.Table().FastMode(true).
				Rows(buildTable(32, 8, func(offset int) g.Widget {
					return g.Label(arrayValueToString(uint64(offset), cpu.State.Ram[:], 3))
				})...).Size(0, 995 * scalar),
		),
		g.TreeNode("Other").Flags(g.TreeNodeFlagsCollapsingHeader).Layout(
			g.Row(g.Label("Current Instruction:"), g.Label(cpu.InstructionToString(cpu.State.Pc)),),
			g.Row(g.Label("Program Counter:"), g.Label(NumberToString(cpu.State.Pc, 3, " ", 10)),),
			g.Row(g.Label("Carry Flag:"), g.Label(fmt.Sprintf("%t", cpu.State.CarryFlag)),),
		),
		g.Row(g.Label("Loaded Program:"), g.Label(loadedProgramName)),
		g.Row(
			g.Button("Eval").OnClick(eval).Disabled(cpu.IsBusy),
			g.Button("Eval Until IO").OnClick(evalUntilIO).Disabled(cpu.IsBusy),
			g.Button("Eval Until EXIT").OnClick(evalUntilEXIT).Disabled(cpu.IsBusy),
			g.Button("Reset").OnClick(reset).Disabled(cpu.IsBusy),
			g.Button("Stop").OnClick(stop).Disabled(!cpu.IsBusy),
		),
	)
  g.PopStyleColor()

	cpu.Mutex.Unlock()
}

// ---- GUI Entry Point ----

func StartGui() {

	// create window
	wnd = g.NewMasterWindow("Rails Emulator", 1200, 600, 0).
		RegisterKeyboardShortcuts(
			g.WindowShortcut{ Key: g.KeyMinus, Modifier: g.ModControl, Callback: func() { zoomOut() }},
			g.WindowShortcut{ Key: g.KeyEqual, Modifier: g.ModControl, Callback: func() { zoomIn() }},
		)

	// load font
	byteData, err := Asset("FiraCode-Retina.ttf")
	if err != nil {
		fmt.Println("Error loading font, using default font")
	} else {
		g.Context.FontAtlas.SetDefaultFontFromBytes(byteData, 16)
	}

	// set scale stuff
	imgui.IO.SetFontGlobalScale(g.Context.IO(), fontScale)
	updateScalar()

	// start gui loop
	go updater()
	wnd.Run(loop)
}