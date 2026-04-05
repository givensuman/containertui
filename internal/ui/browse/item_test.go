package browse

import (
	"testing"

	"github.com/givensuman/containertui/internal/registry"
)

func TestBrowseItemDescriptionKeepsShortDescriptionUnderMaxWidth(t *testing.T) {
	item := BrowseItem{Image: registry.RegistryImage{ShortDescription: "hello world"}}

	got := item.Description()
	want := "   hello world"

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

func TestBrowseItemDescriptionTruncatesAndAddsEllipsisAtMaxWidth(t *testing.T) {
	item := BrowseItem{Image: registry.RegistryImage{ShortDescription: "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789extra"}}

	got := item.Description()
	want := "   abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ012345678..."

	if got != want {
		t.Fatalf("Description() = %q, want %q", got, want)
	}
}
