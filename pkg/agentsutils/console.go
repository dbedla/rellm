package agentsutils

import (
	"bufio"
	"fmt"
	"os"
	"strings"
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
