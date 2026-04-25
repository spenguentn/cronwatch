package main

import (
	"os"
	"os/exec"
	"testing"
)

// TestVersion_Flag verifies the binary exits 0 and prints version when
// the -version flag is provided. It compiles the binary on the fly so
// the test remains self-contained and does not depend on a pre-built
// artefact.
func TestVersion_Flag(t *testing.T) {
	if os.Getenv("CRONWATCH_INTEGRATION") == "" {
		t.Skip("set CRONWATCH_INTEGRATION=1 to run binary integration tests")
	}

	tmp := t.TempDir()
	bin := tmp + "/cronwatch"

	build := exec.Command("go", "build", "-o", bin, ".")
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build failed: %v\n%s", err, out)
	}

	cmd := exec.Command(bin, "-version")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("expected exit 0, got: %v\noutput: %s", err, out)
	}

	if len(out) == 0 {
		t.Error("expected version output, got empty string")
	}
}

// TestMissingConfig_NonZeroExit verifies the binary exits non-zero when
// the specified config file does not exist.
func TestMissingConfig_NonZeroExit(t *testing.T) {
	if os.Getenv("CRONWATCH_INTEGRATION") == "" {
		t.Skip("set CRONWATCH_INTEGRATION=1 to run binary integration tests")
	}

	tmp := t.TempDir()
	bin := tmp + "/cronwatch"

	build := exec.Command("go", "build", "-o", bin, ".")
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build failed: %v\n%s", err, out)
	}

	cmd := exec.Command(bin, "-config", "/nonexistent/path/cronwatch.yaml")
	if err := cmd.Run(); err == nil {
		t.Fatal("expected non-zero exit for missing config, got exit 0")
	}
}
