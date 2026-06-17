package main

import (
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// ---------------------------------------------------------------------------
// extractKeywords tests
// ---------------------------------------------------------------------------

func TestExtractKeywords_RemovesStopWords(t *testing.T) {
	t.Parallel()
	keywords := extractKeywords("the and of to a in is it")
	for _, kw := range keywords {
		if kw == "the" || kw == "and" || kw == "of" {
			t.Errorf("stop word '%s' should not appear in keywords", kw)
		}
	}
}

func TestExtractKeywords_ReturnsMeaningfulWords(t *testing.T) {
	t.Parallel()
	keywords := extractKeywords("memory search function integration test")
	expected := map[string]bool{"memory": true, "search": true, "function": true, "integration": true, "test": true}
	for _, kw := range keywords {
		if !expected[kw] {
			t.Errorf("unexpected keyword: %s", kw)
		}
	}
	if len(keywords) != len(expected) {
		t.Errorf("expected %d keywords, got %d: %v", len(expected), len(keywords), keywords)
	}
}

func TestExtractKeywords_RemovesPunctuation(t *testing.T) {
	t.Parallel()
	keywords := extractKeywords("hello, world! test: ok?")
	for _, kw := range keywords {
		if strings.ContainsAny(kw, ".,!?;:\"'()[]{}") {
			t.Errorf("keyword '%s' still contains punctuation", kw)
		}
	}
}

func TestExtractKeywords_RemovesShortWords(t *testing.T) {
	t.Parallel()
	keywords := extractKeywords("a an by at of in hi ok go")
	for _, kw := range keywords {
		if len(kw) <= 2 {
			t.Errorf("short word '%s' should be removed", kw)
		}
	}
}

func TestExtractKeywords_EmptyInput(t *testing.T) {
	t.Parallel()
	keywords := extractKeywords("")
	if len(keywords) != 0 {
		t.Errorf("expected empty result for empty input, got %v", keywords)
	}
}

func TestExtractKeywords_OnlyStopWords(t *testing.T) {
	t.Parallel()
	keywords := extractKeywords("the and of to")
	if len(keywords) != 0 {
		t.Errorf("expected empty result for only stop words, got %v", keywords)
	}
}

func TestExtractKeywords_CaseInsensitive(t *testing.T) {
	t.Parallel()
	keywords := extractKeywords("Memory SEARCH Function")
	expected := []string{"memory", "search", "function"}
	if len(keywords) != len(expected) {
		t.Fatalf("expected %d keywords, got %d: %v", len(expected), len(keywords), keywords)
	}
	for i, kw := range keywords {
		if kw != expected[i] {
			t.Errorf("keywords[%d] = %s, want %s", i, kw, expected[i])
		}
	}
}

// ---------------------------------------------------------------------------
// sanitizePath tests
// ---------------------------------------------------------------------------

func TestSanitizePath_ReplacesSlashes(t *testing.T) {
	t.Parallel()
	result := sanitizePath("/home/user/project")
	if strings.Contains(result, "/") {
		t.Errorf("path should not contain slashes: %s", result)
	}
}

func TestSanitizePath_ReplacesBackslashes(t *testing.T) {
	t.Parallel()
	result := sanitizePath(`C:\Users\test`)
	if strings.Contains(result, "\\") {
		t.Errorf("path should not contain backslashes: %s", result)
	}
}

func TestSanitizePath_ReplacesColons(t *testing.T) {
	t.Parallel()
	result := sanitizePath("C:Project")
	if strings.Contains(result, ":") {
		t.Errorf("path should not contain colons: %s", result)
	}
}

func TestSanitizePath_ReplacesSpaces(t *testing.T) {
	t.Parallel()
	result := sanitizePath("my project folder")
	if strings.Contains(result, " ") {
		t.Errorf("path should not contain spaces: %s", result)
	}
	if !strings.Contains(result, "_") {
		t.Errorf("path should contain underscores replacing spaces: %s", result)
	}
}

func TestSanitizePath_AllReplacements(t *testing.T) {
	t.Parallel()
	result := sanitizePath("/a/b/c: d\\e")
	// / -> -, : -> -, \ -> -, space -> _
	expected := "-a-b-c-_d-e"
	if result != expected {
		t.Errorf("sanitizePath result = %q, want %q", result, expected)
	}
}

// ---------------------------------------------------------------------------
// calculateTFIDF tests
// ---------------------------------------------------------------------------

func TestCalculateTFIDF_ZeroDocCount(t *testing.T) {
	t.Parallel()
	idx := &SearchIndex{
		TermFreq: map[string]map[string]int{"go": {"mem1": 3}},
		DocFreq:  map[string]int{"go": 1},
		DocCount: 0,
	}
	score := calculateTFIDF(idx, "go", "mem1")
	if score != 0 {
		t.Errorf("expected 0 for zero DocCount, got %f", score)
	}
}

func TestCalculateTFIDF_TermNotFound(t *testing.T) {
	t.Parallel()
	idx := &SearchIndex{
		TermFreq: map[string]map[string]int{},
		DocFreq:  map[string]int{},
		DocCount: 5,
	}
	score := calculateTFIDF(idx, "nonexistent", "mem1")
	if score != 0 {
		t.Errorf("expected 0 for missing term, got %f", score)
	}
}

func TestCalculateTFIDF_KnownValue(t *testing.T) {
	t.Parallel()
	// DocCount=10, term appears in 2 docs, frequency in this doc = 3
	// tf=3, df=2, idf=log((10+1)/(2+1))=log(11/3)=log(3.666...)
	// tf * idf = 3 * log(11/3)
	idx := &SearchIndex{
		TermFreq: map[string]map[string]int{"test": {"mem1": 3, "mem2": 1}},
		DocFreq:  map[string]int{"test": 2},
		DocCount: 10,
	}
	score := calculateTFIDF(idx, "test", "mem1")
	expected := 3.0 * math.Log(11.0/3.0)
	if math.Abs(score-expected) > 1e-9 {
		t.Errorf("score = %f, want %f", score, expected)
	}
}

// ---------------------------------------------------------------------------
// searchMemories tests
// ---------------------------------------------------------------------------

func TestSearchMemories_EmptyIndex(t *testing.T) {
	t.Parallel()
	idx := &SearchIndex{
		TermFreq: make(map[string]map[string]int),
		DocFreq:  make(map[string]int),
		DocCount: 0,
		Memories: make(map[string]Memory),
	}
	results := searchMemories(idx, "anything")
	if len(results) != 0 {
		t.Errorf("expected 0 results for empty index, got %d", len(results))
	}
}

func TestSearchMemories_ExactMatchBoost(t *testing.T) {
	t.Parallel()
	now := time.Now()
	mem := Memory{
		ID:          "mem1",
		Name:        "Test",
		Description: "",
		Content:     "the quick brown fox jumps over the lazy dog",
		Created:     now,
	}
	idx := &SearchIndex{
		TermFreq: map[string]map[string]int{
			"quick": {"mem1": 1},
			"brown": {"mem1": 1},
			"fox":   {"mem1": 1},
		},
		DocFreq: map[string]int{
			"quick": 1,
			"brown": 1,
			"fox":   1,
		},
		DocCount: 1,
		Memories: map[string]Memory{"mem1": mem},
	}

	results := searchMemories(idx, "quick brown fox")
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d: %+v", len(results), results)
	}
	// Should get a boost for exact phrase match
	// TF-IDF contribution is 0 because term appears in every doc (docFreq==docCount, log(2/2)=0)
	// Score should be 2.0 (exact match boost) * 1.2 (recency < 24h)
	if results[0].Score < 2.0 {
		t.Errorf("expected score >= 2.0 due to exact phrase boost, got %f", results[0].Score)
	}
}

