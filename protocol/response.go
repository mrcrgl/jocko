package protocol

type ResponseBody interface {
	Encoder
	Decoder
}

type Response struct {
	Size          int32
	CorrelationID int32
	Body          ResponseBody
}

func (r *Response) Encode(pe PacketEncoder, version int16) (err error) {
	pe.Push(&SizeField{})
	pe.PutInt32(r.CorrelationID)
	if err != nil {
		return err
	}
	err = r.Body.Encode(pe, version)
	if err != nil {
		return err
	}
	pe.Pop()
	return nil
}

func (r *Response) Decode(pd PacketDecoder, version int16) (err error) {
	r.Size, err = pd.Int32()
	if err != nil {
		return err
	}
	r.CorrelationID, err = pd.Int32()
	if r.Body != nil {
		return r.Body.Decode(pd, version)
	}
	return err
}
