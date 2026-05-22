package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

const Version = "0.1.0"

var (
	jsonOutput     bool
	noInteractive  bool
	memoryDir      string
)

type Config struct {
	MemoryDir      string
	GlobalConfig   GlobalConfig
	ProjectRoot    string
}

type GlobalConfig struct {
	DefaultMemoryType string `json:"default_memory_type"`
	MaxMemorySize     int    `json:"max_memory_size"`
	AutoIndex         bool   `json:"auto_index"`
}

type Memory struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Type        string    `json:"type"`
	Created     time.Time `json:"created"`
	Content     string    `json:"content"`
}

type ErrorResponse struct {
	Error ErrorDetail `json:"error"`
}

type ErrorDetail struct {
	Code       int    `json:"code"`
	Type       string `json:"type"`
	Message    string `json:"message"`
	Recoverable bool  `json:"recoverable"`
}

type SuccessResponse struct {
	Version string      `json:"version"`
	Data    interface{} `json:"data"`
}

func main() {
	if len(os.Args) < 2 {
		printHelp()
		os.Exit(85) // Invalid argument error code
	}

	command := os.Args[1]

	// Handle help and version before flag parsing
	if command == "--help" || command == "-h" || command == "help" {
		printHelp()
		return
	}
	if command == "--version" || command == "-v" || command == "version" {
		handleVersion()
		return
	}

	// Parse global flags manually to avoid flag package issues
	for i := 2; i < len(os.Args); i++ {
		arg := os.Args[i]
		switch arg {
		case "--json", "-j":
			jsonOutput = true
		case "--no-interactive", "-y":
			noInteractive = true
		case "--memory-dir":
			if i+1 < len(os.Args) {
				memoryDir = os.Args[i+1]
				i++
			}
		}
	}

	// Initialize configuration with centralized storage
	cfg := Config{
		GlobalConfig: loadGlobalConfig(),
	}
	
	// Try to use centralized storage if in a git repository
	if memoryDir != "" {
		// User explicitly set memory directory
		cfg.MemoryDir = memoryDir
	} else {
		// Try git-based scoping
		gitRoot, err := findGitRepositoryRoot()
		if err == nil {
			cfg.ProjectRoot = gitRoot
			cfg.MemoryDir = getProjectMemoryPath(gitRoot)
		} else {
			// Fallback to local directory if not in git
			cfg.MemoryDir = getDefaultMemoryDir()
		}
	}

	// Command routing
	switch command {
	case "init":
		handleInit(&cfg)
	case "remember", "keep":
		handleRemember(&cfg)
	case "recall", "search":
		handleRecall(&cfg)
	case "list", "ls":
		handleList(&cfg)
	case "edit":
		handleEdit(&cfg)
	case "delete":
		handleDelete(&cfg)
	case "status":
		handleStatus(&cfg)
	case "config":
		handleConfig(&cfg)
	case "bridge":
		handleBridge(&cfg)
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", command)
		printHelp()
		os.Exit(85)
	}
}

func getDefaultMemoryDir() string {
	// Default to local .sick-memory directory (fallback)
	return ".sick-memory"
}

func getGlobalSickMemoryDir() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return getDefaultMemoryDir()
	}
	return filepath.Join(homeDir, ".sick-memory")
}

func findGitRepositoryRoot() (string, error) {
	// Try git rev-parse --show-toplevel
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("not in a git repository")
	}
	
	root := strings.TrimSpace(string(output))
	return root, nil
}

func sanitizePath(path string) string {
	// Replace slashes and other problematic characters with dashes
	sanitized := strings.ReplaceAll(path, "/", "-")
	sanitized = strings.ReplaceAll(sanitized, "\\", "-")
	sanitized = strings.ReplaceAll(sanitized, ":", "-")
	sanitized = strings.ReplaceAll(sanitized, " ", "_")
	return sanitized
}

