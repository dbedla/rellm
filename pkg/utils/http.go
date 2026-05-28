package utils

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func CloseAndLogIfError_DEFER_ME(closeMe io.Closer) {
	err := closeMe.Close()
	if err != nil {
		fmt.Printf("error closing: %s\n", err)
	}
}

func RadAllAndUnmarshall[K any](input io.ReadCloser) (K, error) {
	rawBody, err := io.ReadAll(input)
	if err != nil {
		return *new(K), fmt.Errorf("error reading body: %s", err)
	}

	var data K
	err = json.Unmarshal(rawBody, &data)
	if err != nil {
		fmt.Printf("Error unmarshalling: %s", err)
		PrintAsOnelinerJson(rawBody)
	}

	return data, nil
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

func RadAllToString(input io.ReadCloser) (string, error) {
	rawBody, err := io.ReadAll(input)
	if err != nil {
		return "", fmt.Errorf("error reading body: %s", err)
	}

	return string(rawBody), nil
}

func WriteHeaderStatusBody(resp http.ResponseWriter, status int, body []byte) {
	resp.Header().Set("Content-Type", "application/json")
	resp.WriteHeader(status)
	resp.Write(body)
}
