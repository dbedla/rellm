package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"reflect"

	"github.com/fatih/color"
	"github.com/invopop/jsonschema"
)

//func PrintJsonSchema[T any]() {
//
//	var reflectMe T
//
//	reflector := jsonschema.Reflector{}
//	schema := reflector.Reflect(&reflectMe)
//
//	b, _ := json.MarshalIndent(schema, "", "  ")
//	fmt.Println(string(b))
//}

func PrintJsonSchema[T any]() {

	b, err := GetSchema[T]()
	if err != nil {
		fmt.Println("schema generation err: " + err.Error())
	}
	color.Yellow(string(b))
}

func GetSchema[T any]() ([]byte, error) {

	var reflectMe T

	reflector := jsonschema.Reflector{}
	schema := reflector.Reflect(&reflectMe)

	b, err := json.MarshalIndent(schema, "", "  ")
	if err != nil {
		return nil, err
	}

	return b, nil
}

func PrintRawMessages(messages []json.RawMessage) {
	for i, raw := range messages {
		var out bytes.Buffer
		err := json.Indent(&out, raw, "", "  ")
		if err != nil {
			fmt.Printf("Item %d (invalid JSON): %s\n", i, raw)
			continue
		}

		fmt.Printf("Item %d:\n%s\n\n", i, out.String())
	}
}

func PrintPrettyJson(data []byte) {
	var out bytes.Buffer

	err := json.Indent(&out, data, "", "  ")
	if err != nil {
		panic(err)
	}

	color.Yellow(out.String())
}

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

func ToPrettyJson(data []byte) string {
	var out bytes.Buffer

	err := json.Indent(&out, data, "", "  ")
	if err != nil {
		panic(err)
	}

	return out.String() + "\n"
}

func PrintAsPrettyJson(data any) {
	marshaled, err := json.Marshal(data)
	if err != nil {
		fmt.Printf("Error marshaling data: %s\n", err.Error())
		return
	}
	PrintPrettyJson(marshaled)
}

func PrintfResponseBodyAsPrettyJson(body io.ReadCloser) {
	data, err := io.ReadAll(body)
	if err != nil {
		fmt.Printf("Error reading response body: %s\n", err.Error())
		return
	}
	PrintPrettyJson(data)
}

func UnmarshallFromStringr[T any](input string) (T, error) {
	var data T
	err := json.Unmarshal([]byte(input), &data)
	if err != nil {
		return *new(T), fmt.Errorf("error parsing json: %s", err)
	}
	return data, nil
}

func MarshallToRawRequestBody(data any) io.ReadCloser {
	body, err := json.Marshal(data)
	if err != nil {
		panic(err)
	}

	reqBody := io.NopCloser(bytes.NewBuffer(body))

	return reqBody
}

func MarshallToByte(data any) []byte {
	body, err := json.Marshal(data)
	if err != nil {
		panic(err)
	}

	return body
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

func AsOnelinerJson(data any) string {
	marshaled, err := json.Marshal(data)
	if err != nil {
		s := fmt.Sprintf("Error marshaling data: %s\n", err.Error())
		panic(s)
		return ""
	}

	return string(marshaled)
}