func getProjectMemoryPath(gitRoot string) string {
	globalDir := getGlobalSickMemoryDir()
	sanitizedRoot := sanitizePath(gitRoot)
	return filepath.Join(globalDir, "projects", sanitizedRoot, "memory")
}

func loadGlobalConfig() GlobalConfig {
	globalDir := getGlobalSickMemoryDir()
	configPath := filepath.Join(globalDir, "config.json")
	
	// Default config
	config := GlobalConfig{
		DefaultMemoryType: "user",
		MaxMemorySize:     1024 * 1024, // 1MB
		AutoIndex:         true,
	}
	
	// Try to load existing config
	if data, err := os.ReadFile(configPath); err == nil {
		if err := json.Unmarshal(data, &config); err == nil {
			return config
		}
	}
	
	// Create default config if it doesn't exist
	if err := os.MkdirAll(globalDir, 0755); err == nil {
		configData, _ := json.MarshalIndent(config, "", "  ")
		_ = os.WriteFile(configPath, configData, 0644)
	}
	
	return config
}

func printHelp() {
	fmt.Println("sick-memory - File-based memory system for AI coding agents")
	fmt.Println()
	fmt.Println("USAGE:")
	fmt.Println("    sick-memory [COMMAND] [OPTIONS]")
	fmt.Println()
	fmt.Println("COMMANDS:")
	fmt.Println("    init              Initialize memory system for current project")
	fmt.Println("    remember <content> Add a memory (alias: keep)")
	fmt.Println("    recall [query]    Retrieve relevant memories (alias: search)")
	fmt.Println("    list              List all memories (alias: ls)")
	fmt.Println("    edit <id>         Edit a memory by ID")
	fmt.Println("    delete <id>       Delete a memory by ID")
	fmt.Println("    status            Show system status")
	fmt.Println("    config            Show configuration and storage location")
	fmt.Println("    bridge <agent>    Generate agent-specific integration")
	fmt.Println("    --help            Show this help message")
	fmt.Println("    --version         Show version information")
	fmt.Println()
	fmt.Println("OPTIONS:")
	fmt.Println("    --json, -j            Output in JSON format")
	fmt.Println("    --no-interactive, -y  Disable all prompts")
	fmt.Println("    --memory-dir <dir>    Override memory directory")
	fmt.Println()
	fmt.Println("AGENT BRIDGES:")
	fmt.Println("    bridge claude-code     Generate Claude Code integration")
	fmt.Println("    bridge opencode         Generate OpenCode integration")
	fmt.Println("    bridge copilot          Generate Copilot integration")
	fmt.Println()
	fmt.Println("EXIT CODES:")
	fmt.Println("    0        Success")
	fmt.Println("    1        Generic failure")
	fmt.Println("    80-89    Input/validation errors")
	fmt.Println("    90-99    Resource/state errors")
	fmt.Println("    100-109  Integration/external errors")
	fmt.Println("    110-119  Internal software errors")
}

func handleVersion() {
	fmt.Printf("sick-memory version %s\n", Version)
}

func handleInit(cfg *Config) {
	// Create memory directory
	if err := os.MkdirAll(cfg.MemoryDir, 0755); err != nil {
		errorResponse(92, "resource_error", fmt.Sprintf("Failed to create memory directory: %v", err), false)
		os.Exit(92)
	}

	// Create MEMORY.md index
	indexPath := filepath.Join(cfg.MemoryDir, "MEMORY.md")
	indexContent := "# Memory Index\nThis file contains pointers to all project memories.\nLast updated: " + time.Now().Format(time.RFC3339) + "\n\n"
	if err := os.WriteFile(indexPath, []byte(indexContent), 0644); err != nil {
		errorResponse(92, "resource_error", fmt.Sprintf("Failed to create index file: %v", err), false)
		os.Exit(92)
	}

	if jsonOutput {
		successResponse(map[string]interface{}{
			"status": "initialized",
			"path":   cfg.MemoryDir,
		})
	} else {
		fmt.Printf("Memory system initialized at: %s\n", cfg.MemoryDir)
		fmt.Fprintf(os.Stderr, "Run 'sick-memory remember <content>' to add your first memory.\n")
	}
}

