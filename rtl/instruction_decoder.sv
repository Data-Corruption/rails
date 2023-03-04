module instruction_decoder (
    input clk,
    input clk_en,
    input sync_rst,
    input [15:0] instruction,
    // ram interface
    output [7:0] ram_address,
    output [7:0] ram_write_data,
    input [7:0] ram_read_data,
    output ram_operation, // 0 = read, 1 = write
    output ram_en,
    // io interface
    output [3:0] io_regs_address,
    output [7:0] io_regs_write_data,
    input [7:0] io_regs_read_data,
    output io_regs_operation, // 0 = read, 1 = write
    output io_regs_en,
    // program counter
    output [7:0] pc_overwrite_data,
    output pc_overwrite_en,
    input [7:0] pc_out,
    // regfile
    output regfile_write_en,
    output [3:0] regfile_write_address,
    output [7:0] regfile_write_data,
    output [3:0] regfile_read_address_a,
    output [3:0] regfile_read_address_b,
    input [7:0] regfile_read_data_a,
    input [7:0] regfile_read_data_b,
    // alu
    input [7:0] alu_out,
    input alu_condition,
    output [7:0] alu_in_a,
    output [7:0] alu_in_b
);

localparam FALSE = 1'b0;
localparam TRUE = 1'b1;
localparam READ = 1'b0;
localparam WRITE = 1'b1;

wire [3:0] opcode = instruction[15:12];
wire [3:0] a_operand = instruction[11:8];
wire [3:0] b_operand = instruction[7:4];
wire [3:0] c_operand = instruction[3:0];
wire [7:0] imm = instruction[11:4];

// temps

wire [7:0] temp_ram_address;
wire [7:0] temp_ram_write_data;
wire temp_ram_operation;
wire temp_ram_en;

wire [3:0] temp_io_regs_address;
wire [7:0] temp_io_regs_write_data;
wire temp_io_regs_operation;
wire temp_io_regs_en;

wire [7:0] temp_pc_overwrite_data;
wire temp_pc_overwrite_en;

wire temp_regfile_write_en;
wire [3:0] temp_regfile_write_address;
wire [7:0] temp_regfile_write_data;
wire [3:0] temp_regfile_read_address_a;
wire [3:0] temp_regfile_read_address_b;

wire [7:0] temp_alu_in_a;
wire [7:0] temp_alu_in_b;

// assign outputs

assign ram_address = temp_ram_address;
assign ram_write_data = temp_ram_write_data;
assign ram_operation = temp_ram_operation;
assign ram_en = temp_ram_en;

assign io_regs_address = temp_io_regs_address;
assign io_regs_write_data = temp_io_regs_write_data;
assign io_regs_operation = temp_io_regs_operation;
assign io_regs_en = temp_io_regs_en;

assign pc_overwrite_data = temp_pc_overwrite_data;
assign pc_overwrite_en = temp_pc_overwrite_en;

assign regfile_write_en = temp_regfile_write_en;
assign regfile_write_address = temp_regfile_write_address;
assign regfile_write_data = temp_regfile_write_data;
assign regfile_read_address_a = temp_regfile_read_address_a;
assign regfile_read_address_b = temp_regfile_read_address_b;

assign alu_in_a = temp_alu_in_a;
assign alu_in_b = temp_alu_in_b;

