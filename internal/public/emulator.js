// Emulator represents a generic emulator that can be started, stopped, individually stepped, and reset.
class Emulator {
  constructor() {
    this.StepRateLimit = 250; // min ms between each execution step when running
    this.UiUpdateLimit = 250; // min ms between UI updates
    this.isRunning = false;
    this.stopRequested = false;
    this.lastUiUpdate = 0; // timestamp of the last UI update
    this.initialized = false;
  }

  // ---- Methods intended to be overridden by subclasses ----

  // Step() should execute one step of the emulator. Return false to stop the emulator.
  Step() {
    throw new Error("Step function must be implemented by subclasses");
  }

  // UpdateUI() should update the UI with the current state of the emulator.
  UpdateUI() {
    throw new Error("UI update function must be implemented by subclasses");
  }

  // Reset() should reset the emulator to its initial state.
  async Reset() {
    // async so you can call `await this.Stop();` in the implementation
    throw new Error("Reset function must be implemented by subclasses");
  }

  // ---- Methods that should not be overridden ----

  // Start() starts the emulator. It initializes the emulator if it hasn't been initialized yet.
  Start() {
    if (!this.initialized) {
      this.Init();
    }
    this.isRunning = true;
    this.stopRequested = false;
    console.log("Emulator started");
  }

  /**
   * Stop the emulator with a timeout to wait for it to stop. Usage example:
   * emulator.Stop(3000).then(() => {
   *   console.log("Emulator has stopped");
   * }).catch((error) => {
   *   console.error("Failed to stop emulator:", error);
   * });
   */
  async Stop(timeout = 5000) {
    if (this.isRunning) {
      this.stopRequested = true;
      console.log("Emulator stop requested");

      return new Promise((resolve, reject) => {
        const checkInterval = 10; // ms between checks
        let elapsedTime = 0;

        const checkStopped = () => {
          if (!this.isRunning) {
            console.log("Emulator stopped");
            resolve();
          } else if (elapsedTime >= timeout) {
            console.error("Emulator did not stop within timeout");
            reject(new Error("Emulator did not stop within timeout"));
          } else {
            elapsedTime += checkInterval;
            setTimeout(checkStopped, checkInterval);
          }
        };

        checkStopped();
      });
    }
    return Promise.resolve();
  }

  // Init() initializes the emulator and starts the main loop.
  Init() {
    if (this.initialized) {
      return;
    }
    this.initialized = true;
    this.#mainLoop();
  }

  // mainLoop() is the main loop of the emulator. It executes one step, updates the UI, and schedules the next loop.
  #mainLoop() {
    if (this.stopRequested) {
      this.isRunning = false;
      this.stopRequested = false;
    }
    if (this.isRunning) {
      this.isRunning = this.Step();
    }
    if (Date.now() - this.lastUiUpdate > this.UiUpdateLimit) {
      this.UpdateUI();
      this.lastUiUpdate = Date.now();
    }
    setTimeout(() => this.#mainLoop(), this.StepRateLimit);
  }
}

// RailsEmulator is an emulator for the Rails architecture. It is a subclass of Emulator and implements the Rails instruction set.
// Usage example:
// const emulator = new RailsEmulator();
class RailsEmulator extends Emulator {
  constructor() {
    super();
    this.ProgramRom = Array(256).fill(0n);
    this.RegisterFile = Array(16).fill(0n);
    this.Ram = Array(256).fill(0n);
    this.InRegisters = Array(16).fill(0n);
    this.OutRegisters = Array(16).fill(0n);
    this.ProgramCounter = 0n;
    this.CarryFlag = false;
    this.Breakpoints = [];
    this.stopMessage = null;
    // perf stuff to avoid unnecessary UI updates
    this.modifiedRegisterFile = true;
    this.modifiedRam = true;
    this.modifiedOutRegisters = true;
    this.modifiedProgramCounter = true;
    this.modifiedCarryFlag = true;
  }

