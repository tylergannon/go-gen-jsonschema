package broken

import "example.com/missing"

type Broken struct {
	Value missing.Value `json:"value"`
}
