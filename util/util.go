package util

import (
	"bytes"
	"encoding/json"
)

type Number float64

var _ json.Unmarshaler = (*Number)(nil)

func (n *Number) UnmarshalJSON(data []byte) error {
	var f float64
	if err := json.Unmarshal(bytes.Trim(data, `"`), &f); err != nil {
		return err
	}
	*n = Number(f)
	return nil
}

type M map[string]interface{}
