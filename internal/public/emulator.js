// Emulator represents a generic emulator that can be started, stopped, individually stepped, and reset.
class Emulator {
  constructor() {
    this.StepRateLimit = 100; // min ms between each execution step when running
    this.UiUpdateLimit = 1000; // min ms between UI updates
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
      this.#init();
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

  // init() initializes the emulator and starts the main loop.
  #init() {
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

class RailsEmulator extends Emulator {
  constructor() {
    super();
    this.ProgramRom = new Uint16Array(256);
    this.RegisterFile = new Uint8Array(16);
    this.Ram = new Uint8Array(256);
    this.InRegisters = new Uint8Array(16);
    this.OutRegisters = new Uint8Array(16);
    this.ProgramCounter = 0;
    this.CarryFlag = false;
    this.Breakpoints = [];
  }

  Step() {
    const instruction = this.ProgramRom[this.ProgramCounter];
    // if 1101 0000 0000 0000 (exit instruction) is encountered, stop the emulator
    if (instruction === 0xd000) {
      return false;
    }

    // check for breakpoints
    if (this.isRunning) {
      if (this.Breakpoints.includes(this.ProgramCounter)) {
        this.isRunning = false;
        console.log(`Breakpoint hit at ${this.ProgramCounter}`);
        return false;
      }
    }

    const BYTE_MASK = 0x00ff;
    const opcode = (instruction & 0xf000) >> 12;
    const a = (instruction & 0x0f00) >> 8;
    const b = (instruction & 0x00f0) >> 4;
    const c = instruction & 0x000f;
    const imm = (instruction & 0x0ff0) >> 4;

    switch (opcode) {
      case 0: // ADD
        var result = this.RegisterFile[a] + this.RegisterFile[b];
        this.CarryFlag = result > 255;
        this.RegisterFile[c] = result & BYTE_MASK;
        break;
      case 1: // ADDC
        var carry = this.CarryFlag ? 1 : 0;
        result = this.RegisterFile[a] + this.RegisterFile[b] + carry;
        this.CarryFlag = result > 255;
        this.RegisterFile[c] = result & BYTE_MASK;
        break;
      case 2: // SUB
        result = (this.RegisterFile[a] - this.RegisterFile[b]) & BYTE_MASK;
        this.CarryFlag = this.RegisterFile[a] < this.RegisterFile[b];
        this.RegisterFile[c] = result;
        break;
      case 3: // SWB
        var borrow = this.CarryFlag ? 1 : 0;
        result = (this.RegisterFile[b] - this.RegisterFile[a] - borrow) & BYTE_MASK;
        this.CarryFlag = this.RegisterFile[b] < this.RegisterFile[a] + borrow;
        this.RegisterFile[c] = result;
        break;
      case 4: // NAND
        this.RegisterFile[c] = ~(this.RegisterFile[a] & this.RegisterFile[b]) & BYTE_MASK;
        break;
      case 5: // RSFT
        this.RegisterFile[c] = this.RegisterFile[a] >>> 1;
        break;
      case 6: // IMM
        this.RegisterFile[c] = imm;
        break;
      case 7: // LD
        this.RegisterFile[c] = this.Ram[this.RegisterFile[a]];
        break;
      case 8: // LDIM
        this.RegisterFile[c] = this.Ram[imm];
        break;
      case 9: // ST
        this.Ram[this.RegisterFile[a]] = this.RegisterFile[b];
        break;
      case 10: // STIM
        this.Ram[imm] = this.RegisterFile[c];
        break;
      case 11: // BEQ
        if (this.RegisterFile[15] === this.RegisterFile[c]) {
          this.ProgramCounter = imm;
          this.CarryFlag = false;
          return true;
        }
        break;
      case 12: // BGT
        if (this.RegisterFile[15] > this.RegisterFile[c]) {
          this.ProgramCounter = imm;
          this.CarryFlag = false;
          return true;
        }
        break;
      case 13: // JMPL
        this.RegisterFile[c] = this.ProgramCounter + 1;
        this.RegisterFile[0] = 0; // ensure reg 0 is always 0
        this.ProgramCounter = this.RegisterFile[a];
        this.CarryFlag = false;
        return true;
      case 14: // IN
        this.RegisterFile[c] = this.InRegisters[a];
        break;
      case 15: // OUT
        this.OutRegisters[a] = this.RegisterFile[b];
        break;
      default:
        console.log(`Invalid opcode: ${opcode}`);
        return false;
    }
    this.RegisterFile[0] = 0; // ensure reg 0 is always 0
    this.ProgramCounter++;
    return true;
  }

  UpdateUI() {
    console.log("UI update logic here");
  }

  async Reset() {
    await this.Stop();
    this.RegisterFile.fill(0);
    this.Ram.fill(0);
    // don't reset input registers.
    this.OutRegisters.fill(0);
    this.ProgramCounter = 0;
    this.CarryFlag = false;
    console.log("Emulator reset");
  }
}
