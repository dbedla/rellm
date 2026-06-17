package main

import (
	"fmt"
	"io"
	"net/http"
)

func GetWeather() (string, error) {
	resp, err := http.Get("https://wttr.in")
	if err != nil {
		return "", err
	}
	defer func() {
		err := resp.Body.Close()
		if err != nil {
			fmt.Printf("error closing: %s\n", err)
		}
	}()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(raw), nil
}
