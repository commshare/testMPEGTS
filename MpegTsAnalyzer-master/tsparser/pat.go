package tsparser

import (
	"fmt"

	"../bitbuffer"
)
/*TS流中会定期出现PAT表。PAT表提供了节目号和对应PMT表格的PID的对应关系。*/
// Pat Program Map Table.
type Pat struct {
	startFlag         bool
	continuityCounter uint8
	buf               []byte
	pmtPid            uint16

	tableID                uint8 /*table_id：固定为0x00，标志该表是PAT表。*/
	sectionSyntaxIndicator uint8 /*1 bit*/
	sectionLength          uint16 /*12 bit  表示这个字节后面有用的字节数，包括CRC32。节目套数：（section length-9）/4*/
	transportStreamID      uint16  /*16 program number 16位字段，表示该TS流的ID，区别于同一个网络中其它多路复用流。*/
	versionNumber          uint8 /*5 bit表示PAT的版本号。 */
	currentNextIndicator   uint8 /*1 bit 表示发送的PAT表是当前有效还是下一个PAT有效。*/
	sectionNumber          uint8 /*8  表示分段的号码。PAT可能分为多段传输，第一段为0，以后每个分段加1，最多可能有256个分段。*/
	lastSectionNumber      uint8 /*8 表示PAT最后一个分段的号码。*/
	programInfo            []PatProgramInfo
	crc32                  uint32
}

// PatProgramInfo Program Info of mpeg.
type PatProgramInfo struct {
	programNumber uint16 /*Program number：节目号*/
	networkPid    uint16 /*network_PID：网络信息表（NIT）的PID,节目号为0时对应ID为network_PID。*/
	programMapPid uint16 /*Program map PID：节目映射表（PMT）的PID号，节目号为大于等于1时，对应的ID为program_map_PID。一个PAT中可以有多个program_map_PID。*/
}

// NewPat create new PAT instance
func NewPat() *Pat { return new(Pat) }

// ContinuityCounter return current continuity_counter of TsPacket.
func (p *Pat) ContinuityCounter() uint8 { return p.continuityCounter }

// SetContinuityCounter set current continuity_counter of TsPacket.
func (p *Pat) SetContinuityCounter(continuityCounter uint8) { p.continuityCounter = continuityCounter }

// PmtPid return PMT pid.
func (p *Pat) PmtPid() uint16 { return p.pmtPid }

// Append append ts payload data for buffer.
func (p *Pat) Append(buf []byte) {
	p.buf = append(p.buf, buf...)
}

// Parse PAT data.
func (p *Pat) Parse() error {
	bb := new(bitbuffer.BitBuffer)
	bb.Set(p.buf)

	var err error
	if p.tableID, err = bb.PeekUint8(8); err != nil {
		return err
	}
	if p.sectionSyntaxIndicator, err = bb.PeekUint8(1); err != nil {
		return err
	}
	if err = bb.Skip(1); err != nil {
		return err
	} // ()
	if err = bb.Skip(2); err != nil {
		return err
	} // reserved
	/*sectionlenght表示这个字节后面有用的字节数*/
	if p.sectionLength, err = bb.PeekUint16(12); err != nil {
		return err
	}
	if p.transportStreamID, err = bb.PeekUint16(16); err != nil {
		return err
	}
	if err = bb.Skip(2); err != nil {
		return err
	} // reserved
	if p.versionNumber, err = bb.PeekUint8(5); err != nil {
		return err
	}
	if p.currentNextIndicator, err = bb.PeekUint8(1); err != nil {
		return err
	}
	if p.sectionNumber, err = bb.PeekUint8(8); err != nil {
		return err
	}
	if p.lastSectionNumber, err = bb.PeekUint8(8); err != nil {
		return err
	}
	/*下面是n loop ，节目套数： ((int(p.sectionLength) - 9) / 4)*/
	for i := 0; i < ((int(p.sectionLength) - 9) / 4); i++ {
		var patProgramInfo PatProgramInfo
		/*节目号*/
		if patProgramInfo.programNumber, err = bb.PeekUint16(16); err != nil {
			return err
		}
		if err = bb.Skip(3); err != nil {
			return err
		} // reserved
		if patProgramInfo.programNumber == 0 {
			/*如果节目号为0，那么这个是网络信息表NIT的pid*/
			if patProgramInfo.networkPid, err = bb.PeekUint16(13); err != nil {
				return err
			}
		} else {
			/*节目号大于等于1时，对应的id为 program_map_PID ,一个PAT中，可以有多个program_map_PID*/
			if patProgramInfo.programMapPid, err = bb.PeekUint16(13); err != nil {
				return err
			}
			p.pmtPid = patProgramInfo.programMapPid
		}
		/*slice*/
		p.programInfo = append(p.programInfo, patProgramInfo)
	}
	/*32位的字段，crc32校验码*/
	if p.crc32, err = bb.PeekUint32(32); err != nil {
		return err
	}
	/*TODO 这个是*/
	if len(p.buf) >= int(3+p.sectionLength-4) && p.crc32 != crc32(p.buf[0:3+p.sectionLength-4]) {
		return fmt.Errorf("PAT CRC32 is invalidate")
	}

	return nil
}

// Dump PAT detail.
func (p *Pat) Dump() {
	fmt.Printf("\n===========================================\n")
	fmt.Printf(" PAT")
	fmt.Printf("\n===========================================\n")
	fmt.Printf("PAT : table_id			: 0x%x\n", p.tableID)
	fmt.Printf("PAT : section_syntax_indicator	: %d\n", p.sectionSyntaxIndicator)
	fmt.Printf("PAT : section_length		: %d\n", p.sectionLength)
	fmt.Printf("PAT : transport_stream_id	: %d\n", p.transportStreamID)
	fmt.Printf("PAT : version_number		: %d\n", p.versionNumber)
	fmt.Printf("PAT : current_next_indicator	: %d\n", p.currentNextIndicator)
	fmt.Printf("PAT : section_number		: %d\n", p.sectionNumber)
	fmt.Printf("PAT : last_section_number	: %d\n", p.lastSectionNumber)

	for _, val := range p.programInfo {
		fmt.Printf("PAT : program_number		: %d\n", val.programNumber)
		if val.programNumber == 0 {
			fmt.Printf("PAT : network_PID		: 0x%x\n", val.networkPid)
		} else {
			fmt.Printf("PAT : program_map_PID		: 0x%x\n", val.programMapPid)
		}
	}
	fmt.Printf("PAT : CRC_32			: %x\n", p.crc32)
}