func TestSearchMemories_WordOverlapFallback(t *testing.T) {
	t.Parallel()
	now := time.Now()
	mem := Memory{
		ID:          "mem1",
		Name:        "Test",
		Description: "",
		Content:     "disk-cleanup utility",
		Created:     now,
	}
	idx := &SearchIndex{
		TermFreq: map[string]map[string]int{},
		DocFreq:  map[string]int{},
		DocCount: 1,
		Memories: map[string]Memory{"mem1": mem},
	}

	// Query "disk cleanup" should fallback to word-overlap matching "disk" in "disk-cleanup"
	results := searchMemories(idx, "disk cleanup")
	if len(results) != 1 {
		t.Fatalf("expected 1 result from word-overlap fallback, got %d", len(results))
	}
	if results[0].Score <= 0 {
		t.Errorf("expected positive score from word-overlap, got %f", results[0].Score)
	}
}

func TestSearchMemories_SortsByScoreDescending(t *testing.T) {
	t.Parallel()
	now := time.Now()
	mem1 := Memory{
		ID:      "mem1",
		Content: "the quick brown fox",
		Created: now,
	}
	mem2 := Memory{
		ID:      "mem2",
		Content: "the brown bear",
		Created: now,
	}
	idx := &SearchIndex{
		TermFreq: map[string]map[string]int{
			"brown": {"mem1": 1, "mem2": 1},
			"quick": {"mem1": 1},
			"fox":   {"mem1": 1},
		},
		DocFreq: map[string]int{
			"brown": 2,
			"quick": 1,
			"fox":   1,
		},
		DocCount: 2,
		Memories: map[string]Memory{"mem1": mem1, "mem2": mem2},
	}
	results := searchMemories(idx, "brown")
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
	if results[0].Score < results[1].Score {
		t.Errorf("results not sorted descending by score: %f < %f", results[0].Score, results[1].Score)
	}
}

