package path

import (
	"fmt"
)

type Path struct {
	Input string
}

// String fulfils the flag.Value interface https://pkg.go.dev/flag#Value
func (v Path) String() string {
	return fmt.Sprint(v.Input)
}

// Get fulfils the flag.Getter interface https://pkg.go.dev/flag#Getter
func (v *Path) Get() Path {
	return *v
}

// Set fulfils the flag.Value interface https://pkg.go.dev/flag#Value
func (v *Path) Set(s string) error {
	return nil
}
