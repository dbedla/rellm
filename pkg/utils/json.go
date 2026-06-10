package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/fatih/color"
)

func PrintPrettyJsonFromString(s *string) {
	if s == nil {
		panic("no data to print")
	}

	var out bytes.Buffer
	data := []byte(*s)
	err := json.Indent(&out, data, "", "  ")
	if err != nil {
		panic(err)
	}

	color.Yellow(out.String())
}

func PrintAsOnelinerJson(data any) {
	marshaled, err := json.Marshal(data)
	if err != nil {
		fmt.Printf("Error marshaling data: %s\n", err.Error())
		return
	}

	typeName := reflect.TypeOf(data).String()
	fmt.Printf("%s : %s\n", typeName, string(marshaled))
}