always_comb begin
    case (opcode) 
        4'b0110: begin // IMM
            temp_ram_en = FALSE;
            temp_io_regs_en = FALSE;
            temp_pc_overwrite_en = FALSE;
            temp_regfile_write_en = TRUE;
            temp_regfile_write_address = c_operand;
            temp_regfile_write_data = imm;
        end
        4'b0111: begin // LD
            temp_ram_en = TRUE;
            temp_io_regs_en = FALSE;
            temp_pc_overwrite_en = FALSE;
            temp_regfile_write_en = TRUE;
            temp_regfile_write_address = c_operand;
            temp_regfile_read_address_a = a_operand;
            temp_regfile_write_data = ram_read_data;
            temp_ram_address = regfile_read_data_a;
            temp_ram_operation = READ;
        end
        4'b1000: begin // LDIM
            temp_ram_en = TRUE;
            temp_io_regs_en = FALSE;
            temp_pc_overwrite_en = FALSE;
            temp_regfile_write_en = TRUE;
            temp_regfile_write_address = c_operand;
            temp_regfile_write_data = ram_read_data;
            temp_ram_address = imm;
            temp_ram_operation = READ;
        end
        4'b1001: begin // ST
            temp_ram_en = TRUE;
            temp_io_regs_en = FALSE;
            temp_pc_overwrite_en = FALSE;
            temp_regfile_write_en = FALSE;
            temp_regfile_read_address_a = a_operand;
            temp_regfile_read_address_b = b_operand;
            temp_ram_address = regfile_read_data_a;
            temp_ram_write_data = regfile_read_data_b;
            temp_ram_operation = WRITE;
        end
        4'b1010: begin // STIM
            temp_ram_en = TRUE;
            temp_io_regs_en = FALSE;
            temp_pc_overwrite_en = FALSE;
            temp_regfile_write_en = FALSE;
            temp_regfile_read_address_a = c_operand;
            temp_ram_address = imm;
            temp_ram_write_data = regfile_read_data_a;
            temp_ram_operation = WRITE;
        end
        4'b1011: begin // BEQ
            temp_ram_en = FALSE;
            temp_io_regs_en = FALSE;
            temp_pc_overwrite_en = alu_condition;
            temp_regfile_write_en = FALSE;
            temp_regfile_read_address_a = 4'b1111;
            temp_regfile_read_address_b = c_operand;
            temp_pc_overwrite_data = imm;
            temp_alu_in_a = regfile_read_data_a;
            temp_alu_in_b = regfile_read_data_b;
        end
        4'b1100: begin // BGT
            temp_ram_en = FALSE;
            temp_io_regs_en = FALSE;
            temp_pc_overwrite_en = alu_condition;
            temp_regfile_write_en = FALSE;
            temp_pc_overwrite_data = imm;
            temp_regfile_read_address_a = 4'b1111;
            temp_regfile_read_address_b = c_operand;
            temp_alu_in_a = regfile_read_data_a;
            temp_alu_in_b = regfile_read_data_b;
        end
        4'b1101: begin // JMPL
            temp_ram_en = FALSE;
            temp_io_regs_en = FALSE;
            temp_pc_overwrite_en = TRUE;
            temp_regfile_write_en = TRUE;
            temp_pc_overwrite_data = regfile_read_data_a;
            temp_regfile_write_address = c_operand;
            temp_regfile_write_data = pc_out + 1;
            temp_regfile_read_address_a = a_operand;
        end
        4'b1110: begin // IN
            temp_ram_en = FALSE;
            temp_io_regs_en = TRUE;
            temp_pc_overwrite_en = FALSE;
            temp_regfile_write_en = TRUE;
            temp_regfile_write_address = c_operand;
            temp_regfile_write_data = io_regs_read_data;
            temp_io_regs_address = a_operand;
            temp_io_regs_operation = READ;
        end
        4'b1111: begin // OUT
            temp_ram_en = FALSE;
            temp_io_regs_en = TRUE;
            temp_pc_overwrite_en = FALSE;
            temp_regfile_write_en = FALSE;
            temp_regfile_read_address_b = b_operand;
            temp_io_regs_address = c_operand;
            temp_io_regs_write_data = regfile_read_data_b;
            temp_io_regs_operation = WRITE;
        end
        default: begin // ADD, ADDC, SUB, SWB, NAND, RSFT
            temp_ram_en = FALSE;
            temp_io_regs_en = FALSE;
            temp_pc_overwrite_en = FALSE;
            temp_regfile_write_en = TRUE;
            temp_regfile_write_address = c_operand;
            temp_regfile_write_data = alu_out;
            temp_regfile_read_address_a = a_operand;
            temp_regfile_read_address_b = b_operand;
        end
    endcase
end

endmodule