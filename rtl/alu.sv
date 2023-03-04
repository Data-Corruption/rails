// 8 bit arithmetic logic unit
module alu (
    input clk,
    input clk_en,
    input sync_rst,

    input [7:0] a,
    input [7:0] b,
    input [3:0] op,

    output [7:0] out,
    output condition
);

// buffer that stores the carry out of the previous ALU operation
reg carry_out_buffer;
wire next_carry_out = (sync_rst) ? 1'b0 : temp_out[8];
wire carry_out_buffer_trigger = sync_rst || clk_en;

always_ff @(posedge clk) begin
    if (carry_out_buffer_trigger) begin
        carry_out_buffer <= next_carry_out;
    end
end

wire [8:0] temp_a = {1'b0, a};
wire [8:0] temp_b = {1'b0, b};
logic [8:0] temp_out;

always_comb begin
    case (op)
        4'b0001: temp_out = temp_a + temp_b + carry_out_buffer;  // ADDC
        4'b0010: temp_out = temp_a + ~temp_b + 1;                // SUB
        4'b0011: temp_out = temp_a + ~temp_b + carry_out_buffer; // SWB
        4'b0100: temp_out = ~temp_a | ~temp_b;                   // NAND
        4'b0101: temp_out = {2'b0, a[6:1]};                      // RSFT
        4'b1011: temp_out = {8'b0, (a == b)};                    // Equal
        4'b1100: temp_out = {8'b0, (a > b)};                     // Greater
        default: temp_out = temp_a + temp_b;                     // ADD
    endcase
end

assign out = temp_out[7:0];
assign condition = temp_out[0];

endmodule