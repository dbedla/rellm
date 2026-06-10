package utils

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func ReadConsoleInput() string {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("user input:\n")
	input, _ := reader.ReadString('\n')
	return strings.TrimSpace(input)
}
