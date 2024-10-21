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
  }

  Step() {
    const instruction = this.ProgramRom[this.ProgramCounter];
    // if 1101 0000 0000 0000 (exit instruction) is encountered, stop the emulator
    if (instruction === 0xd000n) {
      return false;
    }

    // check for breakpoints
    if (this.isRunning) {
      if (this.Breakpoints.includes(Number(this.ProgramCounter))) {
        this.isRunning = false;
        console.log(`Breakpoint hit at ${this.ProgramCounter}`);
        return false;
      }
    }

    const BYTE_MASK = 0x00ffn;
    const opcode = (instruction & 0xf000n) >> 12n;
    const a = (instruction & 0x0f00n) >> 8n;
    const b = (instruction & 0x00f0n) >> 4n;
    const c = instruction & 0x000fn;
    const imm = (instruction & 0x0ff0n) >> 4n;

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
          this.CarryFlag = false;
          return true;
        }
        break;
      case 12n: // BGT
        if (this.RegisterFile[15] > this.RegisterFile[Number(c)]) {
          this.ProgramCounter = imm;
          this.CarryFlag = false;
          return true;
        }
        break;
      case 13n: // JMPL
        this.RegisterFile[Number(c)] = this.ProgramCounter + 1n;
        this.RegisterFile[0] = 0n; // ensure reg 0 is always 0
        this.ProgramCounter = this.RegisterFile[Number(a)];
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
        return false;
    }
    this.RegisterFile[0] = 0n; // ensure reg 0 is always 0
    this.ProgramCounter++;
    return true;
  }

  UpdateUI() {
    console.log("UI update logic here");
  }

  async Reset() {
    await this.Stop();
    this.RegisterFile.fill(0n);
    this.Ram.fill(0n);
    // don't reset input registers.
    this.OutRegisters.fill(0n);
    this.ProgramCounter = 0n;
    this.CarryFlag = false;
    console.log("Emulator reset");
  }
}
