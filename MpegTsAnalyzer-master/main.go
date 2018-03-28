package main

import (
	"fmt"
	"io"
	"os"

	"./options"
	"./tsparser"
	"gopkg.in/alecthomas/kingpin.v2"
)

const tsPacketSize = 188

func main() {
	var options options.Options
	filename := kingpin.Arg("input", "Input file name.").Required().String()
	options.SetDumpHeader(*kingpin.Flag("dump-ts-header", "Dump TS packet header.").Bool())
	options.SetDumpPayload(*kingpin.Flag("dump-ts-payload", "Dump TS packet payload binary.").Bool())
	options.SetDumpAdaptationField(*kingpin.Flag("dump-adaptation-field", "Dump TS packet adaptation_field detail.").Bool())
	options.SetDumpPsi(*kingpin.Flag("dump-psi", "Dump PSI(PAT/PMT) detail.").Bool())
	options.SetNotDumpTimestamp(*kingpin.Flag("not-dump-timestamp", "Not Dump PCR/PTS/DTS timestamps.").Short('n').Bool())
	kingpin.Parse()

	if err := parseTsFile(*filename, options); err != nil {
		os.Exit(1)
	}
}

func parseTsFile(filename string, options options.Options) error {
	file, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("File open error: %s %s", filename, err)
	}
	fmt.Println("Input file: ", filename)

	pat := tsparser.NewPat()
	pmt := tsparser.NewPmt()

	const patPid = 0x0
	const bufSize = 65536 /*64k？*/
	var pos int64
	buf := make([]byte, bufSize)
	for {
		size, err := file.Read(buf)
		if err == io.EOF {
			break
		} else if err != nil {
			return fmt.Errorf("File read error: %s", err)
		}
		if pos, err = findPat(buf); err != nil {
			continue
		}

		if _, err = file.Seek(pos, 0); err != nil {
			return fmt.Errorf("File seek error: %s", err)
		}

		// Parse PAT
		err = tsparser.BufferPsi(file, &pos, patPid, pat, options)
		err = pat.Parse()
		if err != nil {
			continue
		}
		/*pat里有pmt的pid*/
		pmtPid := pat.PmtPid()

		if _, err = file.Seek(pos, 0); err != nil {
			return fmt.Errorf("File seek error: %s", err)
		}
		fmt.Printf("Detected PAT: PMT pid = 0x%02x\n", pmtPid)

		// Parse PMT
		err = tsparser.BufferPsi(file, &pos, pmtPid, pmt, options)
		err = pmt.Parse()
		if err != nil {
			continue
		}
		/*所有的节目的pid？*/
		programs := pmt.ProgramInfos()
		pcrPid := pmt.PcrPid()

		if _, err = file.Seek(pos, 0); err != nil {
			return fmt.Errorf("File seek error: %s", err)
		}
		fmt.Println("Detected PMT")
		pmt.DumpProgramInfos()

		/*解析pes么？*/
		err = tsparser.BufferPes(file, &pos, pcrPid, programs, options)
		if err != nil {
			return fmt.Errorf("TS parse error: %s", err)
		}
		if size < bufSize {
			break
		}
		pos += bufSize
	}
	return nil
}
/*
            PAT：节目关联表。提供了节目号码和对应PMT表格的PID的对应关系
			一个个遍历TS包，我们找到PID为0的TS包，这个包叫PAT，这个PAT包里包含了PMT的PID号
			TS包的包头长度不固定，前32比特（4个字节）固定，后面可能跟有自适应字段（适配域）。32个比特（4个字节）是最小包头。
*/
func findPat(data []byte) (int64, error) {
	/*要保证至少有2个包+1的数据量，第一个包的index是0，第二个包是188，第三个包是188+187，也就是188*2，看循环的条件，要求data至少要有188*2个字节*/
	for i := 0; i+188*2 <= len(data)-1; i++ {
		/*这是要连续读取三个ts包么？*/
		if data[i] == 0x47 && data[i+188] == 0x47 && data[i+188*2] == 0x47 {
			if (data[i+1]&0x5F) == 0x40 && data[i+2] == 0x00 {
				return int64(i), nil
			}
		}
	}
	return 0, fmt.Errorf("Cannot find pat")
}
