package adoptiontransport

import (
	"fmt"

	jsonschema "github.com/tylergannon/go-gen-jsonschema"
)

type ShipmentEvent interface{ shipmentEvent() }

type Created struct {
	Actor string `json:"actor"`
	At    string `json:"at"`
}

func (Created) shipmentEvent() {}

type Dispatched struct {
	Carrier  string `json:"carrier"`
	Tracking string `json:"tracking"`
}

func (Dispatched) shipmentEvent() {}

type Status int

const (
	StatusUnknown Status = iota
	StatusReady
	StatusDelivered
)

// String is intentionally display-oriented. The registered constant names,
// rather than this text, are the string-mode enum wire values.
func (status Status) String() string { return fmt.Sprintf("display-status-%d", status) }

type Shipment struct {
	ID       string                      `json:"id"`
	Status   Status                      `json:"status"`
	Event    ShipmentEvent               `json:"event"`
	Note     jsonschema.Optional[string] `json:"note,omitzero"`
	ETA      jsonschema.Nullable[string] `json:"eta"`
	Quantity int                         `json:"quantity"`
}
