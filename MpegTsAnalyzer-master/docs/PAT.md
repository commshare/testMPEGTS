

- https://stackoverflow.com/questions/31220892/reading-information-from-pat-section-mpeg-ts


- I'm writing a MPEG-TS file parser and I'm stuck on getting program_numbers and PIDs from the PAT section.
 I'm using a packet analyser to compare my results.

For example, here's a PAT packet
```
47 40 00 16 00 00 B0 31 00 14 D7 00 00 00 00 E0
10 00 01 E0 24 00 02 E0 25 00 03 E0 30 00 04 E0
31 00 1A E0 67 00 1C E0 6F 43 9D E3 F1 43 A3 E3
F7 43 AC E4 00 C3 69 A6 D8 FF FF FF FF FF FF FF
FF FF FF FF FF FF FF FF FF FF FF FF FF FF FF FF
FF FF FF FF FF FF FF FF FF FF FF FF FF FF FF FF
FF FF FF FF FF FF FF FF FF FF FF FF FF FF FF FF
FF FF FF FF FF FF FF FF FF FF FF FF FF FF FF FF
FF FF FF FF FF FF FF FF FF FF FF FF FF FF FF FF
FF FF FF FF FF FF FF FF FF FF FF FF FF FF FF FF
FF FF FF FF FF FF FF FF FF FF FF FF 


```

- d 

```

Each program_number is 16 bits and is followed by 16 bits consisting of 3 x '1' bits and a 13 bit program_map_pid 
(or network_pid ifprogram_number`=0)

Start at offset 13 in your dump and read pairs of 16-bit words, masking out the top 3 bits of the second word.

e.g.

offset   bytes          words        program_number pid
======   ===========    =========    ============== ======================
000D:    00 00 E0 10 => 0000 E010 => 0000           0010 (network_pid)
0011:    00 01 E0 24 => 0001 E024 => 0001           0024 (program_map_pid)
0015:    00 02 E0 25 => 0002 E025 => 0002           0025 (program_map_pid)
0019:    etc..
001D:    etc..
0021:    etc..
0025:    00 1C E0 6F => 001C E06F => 001C           006F (program_map_pid)
0029:    43 9D E3 F1 => 439D E3F1 => 439D           03F1 (program_map_pid)
002D:    etc..
etc..
In theory it is more complicated than this as there can be multiple program association sections in a PAT and the above will only help with the 1st section.

For more details see section 2.4.4.3 of ISO/IEC 13818-1, specifically table 2-25.
```