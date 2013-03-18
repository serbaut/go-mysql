package mysql

import (
	"bytes"
	"database/sql/driver"
	"fmt"
	"io"
	"math"
	"strings"
	"time"
)

type packet struct {
	bytes.Buffer
}

func newPacket(seq int) (p packet) {
	p.Write([]byte{0, 0, 0, byte(seq)})
	return p
}

func readSize(r io.Reader) (int, error) {
	var h [4]byte
	_, err := io.ReadFull(r, h[:])
	if err != nil {
		return 0, err
	}
	size := int(h[0]) + int(h[1])<<8 + int(h[2])<<16
	if size == 0 || size > MAX_PACKET_SIZE {
		return 0, fmt.Errorf("invalid packet size: %d", size)
	}
	return size, nil
}

func (p *packet) ReadFrom(r io.Reader) (int64, error) {
	size, err := readSize(r)
	if err != nil {
		return 0, err
	}
	buf := make([]byte, size)
	if _, err = io.ReadFull(r, buf); err != nil {
		return 0, err
	}
	for size == MAX_PACKET_SIZE {
		if size, err = readSize(r); err != nil {
			return 0, err
		}
		m := len(buf)
		n := m + int(size)
		if n > cap(buf) {
			t := make([]byte, (n+1)*2)
			copy(t, buf)
			buf = t
		}
		buf = buf[:n]
		if _, err = io.ReadFull(r, buf[m:n]); err != nil {
			return 0, err
		}
	}
	p.Buffer = *bytes.NewBuffer(buf)
	return int64(p.Len()), nil
}

func (p *packet) FirstByte() (v uint8) {
	return p.Bytes()[0]
}

func (p *packet) ReadUint8() (v uint8) {
	v, _ = p.ReadByte()
	return v
}

func (p *packet) ReadUint16() (v uint16) {
	q := p.Next(2)
	return uint16(q[0]) | uint16(q[1])<<8
}

func (p *packet) ReadUint24() (v uint32) {
	q := p.Next(3)
	return uint32(q[0]) | uint32(q[1])<<8 | uint32(q[2])<<16
}

func (p *packet) ReadUint32() (v uint32) {
	q := p.Next(4)
	return uint32(q[0]) | uint32(q[1])<<8 | uint32(q[2])<<16 | uint32(q[3])<<24
}

func (p *packet) ReadUint64() (v uint64) {
	q := p.Next(8)
	return uint64(q[0]) | uint64(q[1])<<8 | uint64(q[2])<<16 | uint64(q[3])<<24 | uint64(q[4])<<32 | uint64(q[5])<<40 | uint64(q[6])<<48 | uint64(q[7])<<56
}

func (p *packet) ReadLCUint64() (v uint64, isnull bool) {
	switch x := p.ReadUint8(); x {
	case 0xfb:
		isnull = true
	case 0xfc:
		v = uint64(p.ReadUint16())
	case 0xfd:
		v = uint64(p.ReadUint24())
	case 0xfe:
		v = p.ReadUint64()
	default:
		v = uint64(x)
	}
	return v, isnull
}

func (p *packet) ReadLCBytes() (v []byte, isnull bool) {
	if n, isnull := p.ReadLCUint64(); !isnull {
		return p.Next(int(n)), false
	}
	return nil, true
}

func (p *packet) ReadLCString() (v string, isnull bool) {
	if n, isnull := p.ReadLCUint64(); !isnull {
		return string(p.Next(int(n))), false
	}
	return "", true
}

func (p *packet) ReadErr() error {
	p.ReadByte()
	errorCode := p.ReadUint16()
	p.ReadByte()
	state := string(p.Next(5))
	info := string(p.Bytes())
	return fmt.Errorf("ERROR %d (%s): %s", errorCode, state, info)
}

func (p *packet) ReadEOF() (warnings, status uint16) {
	p.ReadByte()
	warnings = p.ReadUint16()
	status = p.ReadUint16()
	return
}

func (p *packet) ReadOK() (rowsAffected, lastInsertId int64, warnings uint16) {
	p.ReadByte()
	rows, _ := p.ReadLCUint64()
	last, _ := p.ReadLCUint64()
	p.ReadUint16() // ignore status
	warnings = p.ReadUint16()
	return int64(rows), int64(last), warnings
}

func (p *packet) SkipLCBytes() {
	n, _ := p.ReadLCUint64()
	p.Next(int(n))
}

func (p *packet) WriteTo(w io.Writer) (n int64, err error) {
	buf := p.Bytes()
	size := len(buf) - 4
	buf[0] = byte(size)
	buf[1] = byte(size >> 8)
	buf[2] = byte(size >> 16)
	nn, err := w.Write(buf)
	return int64(nn), err
}

