package engine

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/Kaffyn/Vectora/core/db"
	"github.com/Kaffyn/Vectora/core/infra"
	"github.com/Kaffyn/Vectora/core/llm"
	"github.com/Kaffyn/Vectora/core/policies"
)

type EmbedJobConfig struct {
	RootPath       string
	Include        string
	Exclude        string
	Workspace      string
	Force          bool
	CollectionName string
}

type EmbedProgress struct {
	CurrentIdx      int     `json:"current_idx"`
	TotalFiles      int     `json:"total_files"`
	DisplayPath     string  `json:"display_path"`
	ElapsedSeconds  float64 `json:"elapsed_seconds"`
	FilesSkipped    int     `json:"files_skipped"`
	FilesAlready    int     `json:"files_already"`
	TotalChunks     int     `json:"total_chunks"`
	TotalEmbedded   int     `json:"total_embedded"`
	TotalErrors     int     `json:"total_errors"`
	IsComplete      bool    `json:"is_complete"`
	CurrentFilePath string  `json:"current_file_path"`
	HasError        bool    `json:"has_error"`
	ErrorMsg        string  `json:"error_msg"`
	FileChunks      int     `json:"file_chunks"`
}

var (
	activeJobs     = make(map[string]bool)
	activeJobsLock sync.Mutex
)

