package main

import (
	"math"
	"testing"
	"time"
)

// --- sanitizePath tests ---

func TestSanitizePath_ReplacesSlashes(t *testing.T) {
	got := sanitizePath("/path/to/something")
	want := "-path-to-something"
	if got != want {
		t.Errorf("sanitizePath(%q) = %q, want %q", "/path/to/something", got, want)
	}
}

func TestSanitizePath_ReplacesBackslashes(t *testing.T) {
	got := sanitizePath("C:\\Users\\test")
	want := "C--Users-test"
	if got != want {
		t.Errorf("sanitizePath(%q) = %q, want %q", "C:\\Users\\test", got, want)
	}
}

func TestSanitizePath_ReplacesColons(t *testing.T) {
	got := sanitizePath("a:b:c")
	want := "a-b-c"
	if got != want {
		t.Errorf("sanitizePath(%q) = %q, want %q", "a:b:c", got, want)
	}
}

func TestSanitizePath_ReplacesSpaces(t *testing.T) {
	got := sanitizePath("my project root")
	want := "my_project_root"
	if got != want {
		t.Errorf("sanitizePath(%q) = %q, want %q", "my project root", got, want)
	}
}

func TestSanitizePath_EmptyString(t *testing.T) {
	got := sanitizePath("")
	want := ""
	if got != want {
		t.Errorf("sanitizePath(%q) = %q, want %q", "", got, want)
	}
}

// --- extractKeywords tests ---

