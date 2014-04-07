package univedo

import (
	"bytes"
	"encoding/binary"
	"errors"
	"github.com/nu7hatch/gouuid"
	"time"
)

const (
	majorUInt       = 0
	majorInt        = 1
	majorByteString = 2
	majorTextString = 3
	majorArray      = 4
	majorMap        = 5
	majorTag        = 6
	majorSimple     = 7
)

const (
	tagDateTime     = 0
	tagDecimal      = 4
	tagRemoteObject = 6
	tagUuid         = 7
	tagRecord       = 8
)

const (
	simpleFalse   = 20
	simpleTrue    = 21
	simpleNull    = 22
	simpleFloat32 = 26
	simpleFloat64 = 27
)

type message struct {
	buffer *bytes.Buffer
	offset int
}

func (m *message) getLen(typeByte byte) (uint64, error) {
	smallLen := uint8(typeByte & 0x1F)
	switch smallLen {
	case 24:
		var i uint8
		err := binary.Read(m.buffer, binary.BigEndian, &i)
		return uint64(i), err
	case 25:
		var i uint16
		err := binary.Read(m.buffer, binary.BigEndian, &i)
		return uint64(i), err
	case 26:
		var i uint32
		err := binary.Read(m.buffer, binary.BigEndian, &i)
		return uint64(i), err
	case 27:
		var i uint64
		err := binary.Read(m.buffer, binary.BigEndian, &i)
		return i, err
	default:
		return uint64(smallLen), nil
	}
}

func (m *message) readByteString(typeByte byte) ([]byte, error) {
	len, err := m.getLen(typeByte)
	if err != nil {
		return nil, err
	}
	b := make([]byte, len)
	n, err := m.buffer.Read(b)
	if err != nil {
		return nil, err
	}
	if uint64(n) != len {
		return nil, errors.New("unexpected end of buffer in cbor protocol")
	}
	return b, nil
}

func (m *message) read() (interface{}, error) {
	typeByte, err := m.buffer.ReadByte()
	if err != nil {
		return nil, err
	}

	switch typeByte >> 5 {

	case majorUInt:
		return m.getLen(typeByte)

	case majorInt:
		r, err := m.getLen(typeByte)
		if err != nil {
			return nil, err
		}
		if r > 9223372036854775807 {
			return nil, errors.New("unrepresentable integer in cbor protocol")
		}
		return -int64(r) - 1, nil

	case majorByteString:
		return m.readByteString(typeByte)

	case majorTextString:
		b, err := m.readByteString(typeByte)
		if err != nil {
			return nil, err
		}
		return string(b), nil

	case majorArray:
		len, err := m.getLen(typeByte)
		if err != nil {
			return nil, err
		}
		arr := make([]interface{}, len)
		var i uint64
		for i = 0; i < len; i++ {
			r, err := m.read()
			if err != nil {
				return nil, errors.New("error while receiving array in cbor protocol")
			}
			arr[i] = r
		}
		return arr, nil

	case majorMap:
		len, err := m.getLen(typeByte)
		if err != nil {
			return nil, err
		}
		res := make(map[string]interface{}, len)
		var i uint64
		for i = 0; i < len; i++ {
			key, err := m.read()
			if err != nil {
				return nil, errors.New("error while receiving map in cbor protocol")
			}
			val, err := m.read()
			if err != nil {
				return nil, errors.New("error while receiving map in cbor protocol")
			}
			keyString, ok := key.(string)
			if !ok {
				return nil, errors.New("expected string as map key in cbor protocol")
			}
			res[keyString] = val
		}
		return res, nil

	case majorTag:
		tag, err := m.getLen(typeByte)
		if err != nil {
			return nil, err
		}
		switch tag {

		case tagDateTime:
			data, err := m.read()
			if err != nil {
				return nil, err
			}
			dataStr, ok := data.(string)
			if !ok {
				return nil, errors.New("expected string for datetime in cbor protocol")
			}
			d, err := time.Parse(time.RFC3339Nano, dataStr)
			if err != nil {
				return nil, err
			}
			return d, nil

		case tagUuid:
			data, err := m.read()
			if err != nil {
				return nil, err
			}
			dataSlice, ok := data.([]byte)
			if !ok {
				return nil, errors.New("expected bytestring in cbor protocol")
			}
			return uuid.Parse(dataSlice)

		case tagRecord:
			return m.read()

		default:
			return nil, errors.New("invalid tag in cbor protocol")
		}

	case majorSimple:
		switch typeByte & 0x1F {

		case simpleFalse:
			return false, nil

		case simpleTrue:
			return true, nil

		case simpleNull:
			return nil, nil

		case simpleFloat32:
			var r float32
			err := binary.Read(m.buffer, binary.BigEndian, &r)
			if err != nil {
				return nil, err
			}
			return r, nil

		case simpleFloat64:
			var r float64
			err := binary.Read(m.buffer, binary.BigEndian, &r)
			if err != nil {
				return nil, err
			}
			return r, nil

		default:
			return nil, errors.New("invalid simple in cbor protocol")
		}

	default:
		return nil, errors.New("invalid major in cbor protocol")
	}
}

