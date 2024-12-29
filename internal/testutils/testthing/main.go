package main

import (
	"encoding/json"
	"fmt"
)

type Embedded struct {
	Foo  string `json:"foo"`
	Bar  string `json:"bar"`
	Name string `json:"name"`
}

type Outer struct {
	*Embedded
	Name int `json:"name"` // Overrides the field in Embedded
}

func main() {
	jsonData := []byte(`{"name": 123, "foo": "bar", "bar": "baz"}`)

	var o Outer
	err := json.Unmarshal(jsonData, &o)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	cool := o.Embedded
	fmt.Printf("%#v\n", cool)

	fmt.Printf("Outer.Name: %d\n", o.Name)             // Outer field gets the value
	fmt.Printf("Embedded.Name: %s\n", o.Embedded.Name) // Embedded field is ignored
}
