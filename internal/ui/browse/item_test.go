package browse

import (
	"testing"

	"github.com/givensuman/containertui/internal/registry"
)

func TestBrowseItemDescriptionShortensByFewCharacters(t *testing.T) {
	item := BrowseItem{Image: registry.RegistryImage{ShortDescription: "hello world"}}

	got := item.Description()
	want := "   hello wo"

	if got != want {
		t.Fatalf("Description() = %q, want %q", got, want)
	}
}

func TestBrowseItemDescriptionKeepsVeryShortText(t *testing.T) {
	item := BrowseItem{Image: registry.RegistryImage{ShortDescription: "abc"}}

	got := item.Description()
	want := "   abc"

	if got != want {
		t.Fatalf("Description() = %q, want %q", got, want)
	}
}

func TestBrowseItemDescriptionOnlyTrimsFewCharacters(t *testing.T) {
	item := BrowseItem{Image: registry.RegistryImage{ShortDescription: "12345678901234567890"}}

	got := item.Description()
	want := "   12345678901234567"

	if got != want {
		t.Fatalf("Description() = %q, want %q", got, want)
	}
}
