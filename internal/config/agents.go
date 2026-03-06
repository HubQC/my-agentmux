package config

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// AgentDefinition represents a user-defined agent configuration loaded
// from a Markdown file with YAML frontmatter in the agents directory.
type AgentDefinition struct {
	// Name is derived from the filename (without extension).
	Name string `yaml:"-"`

	// Description is a short summary of the agent's purpose.
	Description string `yaml:"description"`

	// AgentType is the base preset to use (claude, aider, codex, etc.).
	AgentType string `yaml:"agent_type"`

	// Args are extra arguments to pass to the agent CLI.
	Args []string `yaml:"args,omitempty"`

	// Env are environment variables to set for this agent.
	Env map[string]string `yaml:"env,omitempty"`

	// WorkDir overrides the working directory.
	WorkDir string `yaml:"workdir,omitempty"`

	// SystemPrompt is the content body (Markdown after frontmatter).
	SystemPrompt string `yaml:"-"`

	// SourceFile is the path to the definition file.
	SourceFile string `yaml:"-"`
}

// LoadAgentDefinitions reads all agent definition files from the agents directory.
func LoadAgentDefinitions(agentsDir string) ([]AgentDefinition, error) {
	entries, err := os.ReadDir(agentsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("reading agents directory: %w", err)
	}

	var defs []AgentDefinition
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		ext := filepath.Ext(entry.Name())
		if ext != ".md" && ext != ".yaml" && ext != ".yml" {
			continue
		}

		filePath := filepath.Join(agentsDir, entry.Name())
		def, err := LoadAgentDefinition(filePath)
		if err != nil {
			// Skip malformed files but log warning
			fmt.Fprintf(os.Stderr, "Warning: skipping %s: %v\n", entry.Name(), err)
			continue
		}
		defs = append(defs, *def)
	}

	return defs, nil
}

// LoadAgentDefinition reads a single agent definition file.
func LoadAgentDefinition(filePath string) (*AgentDefinition, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("reading agent definition: %w", err)
	}

	def := &AgentDefinition{
		Name:       fileBaseName(filePath),
		SourceFile: filePath,
	}

	content := string(data)

	// Parse YAML frontmatter if present
	if strings.HasPrefix(content, "---") {
		frontmatter, body, err := parseFrontmatter(content)
		if err != nil {
			return nil, err
		}

		decoder := yaml.NewDecoder(strings.NewReader(frontmatter))
		decoder.KnownFields(true)
		if err := decoder.Decode(def); err != nil {
			return nil, fmt.Errorf("parsing frontmatter YAML: %w", err)
		}

		def.SystemPrompt = strings.TrimSpace(body)
	} else if strings.HasSuffix(filePath, ".yaml") || strings.HasSuffix(filePath, ".yml") {
		// Pure YAML file
		decoder := yaml.NewDecoder(strings.NewReader(content))
		decoder.KnownFields(true)
		if err := decoder.Decode(def); err != nil {
			return nil, fmt.Errorf("parsing YAML agent definition: %w", err)
		}
	} else {
		// Plain Markdown — treat entire content as system prompt
		def.SystemPrompt = strings.TrimSpace(content)
	}

	// Default agent type
	if def.AgentType == "" {
		def.AgentType = "claude"
	}

	return def, nil
}

// GetAgentDefinition finds a specific agent definition by name.
func GetAgentDefinition(agentsDir string, name string) (*AgentDefinition, error) {
	defs, err := LoadAgentDefinitions(agentsDir)
	if err != nil {
		return nil, err
	}

	for _, d := range defs {
		if d.Name == name {
			return &d, nil
		}
	}

	return nil, fmt.Errorf("agent definition %q not found in %s", name, agentsDir)
}

// parseFrontmatter splits a document into YAML frontmatter and body.
func parseFrontmatter(content string) (frontmatter, body string, err error) {
	scanner := bufio.NewScanner(strings.NewReader(content))

	// Skip first "---"
	if !scanner.Scan() || strings.TrimSpace(scanner.Text()) != "---" {
		return "", content, nil
	}

	// Read until closing "---"
	var fmLines []string
	foundEnd := false
	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) == "---" {
			foundEnd = true
			break
		}
		fmLines = append(fmLines, line)
	}

	if !foundEnd {
		return "", "", fmt.Errorf("unclosed frontmatter (missing closing ---)")
	}

	frontmatter = strings.Join(fmLines, "\n")

	// Rest is body
	var bodyLines []string
	for scanner.Scan() {
		bodyLines = append(bodyLines, scanner.Text())
	}
	body = strings.Join(bodyLines, "\n")

	return frontmatter, body, nil
}

// fileBaseName returns the filename without extension.
func fileBaseName(path string) string {
	base := filepath.Base(path)
	ext := filepath.Ext(base)
	return strings.TrimSuffix(base, ext)
}