// ---------------------------------------------------------------------------
// parseMemory tests
// ---------------------------------------------------------------------------

func TestParseMemory_ValidFrontmatter(t *testing.T) {
	t.Parallel()
	content := `---
name: Test Memory
description: A test memory entry
type: user
created: 2026-06-16T12:00:00Z
---
This is the memory content.`
	filename := "memory_12345.md"

	mem := parseMemory(content, filename)
	if mem.ID != "memory_12345" {
		t.Errorf("ID = %q, want %q", mem.ID, "memory_12345")
	}
	if mem.Name != "Test Memory" {
		t.Errorf("Name = %q, want %q", mem.Name, "Test Memory")
	}
	if mem.Description != "A test memory entry" {
		t.Errorf("Description = %q, want %q", mem.Description, "A test memory entry")
	}
	if mem.Type != "user" {
		t.Errorf("Type = %q, want %q", mem.Type, "user")
	}
	if mem.Created.IsZero() {
		t.Error("Created should not be zero")
	}
	if !strings.Contains(mem.Content, "memory content") {
		t.Errorf("Content should contain 'memory content', got %q", mem.Content)
	}
}

func TestParseMemory_MissingFrontmatter(t *testing.T) {
	t.Parallel()
	content := "Just some plain content without frontmatter"
	filename := "memory_999.md"

	mem := parseMemory(content, filename)
	if mem.ID != "memory_999" {
		t.Errorf("ID = %q, want %q", mem.ID, "memory_999")
	}
	if mem.Name != "" {
		t.Errorf("Name should be empty, got %q", mem.Name)
	}
	if mem.Content != content {
		t.Errorf("Content = %q, want %q", mem.Content, content)
	}
}

func TestParseMemory_InvalidDate(t *testing.T) {
	t.Parallel()
	content := `---
name: Bad Date
created: not-a-date
---
Content`

	mem := parseMemory(content, "memory_1.md")
	if !mem.Created.IsZero() {
		t.Error("Created should be zero for invalid date")
	}
}

func TestParseMemory_EmptyContent(t *testing.T) {
	t.Parallel()
	mem := parseMemory("", "memory_0.md")
	if mem.ID != "memory_0" {
		t.Errorf("ID = %q, want %q", mem.ID, "memory_0")
	}
	if mem.Content != "" {
		t.Errorf("Content should be empty, got %q", mem.Content)
	}
}

// ---------------------------------------------------------------------------
// getDefaultMemoryDir tests
// ---------------------------------------------------------------------------

func TestGetDefaultMemoryDir(t *testing.T) {
	t.Parallel()
	dir := getDefaultMemoryDir()
	if dir != ".sick-memory" {
		t.Errorf("expected '.sick-memory', got %q", dir)
	}
}

// ---------------------------------------------------------------------------
// SuccessResponse / ErrorResponse formatting tests
// ---------------------------------------------------------------------------

func TestSuccessResponse_JSON(t *testing.T) {
	// Capture stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	old := os.Stdout
	os.Stdout = w

	successResponse("test-data")

	w.Close()
	os.Stdout = old

	var buf strings.Builder
	_, _ = copyBuffer(&buf, r)
	r.Close()

	output := strings.TrimSpace(buf.String())
	if !strings.Contains(output, `"version"`) {
		t.Errorf("response should contain 'version' field: %s", output)
	}
	if !strings.Contains(output, `"test-data"`) {
		t.Errorf("response should contain data: %s", output)
	}
}

// ---------------------------------------------------------------------------
// getProjectMemoryPath tests
// ---------------------------------------------------------------------------

func TestGetProjectMemoryPath_Format(t *testing.T) {
	t.Parallel()
	// This function uses global dir from user home, so we just verify structure
	path := getProjectMemoryPath("/some/project/root")
	if !strings.Contains(path, "projects") {
		t.Errorf("path should contain 'projects' subdirectory: %s", path)
	}
	if !strings.Contains(path, "root") {
		t.Errorf("path should contain sanitized root: %s", path)
	}
	if !strings.HasSuffix(path, "memory") {
		t.Errorf("path should end with 'memory': %s", path)
	}
}

