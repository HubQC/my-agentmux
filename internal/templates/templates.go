package templates

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Template represents a reusable agent definition template.
type Template struct {
	Name        string            `yaml:"name"`
	Description string            `yaml:"description"`
	AgentType   string            `yaml:"agent_type"`
	Prompt      string            `yaml:"prompt"`
	Env         map[string]string `yaml:"env,omitempty"`
	Tags        []string          `yaml:"tags,omitempty"`
}

// BuiltinTemplates returns the curated set of built-in agent templates.
func BuiltinTemplates() []Template {
	return []Template{
		{
			Name:        "code-reviewer",
			Description: "Reviews code for bugs, style issues, and best practices",
			AgentType:   "claude",
			Prompt:      "You are an expert code reviewer. Review the codebase for bugs, security issues, performance problems, and style violations. Provide actionable suggestions with file paths and line numbers.",
			Tags:        []string{"review", "quality"},
		},
		{
			Name:        "test-writer",
			Description: "Generates comprehensive unit and integration tests",
			AgentType:   "claude",
			Prompt:      "You are a test engineer. Analyze the codebase and write comprehensive unit tests and integration tests. Aim for high code coverage, edge cases, and error paths. Use the project's existing test framework.",
			Tags:        []string{"testing", "quality"},
		},
		{
			Name:        "docs-generator",
			Description: "Generates or updates project documentation",
			AgentType:   "claude",
			Prompt:      "You are a technical writer. Review the codebase and generate clear, comprehensive documentation. Include API references, usage examples, architecture overviews, and setup guides.",
			Tags:        []string{"documentation"},
		},
		{
			Name:        "refactorer",
			Description: "Identifies and applies refactoring opportunities",
			AgentType:   "claude",
			Prompt:      "You are a refactoring expert. Analyze the codebase for code smells, duplication, complex functions, and architectural issues. Apply clean code principles and suggest or implement refactoring improvements.",
			Tags:        []string{"refactoring", "quality"},
		},
		{
			Name:        "security-auditor",
			Description: "Performs security analysis and vulnerability scanning",
			AgentType:   "claude",
			Prompt:      "You are a security engineer. Audit the codebase for security vulnerabilities including injection attacks, authentication issues, data leaks, dependency vulnerabilities, and OWASP Top 10 risks.",
			Tags:        []string{"security", "audit"},
		},
		{
			Name:        "performance-optimizer",
			Description: "Identifies and fixes performance bottlenecks",
			AgentType:   "claude",
			Prompt:      "You are a performance engineer. Profile the codebase for performance bottlenecks, memory leaks, unnecessary allocations, N+1 queries, and inefficient algorithms. Suggest and implement optimizations.",
			Tags:        []string{"performance"},
		},
		{
			Name:        "architect",
			Description: "Designs system architecture and creates implementation plans",
			AgentType:   "claude",
			Prompt:      "You are a software architect. Analyze the project requirements and existing codebase to design scalable, maintainable architectures. Create detailed implementation plans with component diagrams and API contracts.",
			Tags:        []string{"architecture", "planning"},
		},
		{
			Name:        "debugger",
			Description: "Investigates and fixes bugs with systematic approach",
			AgentType:   "claude",
			Prompt:      "You are a debugging expert. Use a systematic approach to investigate bugs: reproduce, isolate, identify root cause, and implement fixes. Add regression tests to prevent recurrence.",
			Tags:        []string{"debugging", "fixing"},
		},
	}
}

// InstallTemplate writes a template as a Markdown agent definition file.
func InstallTemplate(tmpl Template, agentsDir string) (string, error) {
	if err := os.MkdirAll(agentsDir, 0o755); err != nil {
		return "", fmt.Errorf("creating agents directory: %w", err)
	}

	filename := tmpl.Name + ".md"
	filePath := filepath.Join(agentsDir, filename)

	var b strings.Builder
	b.WriteString("---\n")
	b.WriteString(fmt.Sprintf("description: %s\n", tmpl.Description))
	b.WriteString(fmt.Sprintf("agent_type: %s\n", tmpl.AgentType))
	if len(tmpl.Env) > 0 {
		b.WriteString("env:\n")
		for k, v := range tmpl.Env {
			b.WriteString(fmt.Sprintf("  %s: %q\n", k, v))
		}
	}
	b.WriteString("---\n\n")
	b.WriteString(tmpl.Prompt + "\n")

	if err := os.WriteFile(filePath, []byte(b.String()), 0o644); err != nil {
		return "", fmt.Errorf("writing template: %w", err)
	}

	return filePath, nil
}

// FindByTag returns templates with a matching tag.
func FindByTag(tag string) []Template {
	var matches []Template
	for _, t := range BuiltinTemplates() {
		for _, tt := range t.Tags {
			if strings.EqualFold(tt, tag) {
				matches = append(matches, t)
				break
			}
		}
	}
	return matches
}

// FindByName returns a template by exact name.
func FindByName(name string) *Template {
	for _, t := range BuiltinTemplates() {
		if t.Name == name {
			return &t
		}
	}
	return nil
}
