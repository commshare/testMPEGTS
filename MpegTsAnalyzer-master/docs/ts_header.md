
# TS包包头

- TS包的包头提供关于传输方面的信息：同步、有无差错、有无加扰、PCR（节目参考时钟）等标志。TS包的包头长度不固定，前32比特（4个字节）固定，后面可能跟有自适应字段（适配域）。32个比特（4个字节）是最小包头。包头的结构固定如下：

同步字节    传输错误指示  开始指示  传输优先级  PID   加扰控制  适配域控制  连续性计数器  适配域
 8             1          1         1       13     2        2             4 

```
typedef struct TS_packet_header

{

unsigned sync_byte : 8;

unsigned transport_error_indicator : 1;

unsigned payload_unit_start_indicator : 1;

unsigned transport_priority : 1;

unsigned PID : 13;

unsigned transport_scrambling_control : 2;

unsigned adaption_field_control : 2;

unsigned continuity_counter : 4;

} TS_packet_header;
```


- 
sync_byte （同步字节）：固定为0100 0111 (0x47)；该字节由解码器识别，使包头和有效负载可相互分离。
transport_error_indicator（传输错误指示）：‘1’表示在相关的传输包中至少有一个不可纠正的错误位。当被置1后，在错误被纠正之前不能重置为0。
payload_unit_start_indicator（开始指示）：为1时，在前4个字节之后会有一个调整字节，其的数值为后面调整字段的长度length。因此有效载荷开始的位置应再偏移1+[length]个字节。

transport_priority（传输优先级）：‘1’表明优先级比其他具有相同PID 但此位没有被置‘1’的分组高。
PID：指示存储与分组有效负载中数据的类型。PID 值 0x0000—0x000F 保留。其中0x0000为PAT保留；0x0001为CAT保留；0x1fff为分组保留，即空包。

transport_scrambling_control（加扰控制）：表示TS流分组有效负载的加密模式。空包为‘00’，如果传输包包头中包括调整字段，不应被加密。
adaptation_field_control（适配域控制）：表示包头是否有调整字段或有效负载。‘00’为ISO/IEC未来使用保留；‘01’仅含有效载荷，无调整字段；‘10’ 无有效载荷，仅含调整字段；‘11’ 调整字段后为有效载荷，调整字段中的前一个字节表示调整字段的长度length，有效载荷开始的位置应再偏移[length]个字节。空包应为‘10’。
continuity_counter（连续性计数器）：随着每一个具有相同PID的TS流分组而增加，当它达到最大值后又回复到0。范围为0~15。

适配域：