package env

import (
	"fmt"
	"strconv"
	"strings"
)

type Endpoint struct {
	Hostname string
	Port     uint
}

func (e *Endpoint) String() string {
	return fmt.Sprintf("%s:%d", e.Hostname, e.Port)
}

func (e *Endpoint) Set(in string) error {
	parts := strings.Split(in, ":")
	if len(parts) != 2 {
		return fmt.Errorf("invalid endpoint format")
	}
	e.Hostname = parts[0]
	if port, err := strconv.ParseUint(parts[1], 10, 32); err != nil {
		return fmt.Errorf("invalid port: %v", err)
	} else {
		e.Port = uint(port)
	}
	return nil
}
