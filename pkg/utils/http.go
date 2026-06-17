package utils

import (
	"encoding/json"
	"fmt"
	"io"
)

func CloseAndLogIfError_DEFER_ME(closeMe io.Closer) {
	err := closeMe.Close()
	if err != nil {
		fmt.Printf("error closing: %s\n", err)
	}
}

func Unmarshall[K any](rawBody []byte) (K, error) {
	var data K
	err := json.Unmarshal(rawBody, &data)
	if err != nil {
		return *new(K), fmt.Errorf("error unmarshalling: %s; unmarshaling type %T; raw: %s", err, data, string(rawBody))
	}

	return data, nil
}
