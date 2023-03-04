module program_counter (
    input clk,
    input clk_en,
    input sync_rst,

    input [7:0] overwrite_data,
    input overwrite_en,

    output [7:0] out
);

reg [7:0] counter;
wire counter_input = (overwrite_en) ? overwrite_data : counter + 1;
wire next_counter = (sync_rst) ? 0 : counter_input;
wire counter_trigger = sync_rst || clk_en;

always @(posedge clk) begin
    if (counter_trigger) begin
        counter <= next_counter;
    end
end

endmodule