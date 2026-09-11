package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"rellm/pkg/rellm"
	"strings"

	"github.com/fatih/color"
)

func main() {

	providerFlag := argsToFlag(os.Args)
	if providerFlag == flag_Invalid {
		help()
		return
	}

	dirs, fsAgent, err := setup(providerFlag)
	if err != nil {
		panic(err)
	}
	color.Red("Agent space dir: %s", dirs.workspace)

	scenario(fsAgent, dirs)
}

func scenario(fsAgent *rellm.Agent, dirs agentsDirs) {
	f1Name := "names.txt"
	f1Content := "John Smith"
	err := createFile(dirs.readOnlyDir, f1Name, f1Content)
	if err != nil {
		panicWithLog("cannot create file", err)
	}

	f2Name := "locations.txt"
	f2Content := "London"
	err = createFile(dirs.readOnlyDir, f2Name, f2Content)
	if err != nil {
		panicWithLog("cannot create file", err)
	}

	msg := "what files do you see"
	color.Magenta(msg)

	ctx := context.Background()
	llmResp, err := fsAgent.Ask(ctx, msg)
	if err != nil {
		panicWithLog("agent failed for input msg: "+msg, err)
	}
	color.Green(llmResp.Message)

	if !strings.Contains(llmResp.Message, f1Name) {
		panicWithLog("missing expected file name '"+f1Name+"' in llm output", fmt.Errorf("missing file name"))
	}
	if !strings.Contains(llmResp.Message, f2Name) {
		panicWithLog("missing expected file name '"+f2Name+"' in llm output", fmt.Errorf("missing file name"))
	}

	msg = "show me what is in files you mention"
	color.Magenta(msg)

	llmResp, err = fsAgent.Ask(ctx, msg)
	if err != nil {
		panicWithLog("agent failed for input msg: "+msg, err)
	}
	color.Green(llmResp.Message)

	if !strings.Contains(llmResp.Message, f1Content) {
		panicWithLog("missing expected file content '"+f1Content+"' in llm output", fmt.Errorf("missing file content"))
	}
	if !strings.Contains(llmResp.Message, f2Content) {
		panicWithLog("missing expected file content '"+f2Content+"' in llm output", fmt.Errorf("missing file content"))
	}

	msg = "create file in your output directory, file name 'data.txt', file should contain content of both files from read only director"
	color.Magenta(msg)

	llmResp, err = fsAgent.Ask(ctx, msg)
	if err != nil {
		panicWithLog("agent failed for input msg: "+msg, err)
	}
	color.Green(llmResp.Message)

	dataPath := filepath.Join(dirs.outputDir, "data.txt")
	rawData, err := os.ReadFile(dataPath)
	if err != nil {
		panicWithLog("no output file", err)
	}

	outputContent := string(rawData)

	if !strings.Contains(outputContent, f1Content) {
		panicWithLog("missing expected output file content '"+f1Content+"'", fmt.Errorf("missing file content"))
	}
	if !strings.Contains(outputContent, f2Content) {
		panicWithLog("missing expected output file content '"+f2Content+"'", fmt.Errorf("missing file content"))
	}
}