func handleRemember(cfg *Config) {
	// Get content from command line args (skip command and flags)
	args := []string{}
	for i := 2; i < len(os.Args); i++ {
		arg := os.Args[i]
		// Skip flags
		if arg == "--json" || arg == "-j" || arg == "--no-interactive" || arg == "-y" {
			continue
		}
		if arg == "--memory-dir" && i+1 < len(os.Args) {
			i++
			continue
		}
		args = append(args, arg)
	}

	if len(args) == 0 {
		if noInteractive {
			errorResponse(85, "invalid_argument", "Content required for remember command", false)
			os.Exit(85)
		}
		fmt.Fprintf(os.Stderr, "Enter memory content (Ctrl-D to finish):\n")
		// TODO: Implement interactive input
		errorResponse(110, "not_implemented", "Interactive input not yet implemented", false)
		os.Exit(110)
	}

	content := args[0]

	// Create memory file
	timestamp := time.Now().Unix()
	memoryID := fmt.Sprintf("%d", timestamp)
	filename := fmt.Sprintf("memory_%s.md", memoryID)
	filePath := filepath.Join(cfg.MemoryDir, filename)

	memoryContent := fmt.Sprintf(`---
name: Memory %s
description: %s
type: user
created: %d
---

%s
`, memoryID, content, timestamp, content)

	if err := os.WriteFile(filePath, []byte(memoryContent), 0644); err != nil {
		errorResponse(92, "resource_error", fmt.Sprintf("Failed to write memory file: %v", err), false)
		os.Exit(92)
	}

	// Update index
	indexPath := filepath.Join(cfg.MemoryDir, "MEMORY.md")
	indexFile, err := os.OpenFile(indexPath, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		errorResponse(92, "resource_error", fmt.Sprintf("Failed to open index file: %v", err), false)
		os.Exit(92)
	}
	defer indexFile.Close()

	indexEntry := fmt.Sprintf("- [Memory %s](%s) -- %s\n", memoryID, filename, content)
	if _, err := indexFile.WriteString(indexEntry); err != nil {
		errorResponse(92, "resource_error", fmt.Sprintf("Failed to update index: %v", err), false)
		os.Exit(92)
	}

	if jsonOutput {
		successResponse(map[string]interface{}{
			"id":     memoryID,
			"status": "remembered",
		})
	} else {
		fmt.Printf("Memory saved with ID: %s\n", memoryID)
	}
}

func handleRecall(cfg *Config) {
	// TODO: Implement intelligent retrieval
	// For now, list all memories
	memoryPath := cfg.MemoryDir
	
	files, err := os.ReadDir(memoryPath)
	if err != nil {
		errorResponse(92, "resource_not_found", "Memory directory not found", false)
		os.Exit(92)
	}

	memories := []Memory{}
	for _, file := range files {
		if !file.IsDir() && len(file.Name()) > 7 && file.Name()[:7] == "memory_" && file.Name()[len(file.Name())-3:] == ".md" {
			filePath := filepath.Join(memoryPath, file.Name())
			content, err := os.ReadFile(filePath)
			if err != nil {
				continue
			}
			
			memories = append(memories, Memory{
				ID:      file.Name(),
				Content: string(content),
			})
		}
	}

	if jsonOutput {
		successResponse(memories)
	} else {
		fmt.Printf("Memories in %s:\n\n", memoryPath)
		for _, mem := range memories {
			fmt.Printf("ID: %s\n%s\n\n", mem.ID, mem.Content)
		}
		fmt.Printf("Total memories: %d\n", len(memories))
	}
}

