// Command basic_crud demonstrates listing dataset templates and running a
// full create/read/update/delete lifecycle on a record.
//
// Usage:
//
//	FINDAI_API_KEY=fai_... FINDAI_BASE_URL=https://api.example.com FINDAI_TEMPLATE_ID=kt_... go run ./examples/basic_crud
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	findai "github.com/Diferentt/find-ai-sdk-go"
)

func main() {
	client, err := findai.NewClient(
		mustEnv("FINDAI_API_KEY"),
		findai.WithBaseURL(mustEnv("FINDAI_BASE_URL")),
	)
	if err != nil {
		log.Fatal(err)
	}

	templateID := mustEnv("FINDAI_TEMPLATE_ID")
	ctx := context.Background()

	tmpl, err := client.GetTemplate(ctx, templateID)
	if err != nil {
		log.Fatalf("GetTemplate: %v", err)
	}
	fmt.Printf("Template %q has %d fields\n", tmpl.Name, len(tmpl.Fields))

	rec, err := client.CreateRecord(ctx, templateID, map[string]any{
		"company_name": "Acme Corp",
	})
	if err != nil {
		log.Fatalf("CreateRecord: %v", err)
	}
	fmt.Printf("Created record %s\n", rec.ID)

	rec, err = client.UpdateRecord(ctx, templateID, rec.ID, map[string]any{
		"company_name": "Acme Corporation",
	})
	if err != nil {
		log.Fatalf("UpdateRecord: %v", err)
	}
	fmt.Printf("Updated record: %v\n", rec.ValuesData)

	page, err := client.ListRecords(ctx, templateID, findai.ListRecordsOptions{Limit: 5})
	if err != nil {
		log.Fatalf("ListRecords: %v", err)
	}
	fmt.Printf("Template has %d record(s) total\n", page.Total)

	if err := client.DeleteRecord(ctx, templateID, rec.ID); err != nil {
		log.Fatalf("DeleteRecord: %v", err)
	}
	fmt.Println("Deleted record", rec.ID)
}

func mustEnv(name string) string {
	v := os.Getenv(name)
	if v == "" {
		log.Fatalf("environment variable %s is required", name)
	}
	return v
}
