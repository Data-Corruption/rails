// register file, address 0 is always 0
module regfile (
    input clk,
    input clk_en,
    input sync_rst,

    input [3:0] read_address_a,
    input [3:0] read_address_b,
    input [3:0] write_address,
    input [7:0] write_data,

    input write_en,

    output [7:0] read_data_a,
    output [7:0] read_data_b
);

reg [7:0] registers [0:15];
wire [7:0] next_reg = (sync_rst) ? 0 : write_data;
wire write_trigger = (clk_en & write_en) | sync_rst;

always @(posedge clk) begin
    if (write_trigger) begin
        registers[write_address] <= next_reg;
    end
end

assign read_data_a = registers[read_address_a];
assign read_data_b = registers[read_address_b];

endmodule