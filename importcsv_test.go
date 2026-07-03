package findai

import (
	"context"
	"net/http"
	"strings"
	"testing"
)

func TestImportCSV_RejectsNonCSVFilename(t *testing.T) {
	c := newTestServerClient(t, func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("server should not be called for a rejected filename")
	})

	_, err := c.ImportCSV(context.Background(), "kt_1", "data.txt", strings.NewReader("a,b\n1,2\n"))
	if err == nil {
		t.Fatal("expected error for non-.csv filename")
	}
}

func TestImportCSV_Success(t *testing.T) {
	c := newTestServerClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/knowledge/templates/kt_1/import" {
			t.Errorf("path = %q", r.URL.Path)
		}
		if err := r.ParseMultipartForm(1 << 20); err != nil {
			t.Fatalf("ParseMultipartForm: %v", err)
		}
		f, header, err := r.FormFile("file")
		if err != nil {
			t.Fatalf("FormFile: %v", err)
		}
		defer f.Close()
		if header.Filename != "companies.csv" {
			t.Errorf("Filename = %q", header.Filename)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{
			"total_rows": 3, "imported": 2,
			"errors": [{"row": 2, "error": "missing required field company_name"}]
		}`))
	})

	resp, err := c.ImportCSV(context.Background(), "kt_1", "companies.csv", strings.NewReader("company_name\nAcme\n\nGlobex\n"))
	if err != nil {
		t.Fatalf("ImportCSV() error = %v", err)
	}
	if resp.TotalRows != 3 || resp.Imported != 2 {
		t.Fatalf("unexpected response: %+v", resp)
	}
	if len(resp.Errors) != 1 || resp.Errors[0].Row != 2 {
		t.Fatalf("unexpected errors: %+v", resp.Errors)
	}
}