func handleList(cfg *Config) {
	memoryPath := cfg.MemoryDir
	
	files, err := os.ReadDir(memoryPath)
	if err != nil {
		if jsonOutput {
			errorResponse(92, "resource_not_found", "Memory directory not found", false)
		} else {
			fmt.Println("Memory directory not found. Run 'sick-memory init' first.")
		}
		os.Exit(92)
	}

	memoryIDs := []string{}
	for _, file := range files {
		if !file.IsDir() && len(file.Name()) > 7 && file.Name()[:7] == "memory_" && file.Name()[len(file.Name())-3:] == ".md" {
			memoryIDs = append(memoryIDs, file.Name())
		}
	}

	if jsonOutput {
		successResponse(memoryIDs)
	} else {
		fmt.Printf("Memories in %s:\n", memoryPath)
		for _, id := range memoryIDs {
			fmt.Printf("  %s\n", id)
		}
		fmt.Printf("\nTotal memories: %d\n", len(memoryIDs))
	}
}

func handleEdit(cfg *Config) {
	if len(os.Args) < 4 {
		errorResponse(80, "missing_argument", "Memory ID and new content required for edit. Usage: sick-memory edit <id> <new content>", false)
		os.Exit(80)
	}

	memoryID := os.Args[2]
	newContent := strings.Join(os.Args[3:], " ")
	
	// Find the memory file
	memoryPath := cfg.MemoryDir
	var memoryFile string
	
	// Try to find the memory file by ID
	files, err := os.ReadDir(memoryPath)
	if err != nil {
		errorResponse(92, "memory_not_found", fmt.Sprintf("Cannot read memory directory: %v", err), false)
		os.Exit(92)
	}
	
	for _, file := range files {
		// Match either exact ID or ID with memory_ prefix
		fileName := strings.TrimSuffix(file.Name(), ".md")
		if fileName == memoryID || fileName == "memory_"+memoryID || strings.HasSuffix(fileName, "_"+memoryID) {
			memoryFile = filepath.Join(memoryPath, file.Name())
			break
		}
	}
	
	if memoryFile == "" {
		errorResponse(92, "memory_not_found", fmt.Sprintf("Memory with ID %s not found", memoryID), false)
		os.Exit(92)
	}
	
	// Read existing memory
	content, err := os.ReadFile(memoryFile)
	if err != nil {
		errorResponse(110, "read_error", fmt.Sprintf("Cannot read memory file: %v", err), false)
		os.Exit(110)
	}
	
	// Parse the existing memory to separate frontmatter from content
	lines := strings.Split(string(content), "\n")
	var frontmatter []string
	inFrontmatter := false
	
	for _, line := range lines {
		if line == "---" {
			if !inFrontmatter {
				inFrontmatter = true
				frontmatter = append(frontmatter, line)
				continue
			} else {
				inFrontmatter = false
				frontmatter = append(frontmatter, line)
				continue
			}
		}
		
		if inFrontmatter {
			frontmatter = append(frontmatter, line)
		}
	}
	
	// Update description based on new content (first line or first 50 chars)
	description := newContent
	if len(description) > 50 {
		description = description[:50] + "..."
	}
	if idx := strings.Index(description, "\n"); idx != -1 && idx < 50 {
		description = description[:idx]
	}
	
	// Update frontmatter with new description
	updatedFrontmatter := make([]string, 0, len(frontmatter))
	for i, line := range frontmatter {
		if strings.HasPrefix(line, "description:") {
			updatedFrontmatter = append(updatedFrontmatter, fmt.Sprintf("description: %s", description))
		} else if i > 0 && strings.HasPrefix(frontmatter[i-1], "description:") {
			// Skip continuation lines
			continue
		} else {
			updatedFrontmatter = append(updatedFrontmatter, line)
		}
	}
	
	// Reconstruct the file
	var updatedContent strings.Builder
	for _, line := range updatedFrontmatter {
		updatedContent.WriteString(line + "\n")
	}
	updatedContent.WriteString("\n")
	updatedContent.WriteString(newContent)
	
	// Write back to file
	if err := os.WriteFile(memoryFile, []byte(updatedContent.String()), 0644); err != nil {
		errorResponse(110, "write_error", fmt.Sprintf("Cannot write memory file: %v", err), false)
		os.Exit(110)
	}
	
	response := map[string]interface{}{
		"version": Version,
		"data": map[string]interface{}{
			"id":          memoryID,
			"status":      "updated",
			"description": description,
		},
	}
	
	if jsonOutput {
		jsonData, _ := json.MarshalIndent(response, "", "  ")
		fmt.Println(string(jsonData))
	} else {
		fmt.Printf("Memory %s updated successfully\n", memoryID)
	}
}

