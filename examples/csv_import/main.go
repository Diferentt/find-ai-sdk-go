// Command csv_import demonstrates bulk-loading records into a dataset from
// a local CSV file.
//
// Usage:
//
//	FINDAI_API_KEY=fai_... FINDAI_BASE_URL=https://api.example.com FINDAI_TEMPLATE_ID=kt_... go run ./examples/csv_import companies.csv
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	findai "github.com/Diferentt/find-ai-sdk-go"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("usage: csv_import <path-to-file.csv>")
	}
	path := os.Args[1]

	client, err := findai.NewClient(
		mustEnv("FINDAI_API_KEY"),
		findai.WithBaseURL(mustEnv("FINDAI_BASE_URL")),
	)
	if err != nil {
		log.Fatal(err)
	}

	f, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	templateID := mustEnv("FINDAI_TEMPLATE_ID")
	result, err := client.ImportCSV(context.Background(), templateID, path, f)
	if err != nil {
		log.Fatalf("ImportCSV: %v", err)
	}

	fmt.Printf("Imported %d/%d rows\n", result.Imported, result.TotalRows)
	for _, rowErr := range result.Errors {
		fmt.Printf("  row %d: %s\n", rowErr.Row, rowErr.Error)
	}
}

func mustEnv(name string) string {
	v := os.Getenv(name)
	if v == "" {
		log.Fatalf("environment variable %s is required", name)
	}
	return v
}
