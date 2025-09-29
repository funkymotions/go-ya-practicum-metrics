package models

import (
	"fmt"
	"strconv"
)

const (
	Counter = "counter"
	Gauge   = "gauge"
)

// NOTE: Не усложняем пример, вводя иерархическую вложенность структур.
// Органичиваясь плоской моделью.
// Delta и Value объявлены через указатели,
// что бы отличать значение "0", от не заданного значения
// и соответственно не кодировать в структуру.
type Metrics struct {
	ID      string   `json:"id"`
	MType   string   `json:"type"`
	MTypeID uint     `json:"-"`
	Delta   *int64   `json:"delta,omitempty"`
	Value   *float64 `json:"value,omitempty"`
	Hash    string   `json:"hash,omitempty"`
}

func (m *Metrics) String() string {
	if m.MType == Counter {
		return fmt.Sprintf("%d", *m.Delta)
	}
	if m.MType == Gauge {
		str := strconv.FormatFloat(*m.Value, 'f', -1, 64)
		return str
	}
	return ""
}