func handleDelete(cfg *Config) {
	if len(os.Args) < 3 {
		errorResponse(80, "missing_argument", "Memory ID required for delete", false)
		os.Exit(80)
	}

	memoryID := os.Args[2]
	
	// Find the memory file
	memoryPath := cfg.MemoryDir
	var memoryFile string
	
	// Try to find the memory file by ID
	files, err := os.ReadDir(memoryPath)
	if err != nil {
		errorResponse(92, "memory_not_found", fmt.Sprintf("Cannot read memory directory: %v", err), false)
		os.Exit(92)
	}
	
	for _, file := range files {
		// Match either exact ID or ID with memory_ prefix
		fileName := strings.TrimSuffix(file.Name(), ".md")
		if fileName == memoryID || fileName == "memory_"+memoryID || strings.HasSuffix(fileName, "_"+memoryID) {
			memoryFile = filepath.Join(memoryPath, file.Name())
			break
		}
	}
	
	if memoryFile == "" {
		errorResponse(92, "memory_not_found", fmt.Sprintf("Memory with ID %s not found", memoryID), false)
		os.Exit(92)
	}
	
	// Delete the memory file
	if err := os.Remove(memoryFile); err != nil {
		errorResponse(110, "delete_error", fmt.Sprintf("Cannot delete memory file: %v", err), false)
		os.Exit(110)
	}
	
	if jsonOutput {
		response := map[string]interface{}{
			"version": Version,
			"data": map[string]interface{}{
				"id":     memoryID,
				"status": "deleted",
			},
		}
		jsonData, _ := json.MarshalIndent(response, "", "  ")
		fmt.Println(string(jsonData))
	} else {
		fmt.Printf("Memory %s deleted successfully\n", memoryID)
	}
}

func handleStatus(cfg *Config) {
	memoryPath := cfg.MemoryDir
	
	if _, err := os.Stat(memoryPath); os.IsNotExist(err) {
		if jsonOutput {
			successResponse(map[string]interface{}{
				"status": "uninitialized",
			})
		} else {
			fmt.Println("Memory system status: uninitialized")
			fmt.Println("Run 'sick-memory init' to initialize.")
		}
		return
	}

	files, err := os.ReadDir(memoryPath)
	if err != nil {
		errorResponse(92, "resource_error", fmt.Sprintf("Failed to read memory directory: %v", err), false)
		os.Exit(92)
	}

	memoryCount := 0
	for _, file := range files {
		if !file.IsDir() && len(file.Name()) > 7 && file.Name()[:7] == "memory_" && file.Name()[len(file.Name())-3:] == ".md" {
			memoryCount++
		}
	}

	if jsonOutput {
		successResponse(map[string]interface{}{
			"status": "active",
			"path":   memoryPath,
			"count":  memoryCount,
		})
	} else {
		fmt.Println("Memory system status: active")
		fmt.Printf("Memory directory: %s\n", memoryPath)
		fmt.Printf("Total memories: %d\n", memoryCount)
	}
}

