package events

type PushData struct {
	Image  string
	Tag    string
	Digest string
}

type PayloadHandler interface {
	Validate([]byte, string) error
	Decode([]byte, string) (PushData, error)
}
