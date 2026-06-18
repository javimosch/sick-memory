package main

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// Version is the current release version of sick-memory.
const Version = "0.1.0"

// Common stop words for keyword extraction
var stopWords = map[string]bool{
	"a": true, "an": true, "and": true, "are": true, "as": true, "at": true, "be": true, "by": true, "for": true, "from": true,
	"has": true, "have": true, "he": true, "in": true, "is": true, "it": true, "its": true, "of": true, "on": true, "that": true,
	"the": true, "to": true, "was": true, "were": true, "will": true, "with": true, "this": true, "but": true,
	"they": true, "or": true, "one": true, "had": true, "word": true, "not": true,
	"what": true, "all": true, "we": true, "when": true, "your": true, "can": true, "said": true, "there": true,
	"use": true, "each": true, "which": true, "she": true, "do": true, "how": true, "their": true, "if": true,
	"up": true, "other": true, "about": true, "out": true, "many": true, "then": true, "them": true, "these": true,
	"so": true, "some": true, "her": true, "would": true, "make": true, "like": true, "him": true, "into": true, "time": true,
	"look": true, "two": true, "more": true, "write": true, "go": true, "see": true, "number": true, "no": true,
	"way": true, "could": true, "people": true, "my": true, "than": true, "first": true, "been": true, "call": true,
	"who": true, "oil": true, "sit": true, "now": true, "find": true, "down": true, "day": true, "did": true, "get": true,
	"come": true, "made": true, "may": true, "part": true,
}

var (
	jsonOutput     bool
	noInteractive  bool
	memoryDir      string
)

// Config holds the runtime configuration for sick-memory including
// memory directory paths, global settings, and the current project root.
type Config struct {
	MemoryDir      string
	GlobalConfig   GlobalConfig
	ProjectRoot    string
}

// GlobalConfig stores user preferences loaded from ~/.sick-memory/config.json.
type GlobalConfig struct {
	DefaultMemoryType string `json:"default_memory_type"`
	MaxMemorySize     int    `json:"max_memory_size"`
	AutoIndex         bool   `json:"auto_index"`
}

// Memory represents a single stored memory with metadata and content.
type Memory struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Type        string    `json:"type"`
	Created     time.Time `json:"created"`
	Content     string    `json:"content"`
}

// SearchIndex is the TF-IDF based search index used for semantic memory retrieval.
type SearchIndex struct {
	TermFreq   map[string]map[string]int // term -> (memoryID -> frequency)
	DocFreq    map[string]int             // term -> document frequency
	DocCount   int                        // total number of documents
	Memories   map[string]Memory          // memoryID -> Memory metadata
}

// SearchResult holds a single search result with its relevance score.
type SearchResult struct {
	MemoryID    string  `json:"memory_id"`
	Score       float64 `json:"score"`
	Title       string  `json:"title"`
	Content     string  `json:"content"`
	MemoryType  string  `json:"memory_type"`
	Created     string  `json:"created"`
}

// ErrorResponse wraps an error detail for JSON error output.
type ErrorResponse struct {
	Error ErrorDetail `json:"error"`
}

// ErrorDetail describes a structured error with type classification and recovery hint.
type ErrorDetail struct {
	Code       int    `json:"code"`
	Type       string `json:"type"`
	Message    string `json:"message"`
	Recoverable bool  `json:"recoverable"`
}

// SuccessResponse wraps successful command output with a version field.
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

// getDefaultMemoryDir returns the fallback local memory directory path
// used when not in a git repository or when centralized storage is unavailable.
func getDefaultMemoryDir() string {
	// Default to local .sick-memory directory (fallback)
	return ".sick-memory"
}

// getGlobalSickMemoryDir returns the centralized sick-memory directory path
// in the user's home directory (~/.sick-memory).
func getGlobalSickMemoryDir() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return getDefaultMemoryDir()
	}
	return filepath.Join(homeDir, ".sick-memory")
}

// findGitRepositoryRoot uses git rev-parse to find the root directory
// of the current git repository. Returns an error if not in a git repository.
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

// sanitizePath replaces filesystem-unsafe characters with safe alternatives
// to create valid directory names from git repository paths.
func sanitizePath(path string) string {
	// Replace slashes and other problematic characters with dashes
	sanitized := strings.ReplaceAll(path, "/", "-")
	sanitized = strings.ReplaceAll(sanitized, "\\", "-")
	sanitized = strings.ReplaceAll(sanitized, ":", "-")
	sanitized = strings.ReplaceAll(sanitized, " ", "_")
	return sanitized
}

// getProjectMemoryPath constructs the centralized memory storage path
// for a specific git repository, using sanitized repository root as directory name.
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

// Search functions for pseudo semantic search

