// Command search demonstrates full-text and semantic search over a dataset.
//
// Usage:
//
//	FINDAI_API_KEY=fai_... FINDAI_BASE_URL=https://api.example.com FINDAI_TEMPLATE_ID=kt_... go run ./examples/search "acme"
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
		log.Fatal("usage: search <query>")
	}
	query := os.Args[1]

	client, err := findai.NewClient(
		mustEnv("FINDAI_API_KEY"),
		findai.WithBaseURL(mustEnv("FINDAI_BASE_URL")),
	)
	if err != nil {
		log.Fatal(err)
	}

	templateID := mustEnv("FINDAI_TEMPLATE_ID")
	ctx := context.Background()

	fulltext, err := client.Search(ctx, templateID, findai.SearchRequest{Query: query, Limit: 10})
	if err != nil {
		log.Fatalf("Search: %v", err)
	}
	fmt.Println("Full-text hits:")
	for _, hit := range fulltext.Hits {
		fmt.Printf("  %s: %v\n", hit.RecordID, hit.ValuesData)
	}

	semantic, err := client.SemanticSearch(ctx, templateID, findai.SemanticSearchRequest{Query: query, TopK: 10})
	if err != nil {
		log.Fatalf("SemanticSearch: %v", err)
	}
	fmt.Println("Semantic results:")
	for _, r := range semantic.Results {
		fmt.Printf("  score=%.3f: %v\n", r.Score, r.ValuesData)
	}
}

func mustEnv(name string) string {
	v := os.Getenv(name)
	if v == "" {
		log.Fatalf("environment variable %s is required", name)
	}
	return v
}