func (p *packet) WriteUint16(v uint16) {
	p.Write([]byte{byte(v), byte(v >> 8)})
}

func (p *packet) WriteUint24(v uint32) {
	p.Write([]byte{byte(v), byte(v >> 8), byte(v >> 16)})
}

func (p *packet) WriteUint32(v uint32) {
	p.Write([]byte{byte(v), byte(v >> 8), byte(v >> 16), byte(v >> 24)})
}

func (p *packet) WriteUint64(v uint64) {
	p.Write([]byte{byte(v), byte(v >> 8), byte(v >> 16), byte(v >> 24), byte(v >> 32), byte(v >> 40), byte(v >> 48), byte(v >> 56)})
}

func (p *packet) WriteLCUint64(v uint64) {
	switch {
	case v < 251:
		p.WriteByte(byte(v))
	case v < 65536:
		p.WriteByte(0xfc)
		p.WriteUint16(uint16(v))
	case v < 16777216:
		p.WriteByte(0xfd)
		p.WriteUint24(uint32(v))
	default:
		p.WriteByte(0xfe)
		p.WriteUint64(v)
	}
}

func (p *packet) ReadMask(nbits int) (mask []bool) {
	bytes := p.Next(int((nbits + 7) / 8))
	mask = make([]bool, nbits)
	for i := range mask {
		mask[i] = (bytes[i/8]>>byte(i%8))&1 > 0
	}
	return mask
}

func (p *packet) WriteMask(mask []bool) {
	buf := make([]byte, (len(mask)+7)/8)
	for i := range mask {
		if mask[i] {
			buf[i/8] |= (1 << byte(i&7))
		}
	}
	p.Write(buf)
}

func (p *packet) WriteArgs(args []driver.Value) error {
	v := packet{}
	for i := range args {
		switch t := args[i].(type) {
		case nil:
			p.WriteUint16(MYSQL_TYPE_NULL)
		case int:
			p.WriteUint16(MYSQL_TYPE_LONG)
			v.WriteUint32(uint32(t))
		case int32:
			p.WriteUint16(MYSQL_TYPE_LONG)
			v.WriteUint32(uint32(t))
		case int64:
			p.WriteUint16(MYSQL_TYPE_LONGLONG)
			v.WriteUint64(uint64(t))
		case float32:
			p.WriteUint16(MYSQL_TYPE_FLOAT)
			v.WriteUint32(math.Float32bits(t))
		case float64:
			p.WriteUint16(MYSQL_TYPE_DOUBLE)
			v.WriteUint64(math.Float64bits(t))
		case bool:
			p.WriteUint16(MYSQL_TYPE_TINY)
			if t {
				v.WriteByte(1)
			} else {
				v.WriteByte(0)
			}
		case string:
			p.WriteUint16(MYSQL_TYPE_STRING)
			if len(t) <= MAX_DATA_CHUNK {
				v.WriteLCUint64(uint64(len(t)))
				v.WriteString(t)
			}
		case []byte:
			p.WriteUint16(MYSQL_TYPE_BLOB)
			if len(t) <= MAX_DATA_CHUNK {
				v.WriteLCUint64(uint64(len(t)))
				v.Write(t)
			}
		case time.Time:
			t = t.UTC()
			p.WriteUint16(MYSQL_TYPE_DATETIME)
			v.WriteByte(7)
			v.WriteUint16(uint16(t.Year()))
			v.WriteByte(byte(t.Month()))
			v.WriteByte(byte(t.Day()))
			v.WriteByte(byte(t.Hour()))
			v.WriteByte(byte(t.Minute()))
			v.WriteByte(byte(t.Second()))
		case time.Duration:
			p.WriteUint16(MYSQL_TYPE_TIME)
			v.WriteByte(8)
			s, neg := t/1000000000, 0
			if s < 0 {
				s, neg = -s, 1
			}
			ss, s := s%60, s/60
			mm, s := s%60, s/60
			hh, s := s%24, s/24
			days := uint32(s)
			v.WriteByte(byte(neg))
			v.WriteUint32(days)
			v.WriteByte(byte(hh))
			v.WriteByte(byte(mm))
			v.WriteByte(byte(ss))
		default:
			return fmt.Errorf("invalid parameter: %v", args[i])
		}
	}
	_, err := p.Write(v.Bytes())
	return err
}