func extractKeywords(text string) []string {
	// Convert to lowercase and split into words
	text = strings.ToLower(text)
	words := strings.Fields(text)
	
	var keywords []string
	for _, word := range words {
		// Remove punctuation
		word = strings.Trim(word, ".,!?;:\"'()[]{}")
		if len(word) > 2 && !stopWords[word] {
			keywords = append(keywords, word)
		}
	}
	return keywords
}

func buildSearchIndex(memoryPath string) (*SearchIndex, error) {
	index := &SearchIndex{
		TermFreq: make(map[string]map[string]int),
		DocFreq:  make(map[string]int),
		Memories: make(map[string]Memory),
	}
	
	files, err := os.ReadDir(memoryPath)
	if err != nil {
		return nil, err
	}
	
	for _, file := range files {
		if file.IsDir() || strings.HasPrefix(file.Name(), ".") {
			continue
		}

		// Only process memory files (memory_*.md pattern)
		if !strings.HasPrefix(file.Name(), "memory_") || !strings.HasSuffix(file.Name(), ".md") {
			continue
		}

		// Read memory file
		filePath := filepath.Join(memoryPath, file.Name())
		content, err := os.ReadFile(filePath)
		if err != nil {
			continue
		}

		// Parse memory
		memory := parseMemory(string(content), file.Name())
		if memory.ID == "" {
			continue
		}
		
		// Extract keywords from content and description
		fullText := memory.Description + " " + memory.Content
		keywords := extractKeywords(fullText)
		
		// Update term frequencies
		seen := make(map[string]bool)
		for _, keyword := range keywords {
			if index.TermFreq[keyword] == nil {
				index.TermFreq[keyword] = make(map[string]int)
			}
			index.TermFreq[keyword][memory.ID]++
			if !seen[keyword] {
				index.DocFreq[keyword]++
				seen[keyword] = true
			}
		}
		
		// Store memory metadata
		index.Memories[memory.ID] = memory
	}
	
	index.DocCount = len(index.Memories)
	return index, nil
}

func parseMemory(content, filename string) Memory {
	lines := strings.Split(content, "\n")
	var memory Memory
	memory.ID = strings.TrimSuffix(filename, ".md")
	
	inFrontmatter := false
	var contentLines []string
	
	for _, line := range lines {
		if line == "---" {
			if !inFrontmatter {
				inFrontmatter = true
				continue
			} else {
				inFrontmatter = false
				continue
			}
		}
		
		if inFrontmatter {
			if strings.HasPrefix(line, "name:") {
				memory.Name = strings.TrimSpace(strings.TrimPrefix(line, "name:"))
			} else if strings.HasPrefix(line, "description:") {
				memory.Description = strings.TrimSpace(strings.TrimPrefix(line, "description:"))
			} else if strings.HasPrefix(line, "type:") {
				memory.Type = strings.TrimSpace(strings.TrimPrefix(line, "type:"))
			} else if strings.HasPrefix(line, "created:") {
				if timestamp, err := time.Parse(strings.Trim(strings.TrimPrefix(line, "created:"), " "), time.RFC3339); err == nil {
					memory.Created = timestamp
				}
			}
		} else {
			contentLines = append(contentLines, line)
		}
	}
	
	memory.Content = strings.Join(contentLines, "\n")
	return memory
}

func calculateTFIDF(index *SearchIndex, term string, memoryID string) float64 {
	if index.DocCount == 0 {
		return 0
	}
	
	// Term frequency in this document
	tf := float64(index.TermFreq[term][memoryID])
	
	// Document frequency (how many docs contain this term)
	df := float64(index.DocFreq[term])
	
	// IDF calculation
	idf := math.Log(float64(index.DocCount+1) / (df + 1))
	
	return tf * idf
}

