package images

import (
	"strings"
	"testing"
	"time"
)

func TestBuildTempShellContainerConfigSetsLifecycleFields(t *testing.T) {
	config := buildTempShellContainerConfig("sha256:abcdef0123456789", time.Unix(1700000000, 0))

	if config.AutoStart {
		t.Fatal("expected AutoStart false so run-and-exec can explicitly start container")
	}

	if config.AutoRemove {
		t.Fatal("expected AutoRemove false so explicit cleanup can run")
	}

	if config.Network != "bridge" {
		t.Fatalf("expected bridge network, got %q", config.Network)
	}

	if config.ImageID != "sha256:abcdef0123456789" {
		t.Fatalf("unexpected image id %q", config.ImageID)
	}

	if !strings.HasPrefix(config.Name, "tmp-shell-") {
		t.Fatalf("name %q missing tmp-shell prefix", config.Name)
	}
}

func TestGenerateTempContainerNameIncludesPrefixAndImageID(t *testing.T) {
	name := generateTempContainerName("sha256:abcdef0123456789", time.Unix(1700000000, 123456000))

	if !strings.HasPrefix(name, "tmp-shell-") {
		t.Fatalf("name %q missing tmp-shell prefix", name)
	}

	if !strings.Contains(name, "abcdef012345") {
		t.Fatalf("name %q missing image ID segment", name)
	}
}

func TestGenerateTempContainerNameChangesWithTime(t *testing.T) {
	first := generateTempContainerName("sha256:abcdef0123456789", time.Unix(1700000000, 0))
	second := generateTempContainerName("sha256:abcdef0123456789", time.Unix(1700000001, 0))

	if first == second {
		t.Fatalf("expected unique names, got %q and %q", first, second)
	}
}
