package json

type Event interface {
	event()
}

type Created struct {
	Name string `json:"name"`
}

func (Created) event() {}