func searchMemories(index *SearchIndex, query string) []SearchResult {
	queryKeywords := extractKeywords(query)
	results := make([]SearchResult, 0)
	
	for memoryID, memory := range index.Memories {
		score := 0.0
		
		// Calculate TF-IDF score for matching terms
		for _, keyword := range queryKeywords {
			if index.TermFreq[keyword] != nil && index.TermFreq[keyword][memoryID] > 0 {
				tfidf := calculateTFIDF(index, keyword, memoryID)
				score += tfidf
			}
		}
		
		// Boost score for exact phrase matches
		lowerContent := strings.ToLower(memory.Content + " " + memory.Description)
		lowerQuery := strings.ToLower(query)
		if strings.Contains(lowerContent, lowerQuery) {
			score += 2.0 // Boost for exact phrase match
		}
		
		// Word-overlap fallback: when TF-IDF and exact phrase both give 0,
		// score by how many individual query keywords exist as substrings.
		// Handles cases like "UI design" matching "UI/Design" or
		// "disk cleanup" matching "disk-cleanup".
		if score == 0 && len(queryKeywords) > 0 {
			matchedWords := 0
			for _, kw := range queryKeywords {
				if strings.Contains(lowerContent, kw) {
					matchedWords++
				}
			}
			if matchedWords > 0 {
				score = (float64(matchedWords) / float64(len(queryKeywords))) * 2.0
			}
		}
		
		// Recency boost (more recent = higher score)
		hoursSinceCreation := time.Since(memory.Created).Hours()
		if hoursSinceCreation < 24 {
			score *= 1.2 // Boost for memories < 24h old
		} else if hoursSinceCreation < 168 {
			score *= 1.1 // Boost for memories < 7 days old
		}
		
		// Type-based boosting (project memories are more relevant for context)
		if memory.Type == "project" {
			score *= 1.15
		}
		
		if score > 0 {
			results = append(results, SearchResult{
				MemoryID:   memoryID,
				Score:      score,
				Title:      memory.Name,
				Content:    memory.Content,
				MemoryType: memory.Type,
				Created:    memory.Created.Format(time.RFC3339),
			})
		}
	}
	
	// Sort by score (descending)
	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})
	
	return results
}

func loadSearchIndex(memoryPath string) (*SearchIndex, error) {
	// Try to load cached index
	indexFile := filepath.Join(memoryPath, "search_index.json")
	if data, err := os.ReadFile(indexFile); err == nil {
		var index SearchIndex
		if err := json.Unmarshal(data, &index); err == nil {
			return &index, nil
		}
	}
	
	// Build index from scratch and cache it
	index, err := buildSearchIndex(memoryPath)
	if err == nil {
		saveSearchIndex(memoryPath, index)
	}
	return index, err
}

func saveSearchIndex(memoryPath string, index *SearchIndex) error {
	indexFile := filepath.Join(memoryPath, "search_index.json")
	data, err := json.MarshalIndent(index, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(indexFile, data, 0644)
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

	// Update search index
	if cfg.GlobalConfig.AutoIndex {
		// Rebuild search index for semantic search
		index, err := buildSearchIndex(cfg.MemoryDir)
		if err == nil {
			saveSearchIndex(cfg.MemoryDir, index)
		}
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
	query := ""
	if len(os.Args) > 2 {
		queryParts := []string{}
		for _, arg := range os.Args[2:] {
			if arg == "--json" || arg == "-j" || arg == "--no-interactive" || arg == "-y" || strings.HasPrefix(arg, "--memory-dir") {
				continue
			}
			queryParts = append(queryParts, arg)
		}
		query = strings.Join(queryParts, " ")
	}

	// Load or build search index
	index, err := loadSearchIndex(cfg.MemoryDir)
	if err != nil {
		errorResponse(92, "resource_not_found", "Memory directory not found", false)
		os.Exit(92)
	}
	
	if query == "" {
		// If no query, return all memories sorted by recency
		var allMemories []SearchResult
		for _, memory := range index.Memories {
			hoursSinceCreation := time.Since(memory.Created).Hours()
			score := 100.0 - hoursSinceCreation // Newer = higher score
			allMemories = append(allMemories, SearchResult{
				MemoryID:   memory.ID,
				Score:      score,
				Title:      memory.Name,
				Content:    memory.Content,
				MemoryType: memory.Type,
				Created:    memory.Created.Format(time.RFC3339),
			})
		}
		
		// Sort by score (newest first)
		sort.Slice(allMemories, func(i, j int) bool {
			return allMemories[i].Score > allMemories[j].Score
		})
		
		if jsonOutput {
			successResponse(allMemories)
		} else {
			fmt.Printf("All memories in %s:\n\n", cfg.MemoryDir)
			for _, result := range allMemories {
				fmt.Printf("ID: %s\n---\n%s\n\n", result.MemoryID, result.Content)
			}
			fmt.Printf("Total memories: %d\n", len(allMemories))
		}
		return
	}
	
	// Perform semantic search
	results := searchMemories(index, query)
	
	if len(results) == 0 {
		if jsonOutput {
			successResponse([]SearchResult{})
		} else {
			fmt.Printf("No memories found matching: %s\n", query)
		}
		return
	}
	
	if jsonOutput {
		successResponse(results)
	} else {
		fmt.Printf("Found %d memories matching: %s\n\n", len(results), query)
		for _, result := range results {
			fmt.Printf("ID: %s (score: %.2f)\n", result.MemoryID, result.Score)
			fmt.Printf("Type: %s\n", result.MemoryType)
			fmt.Printf("Created: %s\n", result.Created)
			fmt.Printf("%s\n\n", result.Content)
		}
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