  Step() {
    const instruction = this.ProgramRom[this.ProgramCounter];
    // if 1101 0000 0000 0000 (exit instruction) is encountered, stop the emulator
    if (instruction === 0xd000n) {
      console.log(`Exit instruction hit at ${this.ProgramCounter}`);
      this.stopMessage = "Exit Instruction Hit!";
      return false;
    }

    // check for breakpoints
    if (this.isRunning) {
      if (this.Breakpoints.includes(Number(this.ProgramCounter))) {
        this.isRunning = false;
        this.stopMessage = "Breakpoint Hit!";
        console.log(`Breakpoint hit at ${this.ProgramCounter}`);
        return false;
      }
    }

    // if 0000 0000 0000 0000 (nop instruction) is encountered, continue to next instruction
    if (instruction === 0x0000n) {
      this.ProgramCounter++;
      this.modifiedProgramCounter = true;
      return true;
    }

    const BYTE_MASK = 0x00ffn;
    const opcode = (instruction & 0xf000n) >> 12n;
    const a = (instruction & 0x0f00n) >> 8n;
    const b = (instruction & 0x00f0n) >> 4n;
    const c = instruction & 0x000fn;
    const imm = (instruction & 0x0ff0n) >> 4n;

    const regfileModInsts = [0n, 1n, 2n, 3n, 4n, 5n, 6n, 7n, 8n, 13n, 14n];
    const ramModInsts = [9n, 10n];
    const outRegModInsts = [15n];
    const carryModInsts = [0n, 1n, 2n, 3n, 11n, 12n, 13n];

    if (regfileModInsts.includes(opcode)) {
      this.modifiedRegisterFile = true;
    }
    if (ramModInsts.includes(opcode)) {
      this.modifiedRam = true;
    }
    if (outRegModInsts.includes(opcode)) {
      this.modifiedOutRegisters = true;
    }
    if (carryModInsts.includes(opcode)) {
      this.modifiedCarryFlag = true;
    }

    let result;

    switch (opcode) {
      case 0n: // ADD
        result = this.RegisterFile[Number(a)] + this.RegisterFile[Number(b)];
        this.CarryFlag = result > 255n;
        this.RegisterFile[Number(c)] = result & BYTE_MASK;
        break;
      case 1n: // ADDC
        var carry = this.CarryFlag ? 1n : 0n;
        result = this.RegisterFile[Number(a)] + this.RegisterFile[Number(b)] + carry;
        this.CarryFlag = result > 255n;
        this.RegisterFile[Number(c)] = result & BYTE_MASK;
        break;
      case 2n: // SUB
        result = (this.RegisterFile[Number(a)] - this.RegisterFile[Number(b)]) & BYTE_MASK;
        this.CarryFlag = this.RegisterFile[Number(a)] < this.RegisterFile[Number(b)];
        this.RegisterFile[Number(c)] = result;
        break;
      case 3n: // SWB
        var borrow = this.CarryFlag ? 1n : 0n;
        result = (this.RegisterFile[Number(b)] - this.RegisterFile[Number(a)] - borrow) & BYTE_MASK;
        this.CarryFlag = this.RegisterFile[Number(b)] < this.RegisterFile[Number(a)] + borrow;
        this.RegisterFile[Number(c)] = result;
        break;
      case 4n: // NAND
        this.RegisterFile[Number(c)] = ~(this.RegisterFile[Number(a)] & this.RegisterFile[Number(b)]) & BYTE_MASK;
        break;
      case 5n: // RSFT
        this.RegisterFile[Number(c)] = this.RegisterFile[Number(a)] >> 1n;
        break;
      case 6n: // IMM
        this.RegisterFile[Number(c)] = imm;
        break;
      case 7n: // LD
        this.RegisterFile[Number(c)] = this.Ram[Number(this.RegisterFile[Number(a)])];
        break;
      case 8n: // LDIM
        this.RegisterFile[Number(c)] = this.Ram[Number(imm)];
        break;
      case 9n: // ST
        this.Ram[Number(this.RegisterFile[Number(a)])] = this.RegisterFile[Number(b)];
        break;
      case 10n: // STIM
        this.Ram[Number(imm)] = this.RegisterFile[Number(c)];
        break;
      case 11n: // BEQ
        if (this.RegisterFile[15] === this.RegisterFile[Number(c)]) {
          this.ProgramCounter = imm;
          this.modifiedProgramCounter = true;
          this.CarryFlag = false;
          return true;
        }
        break;
      case 12n: // BGT
        if (this.RegisterFile[15] > this.RegisterFile[Number(c)]) {
          this.ProgramCounter = imm;
          this.modifiedProgramCounter = true;
          this.CarryFlag = false;
          return true;
        }
        break;
      case 13n: // JMPL
        this.RegisterFile[Number(c)] = this.ProgramCounter + 1n;
        this.RegisterFile[0] = 0n; // ensure reg 0 is always 0
        this.ProgramCounter = this.RegisterFile[Number(a)];
        this.modifiedProgramCounter = true;
        this.CarryFlag = false;
        return true;
      case 14n: // IN
        this.RegisterFile[Number(c)] = this.InRegisters[Number(a)];
        break;
      case 15n: // OUT
        this.OutRegisters[Number(a)] = this.RegisterFile[Number(b)];
        break;
      default:
        console.log(`Invalid opcode: ${opcode}`);
        this.stopMessage = `Invalid opcode: ${opcode}`;
        return false;
    }
    this.RegisterFile[0] = 0n; // ensure reg 0 is always 0
    this.ProgramCounter++;
    this.modifiedProgramCounter = true;
    return true;
  }

