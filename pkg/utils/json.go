package utils

import (
	"encoding/json"
	"fmt"
	"reflect"
)

func PrintAsOnelinerJson(data any) {
	marshaled, err := json.Marshal(data)
	if err != nil {
		fmt.Printf("Error marshaling data: %s\n", err.Error())
		return
	}

	typeName := reflect.TypeOf(data).String()
	fmt.Printf("%s : %s\n", typeName, string(marshaled))
}
