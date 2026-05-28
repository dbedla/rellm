package utils

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
)

func PressEnterToContinue() {
	color.Red("=============  Press Enter to continue... =============")
	reader := bufio.NewReader(os.Stdin)

	_, errInput := reader.ReadString('\n')
	if errInput != nil {
		panic(errInput)
	}

	color.Red("continue...")
}

func ReadConsoleInput() string {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("user input:\n")
	input, _ := reader.ReadString('\n')
	return strings.TrimSpace(input)
}
