package util

import (
	"os"
	"regexp"
	"strings"
	"testing"
)

func TestSeedSKUsHaveProductMasterRows(t *testing.T) {
	t.Parallel()

	sourceBytes, err := os.ReadFile("seed.go")
	if err != nil {
		t.Fatalf("read seed.go: %v", err)
	}
	source := string(sourceBytes)

	productSection := sourceSection(t, source, "func seedProducts", "func seedSales")
	productSKUs := collectSKUs(productSection)

	referenceSections := []string{
		sourceSection(t, source, "func seedTopProducts", "func seedInventory"),
		sourceSection(t, source, "func seedInventory", "func seedDeliveries"),
	}

	for _, section := range referenceSections {
		for sku := range collectSKUs(section) {
			if !productSKUs[sku] {
				t.Fatalf("seed references sku %q without a products row", sku)
			}
		}
	}
}

func sourceSection(t *testing.T, source string, start string, end string) string {
	t.Helper()

	startIndex := strings.Index(source, start)
	if startIndex < 0 {
		t.Fatalf("missing source section start %q", start)
	}
	endIndex := strings.Index(source[startIndex:], end)
	if endIndex < 0 {
		t.Fatalf("missing source section end %q", end)
	}

	return source[startIndex : startIndex+endIndex]
}

func collectSKUs(source string) map[string]bool {
	matches := regexp.MustCompile(`SKU:\s*"([^"]+)"`).FindAllStringSubmatch(source, -1)
	skus := make(map[string]bool, len(matches))
	for _, match := range matches {
		skus[match[1]] = true
	}

	return skus
}
