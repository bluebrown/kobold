package events

type PushData struct {
	Image  string
	Tag    string
	Digest string
}

type PayloadHandler interface {
	Validate([]byte) error
	Decode([]byte) ([]PushData, error)
}
