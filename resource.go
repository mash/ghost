package ghost

type Resource interface {
	PKeys() []PKey
	SetPKeys([]PKey)
}
