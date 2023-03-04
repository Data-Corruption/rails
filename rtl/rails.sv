// ------------------------------
// Rails CPU - A simple 8-bit CPU 
// ------------------------------

module cpu (
    input clk,
    input clk_en,
    input sync_rst,

    input [15:0] instruction,
    output [7:0] program_counter,

    // ram interface
    output [7:0] ram_address,
    input [7:0] ram_read_data,
    output [7:0] ram_write_data,
    output ram_operation, // 0 = read, 1 = write
    output ram_en,

    // io registers interface
    output [3:0] io_regs_address,
    input [7:0] io_regs_read_data,
    output [7:0] io_regs_write_data,
    output io_regs_operation, // 0 = read, 1 = write
    output io_regs_en
);

// temps

wire [7:0] temp_pc;

wire [7:0] temp_ram_address;
wire [7:0] temp_ram_write_data;
wire temp_ram_operation;
wire temp_ram_en;

wire [3:0] temp_io_regs_address;
wire [7:0] temp_io_regs_write_data;
wire temp_io_regs_operation;
wire temp_io_regs_en;

// assign outputs

assign program_counter = temp_pc;

assign ram_address = temp_ram_address;
assign ram_write_data = temp_ram_write_data;
assign ram_operation = temp_ram_operation;
assign ram_en = temp_ram_en;

assign io_regs_address = temp_io_regs_address;
assign io_regs_write_data = temp_io_regs_write_data;
assign io_regs_operation = temp_io_regs_operation;
assign io_regs_en = temp_io_regs_en;

// instruction decoder
instruction_decoder instruction_decoder_inst (
    .clk(clk),
    .clk_en(clk_en),
    .sync_rst(sync_rst),
    .instruction(instruction),
    // ram interface
    .ram_address(temp_ram_address),
    .ram_write_data(temp_ram_write_data),
    .ram_read_data(ram_read_data),
    .ram_operation(temp_ram_operation),
    .ram_en(temp_ram_en),
    // io registers interface
    .io_regs_address(temp_io_regs_address),
    .io_regs_write_data(temp_io_regs_write_data),
    .io_regs_read_data(io_regs_read_data),
    .io_regs_operation(temp_io_regs_operation),
    .io_regs_en(temp_io_regs_en),
    // program counter
    .pc_overwrite_data(pc_overwrite_data),
    .pc_overwrite_en(pc_overwrite_en),
    .pc_out(temp_pc),
    // regfile
    .regfile_write_en(regfile_write_en),
    .regfile_write_address(regfile_write_address),
    .regfile_write_data(regfile_write_data),
    .regfile_read_address_a(regfile_read_address_a),
    .regfile_read_address_b(regfile_read_address_b),
    .regfile_read_data_a(regfile_read_data_a),
    .regfile_read_data_b(regfile_read_data_b),
    // alu
    .alu_out(alu_out),
    .alu_condition(alu_condition),
    .alu_in_a(alu_in_a),
    .alu_in_b(alu_in_b)
);

// program counter
wire [7:0] pc_overwrite_data;
wire pc_overwrite_en;

program_counter pc_inst (
    .clk(clk),
    .clk_en(clk_en),
    .sync_rst(sync_rst),
    .overwrite_data(pc_overwrite_data),
    .overwrite_en(pc_overwrite_en),
    .out(temp_pc)
);

// register file
wire regfile_write_en;
wire [3:0] regfile_write_address;
wire [7:0] regfile_write_data;

wire [3:0] regfile_read_address_a;
wire [3:0] regfile_read_address_b;

wire [7:0] regfile_read_data_a;
wire [7:0] regfile_read_data_b;

regfile regfile_inst (
    .clk(clk),
    .clk_en(clk_en),
    .sync_rst(sync_rst),
    .read_address_a(regfile_read_address_a),
    .read_address_b(regfile_read_address_b),
    .write_address(regfile_write_address),
    .write_data(regfile_write_data),
    .write_en(regfile_write_en),
    .read_data_a(regfile_read_data_a),
    .read_data_b(regfile_read_data_b)
);


// alu
wire [7:0] alu_out;
wire alu_condition;

wire [7:0] alu_in_a;
wire [7:0] alu_in_b;

alu alu_inst (
    .clk(clk),
    .clk_en(clk_en),
    .sync_rst(sync_rst),
    .a(alu_in_a),
    .b(alu_in_b),
    .op(opcode),
    .out(alu_out),
    .condition(alu_condition)
);
    
endmodule