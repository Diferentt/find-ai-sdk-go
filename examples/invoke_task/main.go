// Command invoke_task runs a task (a flow executed without a conversation)
// and prints its result — the "call a flow like a lambda" path.
//
// Usage:
//
//	FINDAI_API_KEY=fai_... FINDAI_BASE_URL=https://api.example.com go run ./examples/invoke_task <task-id> [key=value ...]
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	findai "github.com/Diferentt/find-ai-sdk-go"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("usage: invoke_task <task-id> [key=value ...]")
	}
	taskID := os.Args[1]

	params := map[string]any{}
	for _, arg := range os.Args[2:] {
		key, value, ok := strings.Cut(arg, "=")
		if !ok {
			log.Fatalf("parameter %q is not key=value", arg)
		}
		params[key] = value
	}

	client, err := findai.NewClient(
		mustEnv("FINDAI_API_KEY"),
		findai.WithBaseURL(mustEnv("FINDAI_BASE_URL")),
	)
	if err != nil {
		log.Fatal(err)
	}
	ctx := context.Background()

	// Which parameters does this task take? Derived from the flow's nodes.
	task, err := client.GetTask(ctx, taskID)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("task %q accepts params: %v\n", task.Name, task.Inputs)

	res, err := client.InvokeTask(ctx, taskID, params)
	if err != nil {
		log.Fatal(err)
	}
	if !res.Success {
		log.Fatalf("flow failed after %dms: %s", res.DurationMS, *res.Error)
	}

	out, _ := json.MarshalIndent(res.Output, "", "  ")
	fmt.Printf("done in %dms, output:\n%s\n", res.DurationMS, out)
}

func mustEnv(name string) string {
	v := os.Getenv(name)
	if v == "" {
		log.Fatalf("%s is required", name)
	}
	return v
}
