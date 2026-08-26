package examplesutils

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"rellm/pkg/rellm"
	"strings"

	"github.com/fatih/color"
)

func ReadConsoleInput() (string, error) {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("user input:\n")
	input, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(input), nil
}

func InspectWithReqLog(req *rellm.ResponsesAPIReq) {
	color.White(" === REQ ===")
	b, err := json.Marshal(req)
	if err != nil {
		panic(err)
	}
	color.White(string(b))
}

func InspectWithRespLog(req *rellm.ResponsesAPIResp) {
	color.White(" === RESP ===")
	b, err := json.Marshal(req)
	if err != nil {
		panic(err)
	}
	color.White(string(b))
}
