package main

import (
	"io"
	"net/http"
	"rellm/pkg/utils"
)

func GetWeather() (string, error) {
	resp, err := http.Get("https://wttr.in")
	if err != nil {
		return "", err
	}
	defer utils.CloseAndLogIfError_DEFER_ME(resp.Body)

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(raw), nil
}
