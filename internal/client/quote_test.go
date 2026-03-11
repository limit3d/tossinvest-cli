package client

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestGetQuoteFromFixtures(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fixturePath := fixturePathForRequest(t, r.URL.Path)
		http.ServeFile(w, r, fixturePath)
	}))
	defer server.Close()

	client := New(Config{
		HTTPClient:  server.Client(),
		InfoBaseURL: server.URL,
	})

	quote, err := client.GetQuote(context.Background(), "005930")
	if err != nil {
		t.Fatalf("GetQuote returned error: %v", err)
	}

	if quote.ProductCode != "A005930" {
		t.Fatalf("unexpected product code: %s", quote.ProductCode)
	}

	if quote.Symbol != "005930" {
		t.Fatalf("unexpected symbol: %s", quote.Symbol)
	}

	if quote.Name != "삼성전자" {
		t.Fatalf("unexpected name: %s", quote.Name)
	}

	if quote.Last != 193900 {
		t.Fatalf("unexpected last price: %v", quote.Last)
	}

	if quote.ReferencePrice != 187900 {
		t.Fatalf("unexpected reference price: %v", quote.ReferencePrice)
	}

	if quote.Volume != 27306483 {
		t.Fatalf("unexpected volume: %v", quote.Volume)
	}
}

func fixturePathForRequest(t *testing.T, path string) string {
	t.Helper()

	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("failed to resolve test path")
	}

	root := filepath.Join(filepath.Dir(filename), "..", "..", "fixtures", "responses", "public")

	switch path {
	case "/api/v2/stock-infos/A005930":
		return mustPublicFixturePath(t, filepath.Join(root, "stock-info.json"))
	case "/api/v1/stock-detail/ui/A005930/common":
		return mustPublicFixturePath(t, filepath.Join(root, "stock-detail-common.json"))
	case "/api/v1/product/stock-prices":
		return mustPublicFixturePath(t, filepath.Join(root, "stock-price.json"))
	default:
		t.Fatalf("unexpected request path: %s", path)
		return ""
	}
}

func mustPublicFixturePath(t *testing.T, path string) string {
	t.Helper()
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("fixture missing: %s: %v", path, err)
	}
	return path
}
