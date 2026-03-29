package plugin

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestHookRunnerRegisterAndFire(t *testing.T) {
	hr := NewHookRunner()

	hooks := []Hook{
		{Event: HookPreStart, Command: "echo pre_start"},
		{Event: HookPostStart, Command: "echo post_start"},
	}
	hr.Register("test-agent", hooks)

	hctx := HookContext{
		AgentName: "test-agent",
		AgentType: "shell",
		WorkDir:   t.TempDir(),
		Event:     HookPreStart,
	}

	errs := hr.Fire(context.Background(), hctx)
	if len(errs) != 0 {
		t.Errorf("expected no errors, got %v", errs)
	}
}

func TestHookRunnerFireNonMatchingEvent(t *testing.T) {
	hr := NewHookRunner()

	hooks := []Hook{
		{Event: HookPreStart, Command: "echo pre_start"},
	}
	hr.Register("test-agent", hooks)

	// Fire a different event — should be a no-op
	hctx := HookContext{
		AgentName: "test-agent",
		Event:     HookPostStop,
	}

	errs := hr.Fire(context.Background(), hctx)
	if len(errs) != 0 {
		t.Errorf("expected no errors for non-matching event, got %v", errs)
	}
}

func TestHookRunnerFireUnregisteredAgent(t *testing.T) {
	hr := NewHookRunner()

	hctx := HookContext{
		AgentName: "nonexistent-agent",
		Event:     HookPreStart,
	}

	errs := hr.Fire(context.Background(), hctx)
	if errs != nil {
		t.Errorf("expected nil for unregistered agent, got %v", errs)
	}
}

func TestHookRunnerUnregister(t *testing.T) {
	hr := NewHookRunner()

	hr.Register("test-agent", []Hook{
		{Event: HookPreStart, Command: "echo hello"},
	})
	hr.Unregister("test-agent")

	hctx := HookContext{
		AgentName: "test-agent",
		Event:     HookPreStart,
	}

	errs := hr.Fire(context.Background(), hctx)
	if errs != nil {
		t.Errorf("expected nil after unregister, got %v", errs)
	}
}

func TestHookRunnerFireCommandFailure(t *testing.T) {
	hr := NewHookRunner()

	hooks := []Hook{
		{Event: HookPreStart, Command: "false"}, // `false` exits with 1
	}
	hr.Register("test-agent", hooks)

	hctx := HookContext{
		AgentName: "test-agent",
		Event:     HookPreStart,
	}

	errs := hr.Fire(context.Background(), hctx)
	if len(errs) == 0 {
		t.Error("expected error from failing command")
	}
}

func TestHookEventConstants(t *testing.T) {
	events := []HookEvent{
		HookPreStart, HookPostStart, HookPreStop,
		HookPostStop, HookOnOutput, HookOnError,
	}
	expected := []string{
		"pre_start", "post_start", "pre_stop",
		"post_stop", "on_output", "on_error",
	}

	for i, event := range events {
		if string(event) != expected[i] {
			t.Errorf("HookEvent %d = %q, want %q", i, string(event), expected[i])
		}
	}
}

func TestDiscoverPlugins(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a non-executable file (should be skipped)
	nonExec := filepath.Join(tmpDir, "not-a-plugin.txt")
	os.WriteFile(nonExec, []byte("not a plugin"), 0o644)

	// Create an executable file
	execFile := filepath.Join(tmpDir, "agentmux-plugin-test")
	os.WriteFile(execFile, []byte("#!/bin/sh\necho '{}'"), 0o755)

	plugins, err := DiscoverPlugins(tmpDir)
	if err != nil {
		t.Fatalf("DiscoverPlugins failed: %v", err)
	}

	foundExec := false
	for _, p := range plugins {
		if filepath.Base(p) == "agentmux-plugin-test" {
			foundExec = true
		}
		if filepath.Base(p) == "not-a-plugin.txt" {
			t.Error("non-executable file should not be discovered")
		}
	}

	if !foundExec {
		t.Error("expected to discover the executable plugin")
	}
}

func TestPluginProtocolCreation(t *testing.T) {
	pp := NewPluginProtocol("echo")

	if pp == nil {
		t.Fatal("expected non-nil PluginProtocol")
	}
	if pp.binary != "echo" {
		t.Errorf("binary = %q, want %q", pp.binary, "echo")
	}
}
