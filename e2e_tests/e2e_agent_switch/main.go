package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/fatih/color"
)

// Scenario reproduces the e2e_fs_agent flow, but the shared conversation is
// handled by two agents: the OpenRouter (luna) agent performs the read-only
// tasks, then the LMStudio agent takes over the last task (creating the
// output file) — proving the conversation elements are interchangeable.
func main() {

	dirs, orAgent, lmsAgent, err := setup()
	if err != nil {
		panic(err)
	}
	color.Red("Agent space dir: %s", dirs.workspace)

	f1Name := "names.txt"
	f1Content := "John Smith"
	err = createFile(dirs.readOnlyDir, f1Name, f1Content)
	if err != nil {
		panicWithLog("cannot create file", err)
	}

	f2Name := "locations.txt"
	f2Content := "London"
	err = createFile(dirs.readOnlyDir, f2Name, f2Content)
	if err != nil {
		panicWithLog("cannot create file", err)
	}

	ctx := context.Background()

	msg := "what files do you see"
	color.Magenta(msg)

	llmResp, err := orAgent.Ask(ctx, msg)
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

	llmResp, err = orAgent.Ask(ctx, msg)
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

	// LMStudio agent takes over the same conversation for the last task.
	msg = "create file in your output directory, file name 'data.txt', file should contain content of both files from read only director"
	color.Magenta(msg)

	llmResp, err = lmsAgent.Ask(ctx, msg)
	if err != nil {
		panicWithLog("lms agent failed with an openrouter-built conversation, msg: "+msg, err)
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