func TestGetProjectMemoryPath_SanitizesRoot(t *testing.T) {
	t.Parallel()
	path := getProjectMemoryPath("/home/user/my project:test")
	// The git root should be sanitized (slashes, colons, spaces replaced)
	// The path will be: <globalDir>/projects/<sanitized>/memory
	parts := strings.Split(path, string(filepath.Separator))
	sanitizedPart := parts[len(parts)-2] // the second-to-last part before "memory"
	if strings.ContainsAny(sanitizedPart, "/: ") {
		t.Errorf("sanitized part still has special chars: %s", sanitizedPart)
	}
}

// ---------------------------------------------------------------------------
// buildSearchIndex tests
// ---------------------------------------------------------------------------

func TestBuildSearchIndex_NoMemoryDir(t *testing.T) {
	t.Parallel()
	_, err := buildSearchIndex("/tmp/nonexistent-path-sick-memory-test")
	if err == nil {
		t.Error("expected error for non-existent memory directory")
	}
}

func TestBuildSearchIndex_WithMemoryFiles(t *testing.T) {
	t.Parallel()
	// Create a temp directory with memory files
	tmpDir := t.TempDir()

	mem1 := `---
name: Alpha Memory
description: first test entry
type: user
created: 2026-06-16T10:00:00Z
---
Alpha content for testing.`
	mem2 := `---
name: Beta Memory
description: second test entry
type: project
created: 2026-06-16T11:00:00Z
---
Beta content for testing.`

	if err := os.WriteFile(filepath.Join(tmpDir, "memory_1.md"), []byte(mem1), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "memory_2.md"), []byte(mem2), 0644); err != nil {
		t.Fatal(err)
	}
	// Write a non-memory file (should be ignored)
	if err := os.WriteFile(filepath.Join(tmpDir, "README.md"), []byte("# ignore"), 0644); err != nil {
		t.Fatal(err)
	}
	// Write a dotfile (should be ignored)
	if err := os.WriteFile(filepath.Join(tmpDir, ".hidden"), []byte("hidden"), 0644); err != nil {
		t.Fatal(err)
	}

	idx, err := buildSearchIndex(tmpDir)
	if err != nil {
		t.Fatalf("buildSearchIndex failed: %v", err)
	}
	if idx.DocCount != 2 {
		t.Errorf("expected 2 documents, got %d", idx.DocCount)
	}
	if len(idx.Memories) != 2 {
		t.Errorf("expected 2 memories, got %d", len(idx.Memories))
	}
	// Check that keywords were extracted (both contain "content" and "testing")
	if idx.TermFreq["content"]["memory_1"] == 0 {
		t.Error("memory_1 should have 'content' keyword")
	}
	if idx.TermFreq["testing"]["memory_2"] == 0 {
		t.Error("memory_2 should have 'testing' keyword")
	}
}

// ---------------------------------------------------------------------------
// findGitRepositoryRoot tests
// ---------------------------------------------------------------------------

func TestFindGitRepositoryRoot_OutsideGit(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	_, err := findGitRepositoryRoot()
	if err == nil {
		t.Error("expected error when not in a git repo")
	}
}

// ---------------------------------------------------------------------------
// handleVersion tests
// ---------------------------------------------------------------------------

func TestHandleVersion_Output(t *testing.T) {
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	old := os.Stdout
	os.Stdout = w

	handleVersion()

	w.Close()
	os.Stdout = old

	var buf strings.Builder
	_, _ = copyBuffer(&buf, r)
	r.Close()

	output := strings.TrimSpace(buf.String())
	if !strings.Contains(output, Version) {
		t.Errorf("version output should contain %s, got: %s", Version, output)
	}
	if !strings.HasPrefix(output, "sick-memory version") {
		t.Errorf("version output should start with 'sick-memory version', got: %s", output)
	}
}

// ---------------------------------------------------------------------------
// errorResponse tests
// ---------------------------------------------------------------------------

func TestErrorResponse_JSON(t *testing.T) {
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	old := os.Stdout
	os.Stdout = w

	errorResponse(85, "invalid_argument", "test error", false)

	w.Close()
	os.Stdout = old

	var buf strings.Builder
	_, _ = copyBuffer(&buf, r)
	r.Close()

	output := strings.TrimSpace(buf.String())
	if !strings.Contains(output, `"code":85`) {
		t.Errorf("response should contain code 85: %s", output)
	}
	if !strings.Contains(output, `"type":"invalid_argument"`) {
		t.Errorf("response should contain error type: %s", output)
	}
	if !strings.Contains(output, `"message":"test error"`) {
		t.Errorf("response should contain message: %s", output)
	}
	if !strings.Contains(output, `"recoverable":false`) {
		t.Errorf("response should contain recoverable=false: %s", output)
	}
}