// RunEmbedJob contains the entire directory walking and chunking logic.
func RunEmbedJob(ctx context.Context, cfg EmbedJobConfig, kvStore db.KVStore, storage db.VectorStore, provider llm.Provider, onProgress func(EmbedProgress)) {
	activeJobsLock.Lock()
	if activeJobs[cfg.Workspace] {
		activeJobsLock.Unlock()
		onProgress(EmbedProgress{IsComplete: true, HasError: true, ErrorMsg: "A job is already running for this workspace"})
		return
	}
	activeJobs[cfg.Workspace] = true
	activeJobsLock.Unlock()

	defer func() {
		activeJobsLock.Lock()
		delete(activeJobs, cfg.Workspace)
		activeJobsLock.Unlock()
	}()

	absPath, err := filepath.Abs(cfg.RootPath)
	if err != nil {
		onProgress(EmbedProgress{IsComplete: true, HasError: true, ErrorMsg: fmt.Sprintf("invalid path: %v", err)})
		return
	}

	// Register workspace path mapping for `workspace ls` display
	registerWorkspacePath(cfg.CollectionName, absPath)

	guardian := policies.NewGuardian(absPath)
	if provider == nil || !provider.IsConfigured() {
		onProgress(EmbedProgress{IsComplete: true, HasError: true, ErrorMsg: "No LLM provider configured. Set it in .env or via API."})
		return
	}

	// Parse includes and excludes similarly
	var includePatterns []string
	if cfg.Include != "" {
		includePatterns = strings.Split(cfg.Include, ",")
		for i, p := range includePatterns {
			includePatterns[i] = strings.TrimSpace(p)
		}
	}

	var excludePatterns []string
	var unignorePatterns []string
	if cfg.Exclude != "" {
		parts := strings.Split(cfg.Exclude, ",")
		for _, p := range parts {
			p = strings.TrimSpace(p)
			if strings.HasPrefix(p, "!") {
				unignorePatterns = append(unignorePatterns, strings.TrimPrefix(p, "!"))
			} else {
				excludePatterns = append(excludePatterns, p)
			}
		}
	}

	ignorePath := filepath.Join(absPath, ".embedignore")
	if data, err := os.ReadFile(ignorePath); err == nil {
		lines := strings.Split(string(data), "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}
			if strings.HasPrefix(line, "!") {
				unignorePatterns = append(unignorePatterns, strings.TrimPrefix(line, "!"))
			} else {
				excludePatterns = append(excludePatterns, line)
			}
		}
	}

	var filesToEmbed []string
	var filesSkipped int
	var filesAlreadyIndexed int

	err = filepath.WalkDir(absPath, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return nil
		}
		if d.IsDir() {
			relPath, _ := filepath.Rel(absPath, path)
			relPath = filepath.ToSlash(relPath)
			name := d.Name()

			if guardian.IsExcludedDir(name) {
				return filepath.SkipDir
			}

			isIgnored := false
			for _, pattern := range excludePatterns {
				if m, _ := filepath.Match(pattern, relPath); m {
					isIgnored = true
				} else if m, _ := filepath.Match(pattern, name); m {
					isIgnored = true
				} else if strings.Contains(pattern, "/") {
					if strings.HasPrefix(relPath, pattern) {
						isIgnored = true
					} else {
						m, _ := filepath.Match(pattern+"/*", relPath)
						if m {
							isIgnored = true
						}
					}
				}
			}

			if isIgnored {
				for _, pattern := range unignorePatterns {
					if m, _ := filepath.Match(pattern, relPath); m {
						isIgnored = false
					} else if m, _ := filepath.Match(pattern, name); m {
						isIgnored = false
					} else if strings.Contains(pattern, "/") {
						if strings.HasPrefix(relPath, pattern) {
							isIgnored = false
						} else {
							m, _ := filepath.Match(pattern+"/*", relPath)
							if m {
								isIgnored = false
							}
						}
					}
				}
			}

			if isIgnored {
				return filepath.SkipDir
			}
			return nil
		}

		if guardian.IsProtected(path) {
			filesSkipped++
			return nil
		}

		if len(includePatterns) > 0 {
			name := d.Name()
			matched := false
			for _, pattern := range includePatterns {
				if m, _ := filepath.Match(pattern, name); m {
					matched = true
					break
				}
			}
			if !matched {
				filesSkipped++
				return nil
			}
		}

		relPath, _ := filepath.Rel(absPath, path)
		relPath = filepath.ToSlash(relPath)
		name := d.Name()

		isIgnored := false
		for _, pattern := range excludePatterns {
			if m, _ := filepath.Match(pattern, relPath); m {
				isIgnored = true
			} else if m, _ := filepath.Match(pattern, name); m {
				isIgnored = true
			} else if strings.Contains(pattern, "/") {
				if strings.HasPrefix(relPath, pattern) {
					isIgnored = true
				} else {
					m, _ := filepath.Match(pattern+"/*", relPath)
					if m {
						isIgnored = true
					}
				}
			}
		}

		if isIgnored {
			for _, pattern := range unignorePatterns {
				if m, _ := filepath.Match(pattern, relPath); m {
					isIgnored = false
				} else if m, _ := filepath.Match(pattern, name); m {
					isIgnored = false
				} else if strings.Contains(pattern, "/") {
					if strings.HasPrefix(relPath, pattern) {
						isIgnored = false
					} else {
						m, _ := filepath.Match(pattern+"/*", relPath)
						if m {
							isIgnored = false
						}
					}
				}
			}
		}

		if isIgnored {
			filesSkipped++
			return nil
		}

		content, readErr := os.ReadFile(path)
		if readErr != nil {
			filesSkipped++
			return nil
		}

		if !cfg.Force {
			existing, _ := kvStore.Get(ctx, "file_index", relPath)
			if existing != nil {
				var entry db.FileIndexEntry
				if err := json.Unmarshal(existing, &entry); err == nil {
					contentHash := db.CalculateHash(string(content))
					if entry.ContentHash == contentHash {
						filesAlreadyIndexed++
						return nil
					}
				}
			}
		}

		filesToEmbed = append(filesToEmbed, path)
		return nil
	})

	if err != nil {
		onProgress(EmbedProgress{IsComplete: true, HasError: true, ErrorMsg: fmt.Sprintf("walk error: %v", err)})
		return
	}

	if len(filesToEmbed) == 0 {
		onProgress(EmbedProgress{
			IsComplete:   true,
			FilesSkipped: filesSkipped,
			FilesAlready: filesAlreadyIndexed,
			TotalFiles:   0,
		})
		return
	}

	infra.NotifyOS("Vectora Indexing", fmt.Sprintf("Started indexing %d files", len(filesToEmbed)))

	totalEmbedded := 0
	totalChunks := 0
	totalErrors := 0
	startTime := time.Now()

	for i, filePath := range filesToEmbed {
		select {
		case <-ctx.Done():
			onProgress(EmbedProgress{IsComplete: true, HasError: true, ErrorMsg: "context canceled"})
			return
		default:
		}

		relPath, _ := filepath.Rel(absPath, filePath)
		displayPath := relPath
		if len(displayPath) > 40 {
			displayPath = "..." + displayPath[len(displayPath)-37:]
		}

		// Emit "Started file" progress
		onProgress(EmbedProgress{
			CurrentIdx:      i,
			TotalFiles:      len(filesToEmbed),
			DisplayPath:     displayPath,
			ElapsedSeconds:  time.Since(startTime).Seconds(),
			CurrentFilePath: relPath,
		})

		content, err := os.ReadFile(filePath)
		if err != nil {
			totalErrors++
			onProgress(EmbedProgress{
				CurrentIdx: i, TotalFiles: len(filesToEmbed), DisplayPath: displayPath,
				ElapsedSeconds: time.Since(startTime).Seconds(), CurrentFilePath: relPath,
				HasError: true, ErrorMsg: "read error", FileChunks: 0,
				TotalEmbedded: totalEmbedded, TotalChunks: totalChunks, TotalErrors: totalErrors,
			})
			continue
		}

		chunks := chunkContent(string(content), 800, 100)
		fileChunks := 0

		language := "text"
		ext := strings.ToLower(filepath.Ext(filePath))
		switch ext {
		case ".go":
			language = "go"
		case ".py":
			language = "python"
		case ".js", ".ts":
			language = "javascript"
		case ".md":
			language = "markdown"
		}

		embedErr := false
		var lastEmbedError string
		for j, chunk := range chunks {
			vec, err := provider.Embed(ctx, chunk, "")
			if err != nil {
				totalErrors++
				embedErr = true
				lastEmbedError = fmt.Sprintf("embed error: %v", err)
				break
			}

			docID := fmt.Sprintf("%s:%d", relPath, j)
			err = storage.UpsertChunk(ctx, cfg.CollectionName, db.Chunk{
				ID:       docID,
				Content:  chunk,
				Metadata: map[string]string{"source": relPath, "filename": filepath.Base(filePath), "language": language},
				Vector:   vec,
			})
			if err != nil {
				totalErrors++
				embedErr = true
				lastEmbedError = fmt.Sprintf("store error: %v", err)
				break
			}
			fileChunks++
		}

		if embedErr {
			onProgress(EmbedProgress{
				CurrentIdx: i, TotalFiles: len(filesToEmbed), DisplayPath: displayPath,
				ElapsedSeconds: time.Since(startTime).Seconds(), CurrentFilePath: relPath,
				HasError: true, ErrorMsg: lastEmbedError, FileChunks: fileChunks,
				TotalEmbedded: totalEmbedded, TotalChunks: totalChunks, TotalErrors: totalErrors,
			})
		} else {
			if fileChunks > 0 {
				contentHash := db.CalculateHash(string(content))
				entry := db.FileIndexEntry{
					AbsolutePath: filePath,
					ContentHash:  contentHash,
					SizeBytes:    int64(len(content)),
				}
				entryBytes, _ := json.Marshal(entry)
				kvStore.Set(ctx, "file_index", relPath, entryBytes)

				totalEmbedded++
				totalChunks += fileChunks
			}

			onProgress(EmbedProgress{
				CurrentIdx: i, TotalFiles: len(filesToEmbed), DisplayPath: displayPath,
				ElapsedSeconds: time.Since(startTime).Seconds(), CurrentFilePath: relPath,
				HasError: false, FileChunks: fileChunks,
				TotalEmbedded: totalEmbedded, TotalChunks: totalChunks, TotalErrors: totalErrors,
			})
		}
	}

	infra.NotifyOS("Vectora Indexing", fmt.Sprintf("Completed indexing. %d files embedded.", totalEmbedded))

	onProgress(EmbedProgress{
		IsComplete:     true,
		FilesSkipped:   filesSkipped,
		FilesAlready:   filesAlreadyIndexed,
		TotalFiles:     len(filesToEmbed),
		TotalEmbedded:  totalEmbedded,
		TotalChunks:    totalChunks,
		TotalErrors:    totalErrors,
		ElapsedSeconds: time.Since(startTime).Seconds(),
	})
}

// Retain local chunk definition since chunking is local per file.
func chunkContent(content string, maxTokens int, overlap int) []string {
	if len(content) == 0 {
		return []string{}
	}
	maxChars := maxTokens * 4
	overlapChars := overlap * 4
	runes := []rune(content)
	totalRunes := len(runes)

	if maxChars >= totalRunes {
		return []string{content}
	}

	var chunks []string
	start := 0
	for start < totalRunes {
		end := start + maxChars
		if end > totalRunes {
			end = totalRunes
		}
		chunkEnd := end
		if end < totalRunes {
			for i := end; i < totalRunes && i < end+maxChars/2; i++ {
				if runes[i] == '\n' {
					chunkEnd = i + 1
					break
				}
			}
		}
		chunks = append(chunks, string(runes[start:chunkEnd]))
		start = chunkEnd - overlapChars
		if start < 0 {
			start = 0
		}
		if start >= chunkEnd {
			break
		}
	}
	return chunks
}
