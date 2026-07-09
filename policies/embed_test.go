package policies

import (
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

// TestDefaultEmbedded verifies the embedded default policy is non-empty and
// looks like the policy YAML we expect (basic shape check; full schema
// validation is covered by internal/policy tests loading these same bytes).
func TestDefaultEmbedded(t *testing.T) {
	if len(Default) == 0 {
		t.Fatal("policies.Default is empty — go:embed did not pick up default.yaml")
	}

	var doc struct {
		Version int    `yaml:"version"`
		Name    string `yaml:"name"`
		Rules   []any  `yaml:"rules"`
	}
	if err := yaml.Unmarshal(Default, &doc); err != nil {
		t.Fatalf("embedded default.yaml does not parse as YAML: %v", err)
	}
	if doc.Name != "default" {
		t.Errorf("name = %q, want %q", doc.Name, "default")
	}
	if len(doc.Rules) == 0 {
		t.Error("embedded default policy has no rules")
	}
	if !strings.Contains(string(Default), "secret-in-args") {
		t.Error("embedded default policy missing expected secret-in-args rule")
	}
}
