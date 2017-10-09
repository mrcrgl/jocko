package protocol

type DecoderConstructor interface {
	New() Decoder
}

type EncoderConstructor interface {
	New() Encoder
}

type APIRequest interface {
	Decoder
	DecoderConstructor
}

type APIResponse interface {
	Encoder
	EncoderConstructor
}