func (m *message) sendSimple(s byte) {
	m.buffer.WriteByte(majorSimple<<5 | s)
}

func (m *message) sendTag(t byte) {
	m.buffer.WriteByte(majorTag<<5 | t)
}

func (m *message) sendLen(typeByte byte, l uint64) {
	typeByte = typeByte << 5
	switch {
	case l <= 23:
		m.buffer.WriteByte(typeByte | byte(l))
	case l < 0x100:
		m.buffer.WriteByte(typeByte | 24)
		m.buffer.WriteByte(uint8(l))
	case l < 0x10000:
		m.buffer.WriteByte(typeByte | 25)
		binary.Write(m.buffer, binary.BigEndian, uint16(l))
	case l < 0x100000000:
		m.buffer.WriteByte(typeByte | 26)
		binary.Write(m.buffer, binary.BigEndian, uint32(l))
	default:
		m.buffer.WriteByte(typeByte | 27)
		binary.Write(m.buffer, binary.BigEndian, uint64(l))
	}
}

func (m *message) send(pObj interface{}) error {
	switch obj := pObj.(type) {

	case nil:
		m.sendSimple(simpleNull)

	case bool:
		if obj == true {
			m.sendSimple(simpleTrue)
		} else {
			m.sendSimple(simpleFalse)
		}

	case uint8:
		m.sendLen(majorUInt, uint64(obj))
	case uint16:
		m.sendLen(majorUInt, uint64(obj))
	case uint32:
		m.sendLen(majorUInt, uint64(obj))
	case uint64:
		m.sendLen(majorUInt, obj)

	case int8:
		if obj >= 0 {
			m.send(uint64(obj))
		} else {
			m.sendLen(majorInt, uint64(-obj-1))
		}
	case int16:
		if obj >= 0 {
			m.send(uint64(obj))
		} else {
			m.sendLen(majorInt, uint64(-obj-1))
		}
	case int32:
		if obj >= 0 {
			m.send(uint64(obj))
		} else {
			m.sendLen(majorInt, uint64(-obj-1))
		}
	case int64:
		if obj >= 0 {
			m.send(uint64(obj))
		} else {
			m.sendLen(majorInt, uint64(-obj-1))
		}

	case int:
		if obj >= 0 {
			m.send(uint64(obj))
		} else {
			m.sendLen(majorInt, uint64(-obj-1))
		}

	case float32:
		m.sendSimple(simpleFloat32)
		return binary.Write(m.buffer, binary.BigEndian, &obj)
	case float64:
		m.sendSimple(simpleFloat64)
		return binary.Write(m.buffer, binary.BigEndian, &obj)

	case string:
		m.sendLen(majorTextString, uint64(len(obj)))
		return binary.Write(m.buffer, binary.BigEndian, []byte(obj))

	case []byte:
		m.sendLen(majorByteString, uint64(len(obj)))
		return binary.Write(m.buffer, binary.BigEndian, obj)

	case []interface{}:
		m.sendLen(majorArray, uint64(len(obj)))
		for _, v := range obj {
			err := m.send(v)
			if err != nil {
				return err
			}
		}

	case map[string]interface{}:
		m.sendLen(majorMap, uint64(len(obj)))
		for k, v := range obj {
			err := m.send(k)
			if err != nil {
				return err
			}
			err = m.send(v)
			if err != nil {
				return err
			}
		}

	case *uuid.UUID:
		m.sendTag(tagUuid)
		d := [16]byte(*obj)
		m.send(d[:])

	case time.Time:
		m.sendTag(tagDateTime)
		m.send(obj.Format(time.RFC3339Nano))

	default:
		return errors.New("cannot send object in cbor protocol")
	}

	return nil
}