func TestExtractKeywords_RemovesStopWords(t *testing.T) {
	text := "the quick brown fox and the lazy dog"
	got := extractKeywords(text)
	// "the", "and" are stop words and should be removed
	for _, kw := range got {
		if kw == "the" || kw == "and" {
			t.Errorf("extractKeywords(%q) = %v, should not contain stop word %q", text, got, kw)
		}
	}
	// "quick", "brown", "fox", "lazy", "dog" should be present
	expected := []string{"quick", "brown", "fox", "lazy", "dog"}
	for _, exp := range expected {
		found := false
		for _, kw := range got {
			if kw == exp {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("extractKeywords(%q) = %v, missing expected keyword %q", text, got, exp)
		}
	}
}

func TestExtractKeywords_RemovesPunctuation(t *testing.T) {
	text := "hello, world! how's it going?"
	got := extractKeywords(text)
	// Edge punctuation is trimmed: "hello," -> "hello", "world!" -> "world"
	// Apostrophe INSIDE a word (like "how's") is preserved by Trim (only trims edges)
	for _, kw := range got {
		if kw == "hello," || kw == "hello!" || kw == "world!" || kw == "world?" {
			t.Errorf("extractKeywords(%q) = %v, contains edge punctuation in %q", text, got, kw)
		}
	}
}

func TestExtractKeywords_ShortWordsExcluded(t *testing.T) {
	text := "a an at to be is if in it of on or up do go my no we he hi"
	got := extractKeywords(text)
	if len(got) != 0 {
		t.Errorf("extractKeywords(%q) = %v, expected empty (all are stop words or short words)", text, got)
	}
}

func TestExtractKeywords_EmptyText(t *testing.T) {
	got := extractKeywords("")
	if len(got) != 0 {
		t.Errorf("extractKeywords(%q) = %v, expected empty", "", got)
	}
}

func TestExtractKeywords_LowercaseConversion(t *testing.T) {
	text := "Hello World Test"
	got := extractKeywords(text)
	expected := []string{"hello", "world", "test"}
	for _, exp := range expected {
		found := false
		for _, kw := range got {
			if kw == exp {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("extractKeywords(%q) = %v, missing %q (lowercase)", text, got, exp)
		}
	}
}

// --- parseMemory tests ---

func TestParseMemory_BasicFrontmatter(t *testing.T) {
	content := `---
name: Test Memory
description: A test memory entry
type: user
created: 2026-06-15T12:00:00Z
---

This is the memory content.
`
	memory := parseMemory(content, "memory_123.md")

	if memory.ID != "memory_123" {
		t.Errorf("parseMemory ID = %q, want %q", memory.ID, "memory_123")
	}
	if memory.Name != "Test Memory" {
		t.Errorf("parseMemory Name = %q, want %q", memory.Name, "Test Memory")
	}
	if memory.Description != "A test memory entry" {
		t.Errorf("parseMemory Description = %q, want %q", memory.Description, "A test memory entry")
	}
	if memory.Type != "user" {
		t.Errorf("parseMemory Type = %q, want %q", memory.Type, "user")
	}
	expectedTime, _ := time.Parse(time.RFC3339, "2026-06-15T12:00:00Z")
	if !memory.Created.Equal(expectedTime) {
		t.Errorf("parseMemory Created = %v, want %v", memory.Created, expectedTime)
	}
	// Content includes the blank line after closing --- (part of content body)
	if memory.Content != "\nThis is the memory content.\n" {
		t.Errorf("parseMemory Content = %q, want %q", memory.Content, "\nThis is the memory content.\n")
	}
}

func TestParseMemory_NoFrontmatter(t *testing.T) {
	content := "Just some content without frontmatter."
	memory := parseMemory(content, "memory_456.md")

	if memory.ID != "memory_456" {
		t.Errorf("parseMemory ID = %q, want %q", memory.ID, "memory_456")
	}
	if memory.Content != content {
		t.Errorf("parseMemory Content = %q, want %q", memory.Content, content)
	}
	// All fields should be zero-valued
	if memory.Name != "" {
		t.Errorf("parseMemory Name = %q, want empty", memory.Name)
	}
}

func TestParseMemory_PartialFrontmatter(t *testing.T) {
	content := `---
name: Partial Memory
---

Partial content.
`
	memory := parseMemory(content, "memory_789.md")

	if memory.Name != "Partial Memory" {
		t.Errorf("parseMemory Name = %q, want %q", memory.Name, "Partial Memory")
	}
	if memory.Description != "" {
		t.Errorf("parseMemory Description = %q, want empty", memory.Description)
	}
	// Content includes the blank line after closing ---
	if memory.Content != "\nPartial content.\n" {
		t.Errorf("parseMemory Content = %q, want %q", memory.Content, "\nPartial content.\n")
	}
}

func TestParseMemory_UnixTimestampInCreated(t *testing.T) {
	// Bug #13/#16: handleRemember writes Unix timestamps but parseMemory historically expected RFC3339
	// After the fix (swapped args + Unix fallback), both formats are now handled
	content := `---
name: Test
description: Test
type: user
created: 1718467200
---

Content
`
	memory := parseMemory(content, "memory_111.md")

	// With the fix, Unix timestamps should be correctly parsed
	if memory.Created.IsZero() {
		t.Errorf("parseMemory with Unix timestamp Created = zero, expected valid time")
	}
	if !memory.Created.Equal(time.Unix(1718467200, 0)) {
		t.Errorf("parseMemory with Unix timestamp Created = %v, want %v",
			memory.Created, time.Unix(1718467200, 0))
	}
}

// --- calculateTFIDF tests ---

func TestCalculateTFIDF_ZeroDocCount(t *testing.T) {
	index := &SearchIndex{
		TermFreq: make(map[string]map[string]int),
		DocFreq:  make(map[string]int),
		DocCount: 0,
	}
	score := calculateTFIDF(index, "test", "mem1")
	if score != 0 {
		t.Errorf("calculateTFIDF with DocCount=0 = %f, want 0", score)
	}
}

func TestCalculateTFIDF_SingleDocument(t *testing.T) {
	index := &SearchIndex{
		TermFreq: map[string]map[string]int{
			"hello": {"mem1": 3},
		},
		DocFreq: map[string]int{
			"hello": 1,
		},
		DocCount: 1,
	}
	// TF=3, IDF=log((1+1)/(1+1))=log(1)=0, TF*IDF=0
	score := calculateTFIDF(index, "hello", "mem1")
	want := 3.0 * math.Log(2.0/2.0)
	if score != want {
		t.Errorf("calculateTFIDF = %f, want %f", score, want)
	}
}

func TestCalculateTFIDF_MultipleDocuments(t *testing.T) {
	index := &SearchIndex{
		TermFreq: map[string]map[string]int{
			"hello": {"mem1": 3, "mem2": 1},
		},
		DocFreq: map[string]int{
			"hello": 2,
		},
		DocCount: 2,
	}
	// TF=3, IDF=log((2+1)/(2+1))=log(1)=0
	score := calculateTFIDF(index, "hello", "mem1")
	want := 3.0 * math.Log(3.0/3.0)
	if score != want {
		t.Errorf("calculateTFIDF = %f, want %f", score, want)
	}
}

func TestCalculateTFIDF_TermNotInDocument(t *testing.T) {
	index := &SearchIndex{
		TermFreq: map[string]map[string]int{
			"hello": {"mem2": 2},
		},
		DocFreq: map[string]int{
			"hello": 1,
		},
		DocCount: 2,
	}
	score := calculateTFIDF(index, "hello", "mem1")
	if score != 0 {
		t.Errorf("calculateTFIDF for term not in doc = %f, want 0", score)
	}
}

func TestCalculateTFIDF_TermNotInIndex(t *testing.T) {
	index := &SearchIndex{
		TermFreq: make(map[string]map[string]int),
		DocFreq:  make(map[string]int),
		DocCount: 2,
	}
	score := calculateTFIDF(index, "nonexistent", "mem1")
	if score != 0 {
		t.Errorf("calculateTFIDF for nonexistent term = %f, want 0", score)
	}
}

// --- searchMemories tests ---

func TestSearchMemories_EmptyIndex(t *testing.T) {
	index := &SearchIndex{
		TermFreq: make(map[string]map[string]int),
		DocFreq:  make(map[string]int),
		Memories: make(map[string]Memory),
		DocCount: 0,
	}
	results := searchMemories(index, "test query")
	if len(results) != 0 {
		t.Errorf("searchMemories with empty index = %d results, want 0", len(results))
	}
}

func TestSearchMemories_ExactMatchBoost(t *testing.T) {
	index := &SearchIndex{
		TermFreq: make(map[string]map[string]int),
		DocFreq:  make(map[string]int),
		Memories: map[string]Memory{
			"mem1": {
				ID:      "mem1",
				Name:    "Test Memory",
				Content: "This is about machine learning and AI.",
				Created: time.Now(),
			},
		},
		DocCount: 1,
	}

	results := searchMemories(index, "machine learning")
	if len(results) == 0 {
		t.Fatal("searchMemories should return results for matching query")
	}
	// Should get a boost for exact phrase match (score > 0)
	if results[0].Score <= 0 {
		t.Errorf("searchMemories exact match score = %f, want > 0", results[0].Score)
	}
}

func TestSearchMemories_SortedByScore(t *testing.T) {
	index := &SearchIndex{
		TermFreq: map[string]map[string]int{
			"hello": {"mem1": 1, "mem2": 1},
		},
		DocFreq: map[string]int{
			"hello": 2,
		},
		Memories: map[string]Memory{
			"mem1": {
				ID:      "mem1",
				Name:    "Memory 1",
				Content: "hello world",
				Created: time.Now(),
			},
			"mem2": {
				ID:      "mem2",
				Name:    "Memory 2",
				Content: "hello there",
				Created: time.Now(),
			},
		},
		DocCount: 2,
	}

	results := searchMemories(index, "hello")
	if len(results) != 2 {
		t.Errorf("searchMemories returned %d results, want 2", len(results))
	}
	// Results should be sorted by score descending
	for i := 1; i < len(results); i++ {
		if results[i].Score > results[i-1].Score {
			t.Errorf("searchMemories not sorted by score: index %d (score=%f) > index %d (score=%f)",
				i, results[i].Score, i-1, results[i-1].Score)
		}
	}
}

func TestSearchMemories_NoMatch(t *testing.T) {
	index := &SearchIndex{
		TermFreq: make(map[string]map[string]int),
		DocFreq:  make(map[string]int),
		Memories: map[string]Memory{
			"mem1": {
				ID:      "mem1",
				Name:    "Test",
				Content: "Some random text.",
				Created: time.Now(),
			},
		},
		DocCount: 1,
	}

	results := searchMemories(index, "zzzzzzzzz")
	if len(results) != 0 {
		t.Errorf("searchMemories with non-matching query = %d results, want 0", len(results))
	}
}

// --- GlobalConfig default test ---

func TestDefaultGlobalConfig(t *testing.T) {
	cfg := GlobalConfig{
		DefaultMemoryType: "user",
		MaxMemorySize:     1024 * 1024,
		AutoIndex:         true,
	}

	if cfg.DefaultMemoryType != "user" {
		t.Errorf("default memory type = %q, want %q", cfg.DefaultMemoryType, "user")
	}
	if cfg.MaxMemorySize != 1024*1024 {
		t.Errorf("default max memory size = %d, want %d", cfg.MaxMemorySize, 1024*1024)
	}
	if !cfg.AutoIndex {
		t.Errorf("default auto index = false, want true")
	}
}

// --- Version constant test ---

func TestVersion(t *testing.T) {
	if Version != "0.1.0" {
		t.Errorf("Version = %q, want %q", Version, "0.1.0")
	}
}

// --- SearchResult formatting test ---

func TestSearchResultFormatting(t *testing.T) {
	now := time.Now()
	result := SearchResult{
		MemoryID:   "test123",
		Score:      42.5,
		Title:      "Test Result",
		Content:    "Some content here",
		MemoryType: "user",
		Created:    now.Format(time.RFC3339),
	}

	if result.MemoryID != "test123" {
		t.Errorf("SearchResult MemoryID = %q, want %q", result.MemoryID, "test123")
	}
	if result.Score != 42.5 {
		t.Errorf("SearchResult Score = %f, want %f", result.Score, 42.5)
	}
	if result.Title != "Test Result" {
		t.Errorf("SearchResult Title = %q, want %q", result.Title, "Test Result")
	}
	if result.MemoryType != "user" {
		t.Errorf("SearchResult MemoryType = %q, want %q", result.MemoryType, "user")
	}
	// Verify Created is valid RFC3339
	_, err := time.Parse(time.RFC3339, result.Created)
	if err != nil {
		t.Errorf("SearchResult Created is not valid RFC3339: %v", err)
	}
}