func handleConfig(cfg *Config) {
	globalDir := getGlobalSickMemoryDir()
	
	if jsonOutput {
		configData := map[string]interface{}{
			"global_directory": globalDir,
			"memory_directory": cfg.MemoryDir,
			"project_root":     cfg.ProjectRoot,
			"global_config":    cfg.GlobalConfig,
		}
		successResponse(configData)
	} else {
		fmt.Println("Sick-Memory Configuration:")
		fmt.Println()
		fmt.Printf("Global Directory: %s\n", globalDir)
		fmt.Printf("Memory Directory: %s\n", cfg.MemoryDir)
		
		if cfg.ProjectRoot != "" {
			fmt.Printf("Project Root: %s\n", cfg.ProjectRoot)
			fmt.Println("Storage Mode: Centralized (git-based scoping)")
		} else {
			fmt.Println("Project Root: Not in a git repository")
			fmt.Println("Storage Mode: Local (fallback)")
		}
		
		fmt.Println()
		fmt.Println("Global Configuration:")
		fmt.Printf("  Default Memory Type: %s\n", cfg.GlobalConfig.DefaultMemoryType)
		fmt.Printf("  Max Memory Size: %d bytes\n", cfg.GlobalConfig.MaxMemorySize)
		fmt.Printf("  Auto Index: %v\n", cfg.GlobalConfig.AutoIndex)
		
		fmt.Println()
		fmt.Println("Configuration File:")
		fmt.Printf("  %s\n", filepath.Join(globalDir, "config.json"))
	}
}

func handleBridge(cfg *Config) {
	// Get agent name from command line args (skip command and flags)
	args := []string{}
	for i := 2; i < len(os.Args); i++ {
		arg := os.Args[i]
		// Skip flags
		if arg == "--json" || arg == "-j" || arg == "--no-interactive" || arg == "-y" {
			continue
		}
		if arg == "--memory-dir" && i+1 < len(os.Args) {
			i++
			continue
		}
		args = append(args, arg)
	}

	if len(args) == 0 {
		fmt.Fprintf(os.Stderr, "Usage: sick-memory bridge <agent-name>\n")
		fmt.Fprintf(os.Stderr, "Available agents: claude-code, opencode, copilot\n")
		os.Exit(85)
	}

	agent := args[0]

	switch agent {
	case "claude-code":
		generateClaudeCodeBridge(cfg)
	case "opencode":
		generateOpenCodeBridge(cfg)
	case "copilot":
		generateCopilotBridge(cfg)
	default:
		fmt.Fprintf(os.Stderr, "Unknown agent: %s\n", agent)
		fmt.Fprintf(os.Stderr, "Available agents: claude-code, opencode, copilot\n")
		os.Exit(85)
	}
}

func generateClaudeCodeBridge(cfg *Config) {
	claudeMDContent := fmt.Sprintf(`# Sick-Memory Integration for Claude Code

This file provides project-specific memory for Claude Code via sick-memory CLI.

## Memory Loading

To load memories at session start, add this to your Claude Code workflow:

%%bash
# Load relevant memories
sick-memory recall --json
%%

## Adding Memories

%%bash
# Add a memory
sick-memory remember "Use real database instances in tests, not mocks"

# Add with type
sick-memory remember --type feedback "Integration tests must hit real DB"
%%

## Memory Location

Memories are stored in: %s

## Storage Mode

This project uses centralized storage with git-based scoping.
All git worktrees of this repository share the same memory directory.

## Bridge Commands

The following bridge commands are available:
- /sm - Access sick-memory functionality
- /sm remember <content> - Add a memory
- /sm recall [query] - Retrieve memories
- /sm status - Check memory system status
- /sm config - Show configuration and storage location
`, cfg.MemoryDir)

	// Write CLAUDE.md file
	claudeMDPath := ".claude/CLAUDE.md"
	if err := os.MkdirAll(".claude", 0755); err != nil {
		errorResponse(92, "resource_error", fmt.Sprintf("Failed to create .claude directory: %v", err), false)
		os.Exit(92)
	}

	if err := os.WriteFile(claudeMDPath, []byte(claudeMDContent), 0644); err != nil {
		errorResponse(92, "resource_error", fmt.Sprintf("Failed to write CLAUDE.md: %v", err), false)
		os.Exit(92)
	}

	if jsonOutput {
		successResponse(map[string]interface{}{
			"status":      "bridge_created",
			"agent":       "claude-code",
			"config_file": claudeMDPath,
		})
	} else {
		fmt.Println("Claude Code bridge created successfully!")
		fmt.Printf("Configuration file: %s\n", claudeMDPath)
		fmt.Fprintf(os.Stderr, "Add the bridge commands to your Claude Code workflow.\n")
	}
}

