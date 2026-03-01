package platform

import (
	"runtime"
	"testing"
)

func TestDetect(t *testing.T) {
	info := Detect()

	if info.OS != runtime.GOOS {
		t.Errorf("expected OS %q, got %q", runtime.GOOS, info.OS)
	}
	if info.Arch != runtime.GOARCH {
		t.Errorf("expected Arch %q, got %q", runtime.GOARCH, info.Arch)
	}

	// We can't assert WSL/Docker values since they depend on environment,
	// but the function should not panic.
	t.Logf("Platform: OS=%s Arch=%s WSL=%v Docker=%v", info.OS, info.Arch, info.IsWSL, info.IsDocker)
}
