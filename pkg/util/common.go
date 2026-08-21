package util

import (
	"encoding/json"
	"io"
)

func DecodeBody[T any](body io.ReadCloser) (*T, error) {
	defer body.Close() // nolint: errcheck
	var data T
	if err := json.NewDecoder(body).Decode(&data); err != nil {
		return nil, err
	}
	return &data, nil
}
