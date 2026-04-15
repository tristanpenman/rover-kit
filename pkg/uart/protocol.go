package uart

import (
	"encoding/binary"
	"errors"
)

const (
	StartByte0 byte = 0xAA
	StartByte1 byte = 0x55

	Version1 byte = 0x01

	DistanceUnitMillimeters byte = 0x00
	DistanceUnitCentimeters byte = 0x01

	HeaderLength = 4
	CRCSize      = 2
)

var (
	ErrFrameTooShort    = errors.New("frame too short")
	ErrInvalidStart     = errors.New("invalid start bytes")
	ErrInvalidCRC       = errors.New("invalid crc16")
	ErrUnsupportedVer   = errors.New("unsupported protocol version")
	ErrInvalidPayloadV1 = errors.New("invalid v1 payload")
)

type SampleV1 struct {
	TimestampMS  uint32
	DistanceUnit byte
	Readings     []uint16
}

func (s SampleV1) MarshalPayload() []byte {
	payload := make([]byte, 6+len(s.Readings)*2)
	binary.LittleEndian.PutUint32(payload[0:4], s.TimestampMS)
	payload[4] = s.DistanceUnit
	payload[5] = byte(len(s.Readings))

	idx := 6
	for _, reading := range s.Readings {
		binary.LittleEndian.PutUint16(payload[idx:idx+2], reading)
		idx += 2
	}
	return payload
}

func ParsePayloadV1(payload []byte) (SampleV1, error) {
	if len(payload) < 6 {
		return SampleV1{}, ErrInvalidPayloadV1
	}

	sensorCount := int(payload[5])
	expected := 6 + sensorCount*2
	if len(payload) != expected {
		return SampleV1{}, ErrInvalidPayloadV1
	}

	out := SampleV1{
		TimestampMS:  binary.LittleEndian.Uint32(payload[0:4]),
		DistanceUnit: payload[4],
		Readings:     make([]uint16, sensorCount),
	}

	idx := 6
	for i := range out.Readings {
		out.Readings[i] = binary.LittleEndian.Uint16(payload[idx : idx+2])
		idx += 2
	}
	return out, nil
}

func EncodeFrame(version byte, payload []byte) []byte {
	frame := make([]byte, HeaderLength+len(payload)+CRCSize)
	frame[0] = StartByte0
	frame[1] = StartByte1
	frame[2] = version
	frame[3] = byte(len(payload))
	copy(frame[4:], payload)

	crc := CRC16CCITTFALSE(frame[2 : 4+len(payload)])
	binary.LittleEndian.PutUint16(frame[4+len(payload):], crc)
	return frame
}

func DecodeFrame(frame []byte) (version byte, payload []byte, err error) {
	if len(frame) < HeaderLength+CRCSize {
		return 0, nil, ErrFrameTooShort
	}
	if frame[0] != StartByte0 || frame[1] != StartByte1 {
		return 0, nil, ErrInvalidStart
	}

	payloadLen := int(frame[3])
	expected := HeaderLength + payloadLen + CRCSize
	if len(frame) != expected {
		return 0, nil, ErrFrameTooShort
	}

	crcWant := binary.LittleEndian.Uint16(frame[4+payloadLen:])
	crcHave := CRC16CCITTFALSE(frame[2 : 4+payloadLen])
	if crcWant != crcHave {
		return 0, nil, ErrInvalidCRC
	}

	version = frame[2]
	payload = frame[4 : 4+payloadLen]
	return version, payload, nil
}

type Decoder struct {
	buf        []byte
	MaxPayload int
}

func NewDecoder(maxPayload int) *Decoder {
	if maxPayload <= 0 {
		maxPayload = 64
	}
	return &Decoder{MaxPayload: maxPayload}
}

func (d *Decoder) Push(data []byte) {
	d.buf = append(d.buf, data...)
}

func (d *Decoder) NextFrame() ([]byte, bool) {
	for {
		if len(d.buf) < HeaderLength+CRCSize {
			return nil, false
		}

		sync := -1
		for i := 0; i+1 < len(d.buf); i++ {
			if d.buf[i] == StartByte0 && d.buf[i+1] == StartByte1 {
				sync = i
				break
			}
		}
		if sync < 0 {
			d.buf = d.buf[:0]
			return nil, false
		}
		if sync > 0 {
			d.buf = d.buf[sync:]
		}
		if len(d.buf) < HeaderLength+CRCSize {
			return nil, false
		}

		payloadLen := int(d.buf[3])
		if payloadLen > d.MaxPayload {
			d.buf = d.buf[1:]
			continue
		}

		total := HeaderLength + payloadLen + CRCSize
		if len(d.buf) < total {
			return nil, false
		}

		candidate := make([]byte, total)
		copy(candidate, d.buf[:total])
		d.buf = d.buf[total:]
		if _, _, err := DecodeFrame(candidate); err != nil {
			d.buf = append(candidate[1:], d.buf...)
			continue
		}

		return candidate, true
	}
}

func CRC16CCITTFALSE(data []byte) uint16 {
	crc := uint16(0xFFFF)
	for _, b := range data {
		crc ^= uint16(b) << 8
		for i := 0; i < 8; i++ {
			if crc&0x8000 != 0 {
				crc = (crc << 1) ^ 0x1021
			} else {
				crc <<= 1
			}
		}
	}
	return crc
}