  UpdateUI() {
    if (this.modifiedProgramCounter) {
      const pc = document.getElementById("pc");
      pc.textContent = this.ProgramCounter.toString(10);
      highlightInstruction(this.ProgramCounter);
      this.modifiedProgramCounter = false;
    }
    if (this.modifiedCarryFlag) {
      const carry = document.getElementById("carry");
      carry.textContent = this.CarryFlag ? "1" : "0";
      this.modifiedCarryFlag = false;
    }
    if (this.modifiedOutRegisters) {
      const ioregs = document.getElementById("ioregs");
      for (let i = 0; i < 16; i++) {
        ioregs.children[i].children[1].textContent = this.OutRegisters[i].toString(10);
      }
      this.modifiedOutRegisters = false;
    }
    if (this.modifiedRegisterFile) {
      const regfile = document.getElementById("regfile");
      for (let i = 0; i < 16; i++) {
        regfile.children[i].textContent = this.RegisterFile[i].toString(10);
      }
      this.modifiedRegisterFile = false;
    }
    if (this.modifiedRam) {
      const memory = document.getElementById("memory");
      for (let i = 0; i < 256; i++) {
        memory.children[i].textContent = this.Ram[i].toString(10);
      }
      this.modifiedRam = false;
    }

    if (this.stopMessage) {
      const toasts = document.getElementById("toasts");
      const toast = document.createElement("div");
      toast.classList.add(
        "alert",
        "alert-warning",
        "opacity-100",
        "transition-opacity",
        "duration-500",
        "ease-in-out"
      );
      const span = document.createElement("span");
      span.textContent = this.stopMessage;
      toast.appendChild(span);
      toasts.appendChild(toast);
      // after 3 seconds, start the fade-out transition
      setTimeout(() => {
        toast.classList.replace("opacity-100", "opacity-0");
        toast.addEventListener("transitionend", () => {
          toast.remove();
        });
      }, 3000);
      this.stopMessage = null;
    }
  }

  async Reset() {
    await this.Stop();
    this.RegisterFile.fill(0n);
    this.Ram.fill(0n);
    // don't reset input registers.
    this.OutRegisters.fill(0n);
    this.ProgramCounter = 0n;
    this.CarryFlag = false;
    this.modifiedRegisterFile = true;
    this.modifiedRam = true;
    this.modifiedOutRegisters = true;
    this.modifiedProgramCounter = true;
    this.modifiedCarryFlag = true;
    console.log("Emulator reset");
  }
}
