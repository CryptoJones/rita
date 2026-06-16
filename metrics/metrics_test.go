package metrics

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestObserveImportRecordsResult(t *testing.T) {
	ImportsTotal.Reset()

	ObserveImport(nil, 1.5)
	ObserveImport(errors.New("boom"), 2.0)
	ObserveImport(nil, 0.5)

	body := scrape(t)
	for _, want := range []string{
		`rita_import_runs_total{result="success"} 2`,
		`rita_import_runs_total{result="error"} 1`,
		"rita_import_duration_seconds_bucket",
	} {
		if !strings.Contains(body, want) {
			t.Errorf("scrape missing %q\n---\n%s", want, body)
		}
	}
}

func TestObserveAnalysisRecordsResult(t *testing.T) {
	AnalysisRunsTotal.Reset()

	ObserveAnalysis(nil)
	ObserveAnalysis(errors.New("nope"))

	body := scrape(t)
	for _, want := range []string{
		`rita_analysis_runs_total{result="success"} 1`,
		`rita_analysis_runs_total{result="error"} 1`,
	} {
		if !strings.Contains(body, want) {
			t.Errorf("scrape missing %q\n---\n%s", want, body)
		}
	}
}

func TestHealthEndpoint(t *testing.T) {
	srv := httptest.NewServer(handler())
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/healthz")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	body, _ := io.ReadAll(resp.Body)
	if strings.TrimSpace(string(body)) != "ok" {
		t.Fatalf("expected body 'ok', got %q", body)
	}
}

// scrape hits the /metrics endpoint of a freshly-built handler and returns the body.
func scrape(t *testing.T) string {
	t.Helper()
	srv := httptest.NewServer(handler())
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/metrics")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	return string(body)
}
