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
		fmt.Printf("Error unmarshalling: %s", err)
		PrintAsOnelinerJson(rawBody)
		return *new(K), err
	}

	return data, nil
}
