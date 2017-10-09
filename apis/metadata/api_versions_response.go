package metadata

import "github.com/travisjeffery/jocko/protocol"

type APIVersionsResponse struct {
	APIVersions []APIVersion
	ErrorCode   int16
}

type APIVersion struct {
	APIKey     int16
	MinVersion int16
	MaxVersion int16
}

func (c *APIVersionsResponse) Encode(e protocol.PacketEncoder, _ int16) error {
	e.PutInt16(c.ErrorCode)

	if err := e.PutArrayLength(len(c.APIVersions)); err != nil {
		return err
	}
	for _, av := range c.APIVersions {
		e.PutInt16(av.APIKey)
		e.PutInt16(av.MinVersion)
		e.PutInt16(av.MaxVersion)
	}
	return nil
}

func (c *APIVersionsResponse) Decode(d protocol.PacketDecoder, _ int16) error {
	l, err := d.ArrayLength()
	if err != nil {
		return err
	}
	c.APIVersions = make([]APIVersion, l)
	for i := range c.APIVersions {
		key, err := d.Int16()
		if err != nil {
			return err
		}

		minVersion, err := d.Int16()
		if err != nil {
			return err
		}

		maxVersion, err := d.Int16()
		if err != nil {
			return err
		}

		c.APIVersions[i] = APIVersion{
			APIKey:     key,
			MinVersion: minVersion,
			MaxVersion: maxVersion,
		}
	}
	return nil
}