func TestErrorResponse_Recoverable(t *testing.T) {
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	old := os.Stdout
	os.Stdout = w

	errorResponse(92, "resource_error", "retryable error", true)

	w.Close()
	os.Stdout = old

	var buf strings.Builder
	_, _ = copyBuffer(&buf, r)
	r.Close()

	output := strings.TrimSpace(buf.String())
	if !strings.Contains(output, `"recoverable":true`) {
		t.Errorf("response should contain recoverable=true: %s", output)
	}
}

// ---------------------------------------------------------------------------
// saveSearchIndex / loadSearchIndex tests
// ---------------------------------------------------------------------------

func TestSaveAndLoadSearchIndex(t *testing.T) {
	dir := t.TempDir()

	idx := &SearchIndex{
		TermFreq: map[string]map[string]int{
			"hello": {"mem1": 2, "mem2": 1},
			"world": {"mem1": 1},
		},
		DocFreq: map[string]int{
			"hello": 2,
			"world": 1,
		},
		DocCount: 2,
		Memories: map[string]Memory{
			"mem1": {ID: "mem1", Name: "First", Content: "hello world"},
			"mem2": {ID: "mem2", Name: "Second", Content: "hello only"},
		},
	}

	// Save index
	if err := saveSearchIndex(dir, idx); err != nil {
		t.Fatalf("saveSearchIndex failed: %v", err)
	}

	// Verify index file was created
	indexFile := filepath.Join(dir, "search_index.json")
	if _, err := os.Stat(indexFile); os.IsNotExist(err) {
		t.Fatal("search_index.json was not created")
	}

	// Load index back
	loaded, err := loadSearchIndex(dir)
	if err != nil {
		t.Fatalf("loadSearchIndex failed: %v", err)
	}

	// Verify loaded data matches
	if loaded.DocCount != idx.DocCount {
		t.Errorf("DocCount = %d, want %d", loaded.DocCount, idx.DocCount)
	}
	if len(loaded.Memories) != len(idx.Memories) {
		t.Errorf("len(Memories) = %d, want %d", len(loaded.Memories), len(idx.Memories))
	}
	if loaded.TermFreq["hello"]["mem1"] != 2 {
		t.Errorf("TermFreq[hello][mem1] = %d, want 2", loaded.TermFreq["hello"]["mem1"])
	}
}

func TestLoadSearchIndex_BuildsFromScratch(t *testing.T) {
	dir := t.TempDir()

	// Create a memory file (no cached index)
	memContent := `---
name: Fresh Memory
description: built from scratch
type: user
created: 2026-06-17T00:00:00Z
---
Fresh content for indexing.`
	if err := os.WriteFile(filepath.Join(dir, "memory_1.md"), []byte(memContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Load should build index from scratch
	idx, err := loadSearchIndex(dir)
	if err != nil {
		t.Fatalf("loadSearchIndex failed: %v", err)
	}

	if idx.DocCount != 1 {
		t.Errorf("DocCount = %d, want 1", idx.DocCount)
	}
	if _, ok := idx.Memories["memory_1"]; !ok {
		t.Error("memory_1 should be in index")
	}

	// Verify cached index was saved
	indexFile := filepath.Join(dir, "search_index.json")
	if _, err := os.Stat(indexFile); os.IsNotExist(err) {
		t.Error("cached search_index.json should exist after build")
	}
}

func TestLoadSearchIndex_NonExistentDir(t *testing.T) {
	_, err := loadSearchIndex("/tmp/nonexistent-path-sick-memory-test")
	if err == nil {
		t.Error("expected error for non-existent directory")
	}
}

// ---------------------------------------------------------------------------
// getGlobalSickMemoryDir tests
// ---------------------------------------------------------------------------

func TestGetGlobalSickMemoryDir_UsesHomeDir(t *testing.T) {
	dir := getGlobalSickMemoryDir()
	if !strings.HasSuffix(dir, ".sick-memory") {
		t.Errorf("expected path to end with '.sick-memory', got %q", dir)
	}
	if dir == ".sick-memory" {
		t.Error("getGlobalSickMemoryDir should return an absolute path, not the fallback")
	}
}

// ---------------------------------------------------------------------------
// copyBuffer helper
// ---------------------------------------------------------------------------

func copyBuffer(dst *strings.Builder, src *os.File) (int64, error) {
	var total int64
	buf := make([]byte, 4096)
	for {
		n, err := src.Read(buf)
		if n > 0 {
			dst.Write(buf[:n])
			total += int64(n)
		}
		if err != nil {
			return total, err
		}
	}
}