func (p *packet) ReadValue(coltype byte, flags uint16, isnull bool) (v interface{}, err error) {
	if isnull {
		switch coltype {
		case MYSQL_TYPE_TIMESTAMP, MYSQL_TYPE_DATETIME, MYSQL_TYPE_DATE, MYSQL_TYPE_NEWDATE:
			return time.Time{}, nil
		case MYSQL_TYPE_TIME:
			return time.Duration(0), nil
		default:
			return nil, nil
		}
	}

	switch coltype {
	case MYSQL_TYPE_TINY:
		if flags&UNSIGNED_FLAG == 0 {
			v = int8(p.ReadUint8())
		} else {
			v = p.ReadUint8()
		}

	case MYSQL_TYPE_SHORT, MYSQL_TYPE_YEAR:
		if flags&UNSIGNED_FLAG == 0 {
			v = int16(p.ReadUint16())
		} else {
			v = p.ReadUint16()
		}

	case MYSQL_TYPE_LONG:
		if flags&UNSIGNED_FLAG == 0 {
			v = int32(p.ReadUint32())
		} else {
			v = p.ReadUint32()
		}

	case MYSQL_TYPE_INT24:
		if flags&UNSIGNED_FLAG == 0 {
			v = int32(p.ReadUint24()<<8) >> 8 // sign extend to 32 bit
		} else {
			v = p.ReadUint24()
		}

	case MYSQL_TYPE_LONGLONG:
		if flags&UNSIGNED_FLAG == 0 {
			v = int64(p.ReadUint64())
		} else {
			v = p.ReadUint64()
		}

	case MYSQL_TYPE_FLOAT:
		v = math.Float32frombits(p.ReadUint32())

	case MYSQL_TYPE_DOUBLE:
		v = math.Float64frombits(p.ReadUint64())

	case MYSQL_TYPE_TIMESTAMP, MYSQL_TYPE_DATETIME, MYSQL_TYPE_DATE, MYSQL_TYPE_NEWDATE:
		if n := p.ReadUint8(); n > 0 {
			y := int(p.ReadUint16())
			m := time.Month(p.ReadUint8())
			d := int(p.ReadUint8())
			var hh, mm, ss, ns int
			if n > 4 {
				hh = int(p.ReadUint8())
				mm = int(p.ReadUint8())
				ss = int(p.ReadUint8())

				if n > 7 {
					ns = int(p.ReadUint32()) * 1000
				}
			}
			v = time.Date(y, m, d, hh, mm, ss, ns, time.UTC)
		} else {
			v = time.Time{}
		}

	case MYSQL_TYPE_TIME:
		if n := p.ReadUint8(); n > 0 {
			neg := p.ReadUint8()
			days := int64(p.ReadUint32())
			hh := int64(p.ReadUint8())
			mm := int64(p.ReadUint8())
			ss := int64(p.ReadUint8())
			var ns int64
			if n > 8 {
				ns = int64(p.ReadUint32()) * 1000
			}

			ns += (((days*24+hh)*60+mm)*60 + ss) * 1000000000
			if neg == 1 {
				ns = -ns
			}
			v = time.Duration(ns)
		} else {
			v = time.Duration(0)
		}

	case MYSQL_TYPE_STRING, MYSQL_TYPE_VARCHAR, MYSQL_TYPE_VAR_STRING,
		MYSQL_TYPE_BLOB, MYSQL_TYPE_LONG_BLOB, MYSQL_TYPE_MEDIUM_BLOB,
		MYSQL_TYPE_DECIMAL, MYSQL_TYPE_NEWDECIMAL, MYSQL_TYPE_BIT:
		if s, isnull := p.ReadLCBytes(); !isnull {
			v = s
		}

	default:
		return nil, fmt.Errorf("unkown colymn type: %d", coltype)
	}

	return v, err
}

func (p *packet) ReadTextValue(coltype byte, flags uint16) (v interface{}, err error) {
	b, isnull := p.ReadLCBytes()

	switch coltype {
	case MYSQL_TYPE_DATETIME, MYSQL_TYPE_TIMESTAMP:
		if isnull || bytes.Equal(b, []byte("0000-00-00 00:00:00")) {
			return time.Time{}, nil
		} else {
			return time.Parse("2006-01-02 15:04:05", string(b))
		}
	case MYSQL_TYPE_DATE:
		if isnull || bytes.Equal(b, []byte("0000-00-00")) {
			return time.Time{}, nil
		} else {
			return time.Parse("2006-01-02", string(b))
		}
	case MYSQL_TYPE_TIME:
		if isnull {
			return time.Duration(0), nil
		} else {
			t := strings.Split(string(b), ":")
			if len(t) != 3 {
				return nil, fmt.Errorf("invalid time: %s", b)
			}
			return time.ParseDuration(fmt.Sprintf("%sh%sm%ss", t[0], t[1], t[2]))
		}
	default:
		if isnull {
			return nil, nil
		} else {
			return b, nil
		}
	}
	panic("unreachable")
}
