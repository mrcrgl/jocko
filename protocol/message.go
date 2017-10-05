package protocol

import "time"

type Message struct {
	Timestamp time.Time
	Key       []byte
	Value     []byte
	MagicByte int8
}

func (m *Message) Encode(e PacketEncoder, version int16) error {
	e.Push(&CRCField{})
	e.PutInt8(m.MagicByte)
	e.PutInt8(0) // attributes
	if version == 1 && m.MagicByte > 0 {
		e.PutInt64(m.Timestamp.UnixNano() / int64(time.Millisecond))
	}
	if err := e.PutBytes(m.Key); err != nil {
		return err
	}
	if err := e.PutBytes(m.Value); err != nil {
		return err
	}
	e.Pop()
	return nil
}

func (m *Message) Decode(d PacketDecoder, version int16) error {
	var err error
	if err = d.Push(&CRCField{}); err != nil {
		return err
	}
	if m.MagicByte, err = d.Int8(); err != nil {
		return err
	}
	if _, err := d.Int8(); err != nil {
		return err
	}
	if version == 1 && m.MagicByte > 0 {
		t, err := d.Int64()
		if err != nil {
			return err
		}
		m.Timestamp = time.Unix(t/1000, (t%1000)*int64(time.Millisecond))
	}
	if m.Key, err = d.Bytes(); err != nil {
		return err
	}
	if m.Value, err = d.Bytes(); err != nil {
		return err
	}
	return d.Pop()
}