func generateOpenCodeBridge(cfg *Config) {
	opencodeConfig := fmt.Sprintf(`# Sick-Memory Integration for OpenCode

## Configuration

Add this to your OpenCode configuration:

%%json
{
  "memory": {
    "enabled": true,
    "command": "sick-memory",
    "path": "%s"
  }
}
%%

## Usage

%%bash
# Load memories
sick-memory recall --json

# Add memory
sick-memory remember "Project-specific context"
%%
`, cfg.MemoryDir)

	configPath := ".opencode/memory.json"
	if err := os.MkdirAll(".opencode", 0755); err != nil {
		errorResponse(92, "resource_error", fmt.Sprintf("Failed to create .opencode directory: %v", err), false)
		os.Exit(92)
	}

	if err := os.WriteFile(configPath, []byte(opencodeConfig), 0644); err != nil {
		errorResponse(92, "resource_error", fmt.Sprintf("Failed to write config: %v", err), false)
		os.Exit(92)
	}

	if jsonOutput {
		successResponse(map[string]interface{}{
			"status":      "bridge_created",
			"agent":       "opencode",
			"config_file": configPath,
		})
	} else {
		fmt.Println("OpenCode bridge created successfully!")
		fmt.Printf("Configuration file: %s\n", configPath)
		fmt.Fprintf(os.Stderr, "Add the configuration to your OpenCode setup.\n")
	}
}

func generateCopilotBridge(cfg *Config) {
	copilotConfig := fmt.Sprintf(`# Sick-Memory Integration for GitHub Copilot

## Configuration

Add this to your .copilot/settings.json:

%%json
{
  "memory": {
    "enabled": true,
    "command": "sick-memory recall",
    "path": "%s"
  }
}
%%

## Usage

%%bash
# Load memories before starting work
sick-memory recall --json

# Add context during work
sick-memory remember "Important project context"
%%
`, cfg.MemoryDir)

	configPath := ".copilot/settings.json"
	if err := os.MkdirAll(".copilot", 0755); err != nil {
		errorResponse(92, "resource_error", fmt.Sprintf("Failed to create .copilot directory: %v", err), false)
		os.Exit(92)
	}

	if err := os.WriteFile(configPath, []byte(copilotConfig), 0644); err != nil {
		errorResponse(92, "resource_error", fmt.Sprintf("Failed to write config: %v", err), false)
		os.Exit(92)
	}

	if jsonOutput {
		successResponse(map[string]interface{}{
			"status":      "bridge_created",
			"agent":       "copilot",
			"config_file": configPath,
		})
	} else {
		fmt.Println("Copilot bridge created successfully!")
		fmt.Printf("Configuration file: %s\n", configPath)
		fmt.Fprintf(os.Stderr, "Add the configuration to your Copilot settings.\n")
	}
}

func successResponse(data interface{}) {
	response := SuccessResponse{
		Version: "1.0",
		Data:    data,
	}
	jsonBytes, err := json.Marshal(response)
	if err != nil {
		fmt.Printf(`{"error":{"code":110,"type":"json_error","message":"Failed to marshal response"}}\n`)
		os.Exit(110)
	}
	fmt.Println(string(jsonBytes))
}

func errorResponse(code int, errorType string, message string, recoverable bool) {
	response := ErrorResponse{
		Error: ErrorDetail{
			Code:        code,
			Type:        errorType,
			Message:     message,
			Recoverable: recoverable,
		},
	}
	jsonBytes, err := json.Marshal(response)
	if err != nil {
		fmt.Printf(`{"error":{"code":110,"type":"json_error","message":"Failed to marshal error response"}}\n`)
		os.Exit(110)
	}
	fmt.Println(string(jsonBytes))
}