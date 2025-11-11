package types

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

// DownloadProgress tracks the download state for resumable downloads
type DownloadProgress struct {
	ModelName    string                   `json:"model_name"`
	TotalSize    int64                    `json:"total_size"`
	Downloaded   int64                    `json:"downloaded"`
	Files        map[string]*FileProgress `json:"files"`
	LastModified time.Time                `json:"last_modified"`
	Version      int                      `json:"version"` // For future compatibility
}

// FileProgress tracks individual file download progress
type FileProgress struct {
	FileName       string             `json:"file_name"`
	TotalSize      int64              `json:"total_size"`
	Downloaded     int64              `json:"downloaded"`
	CompletedRange []CompletedRange   `json:"completed_ranges"`
	SHA256         string             `json:"sha256,omitempty"` // Optional validation
}

// CompletedRange represents a downloaded byte range
type CompletedRange struct {
	Start int64 `json:"start"`
	End   int64 `json:"end"` // Exclusive
}

// NewDownloadProgress creates a new download progress tracker
func NewDownloadProgress(modelName string, totalSize int64) *DownloadProgress {
	return &DownloadProgress{
		ModelName:    modelName,
		TotalSize:    totalSize,
		Files:        make(map[string]*FileProgress),
		LastModified: time.Now(),
		Version:      1,
	}
}

// AddFile adds a file to track
func (dp *DownloadProgress) AddFile(fileName string, size int64) {
	dp.Files[fileName] = &FileProgress{
		FileName:       fileName,
		TotalSize:      size,
		CompletedRange: make([]CompletedRange, 0),
	}
}

// MarkRangeComplete marks a byte range as downloaded and merges adjacent ranges
func (fp *FileProgress) MarkRangeComplete(start, end int64) {
	// Add the new range
	newRange := CompletedRange{Start: start, End: end}
	fp.CompletedRange = append(fp.CompletedRange, newRange)
	
	// Merge adjacent/overlapping ranges
	fp.mergeRanges()
	
	// Update downloaded count
	fp.Downloaded = fp.calculateDownloaded()
}

// mergeRanges merges overlapping or adjacent ranges
func (fp *FileProgress) mergeRanges() {
	if len(fp.CompletedRange) <= 1 {
		return
	}

	// Sort ranges by start position
	for i := 0; i < len(fp.CompletedRange)-1; i++ {
		for j := i + 1; j < len(fp.CompletedRange); j++ {
			if fp.CompletedRange[i].Start > fp.CompletedRange[j].Start {
				fp.CompletedRange[i], fp.CompletedRange[j] = fp.CompletedRange[j], fp.CompletedRange[i]
			}
		}
	}

	// Merge overlapping ranges
	merged := make([]CompletedRange, 0, len(fp.CompletedRange))
	current := fp.CompletedRange[0]

	for i := 1; i < len(fp.CompletedRange); i++ {
		next := fp.CompletedRange[i]
		if current.End >= next.Start {
			// Merge overlapping ranges
			current.End = max(current.End, next.End)
		} else {
			merged = append(merged, current)
			current = next
		}
	}
	merged = append(merged, current)
	fp.CompletedRange = merged
}

// calculateDownloaded calculates total downloaded bytes from ranges
func (fp *FileProgress) calculateDownloaded() int64 {
	var total int64
	for _, r := range fp.CompletedRange {
		total += r.End - r.Start
	}
	return total
}

// IsComplete checks if file is fully downloaded
func (fp *FileProgress) IsComplete() bool {
	return fp.Downloaded >= fp.TotalSize
}

// GetMissingRanges returns ranges that still need to be downloaded
func (fp *FileProgress) GetMissingRanges(chunkSize int64) []CompletedRange {
	if len(fp.CompletedRange) == 0 {
		// Nothing downloaded yet
		return []CompletedRange{{Start: 0, End: fp.TotalSize}}
	}

	missing := make([]CompletedRange, 0)
	
	// Check gap before first range
	if fp.CompletedRange[0].Start > 0 {
		missing = append(missing, CompletedRange{
			Start: 0,
			End:   fp.CompletedRange[0].Start,
		})
	}

	// Check gaps between ranges
	for i := 0; i < len(fp.CompletedRange)-1; i++ {
		gap := fp.CompletedRange[i+1].Start - fp.CompletedRange[i].End
		if gap > 0 {
			missing = append(missing, CompletedRange{
				Start: fp.CompletedRange[i].End,
				End:   fp.CompletedRange[i+1].Start,
			})
		}
	}

	// Check gap after last range
	lastEnd := fp.CompletedRange[len(fp.CompletedRange)-1].End
	if lastEnd < fp.TotalSize {
		missing = append(missing, CompletedRange{
			Start: lastEnd,
			End:   fp.TotalSize,
		})
	}

	return missing
}

// Save saves progress to disk
func (dp *DownloadProgress) Save(outputPath string) error {
	dp.LastModified = time.Now()
	
	progressFile := filepath.Join(outputPath, ".download_progress.json")
	data, err := json.MarshalIndent(dp, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(progressFile, data, 0644)
}

// LoadDownloadProgress loads progress from disk
func LoadDownloadProgress(outputPath string) (*DownloadProgress, error) {
	progressFile := filepath.Join(outputPath, ".download_progress.json")
	
	data, err := os.ReadFile(progressFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil // No previous progress
		}
		return nil, err
	}

	var progress DownloadProgress
	if err := json.Unmarshal(data, &progress); err != nil {
		return nil, err
	}

	return &progress, nil
}

// CleanupProgress removes progress file after successful download
func CleanupProgress(outputPath string) error {
	progressFile := filepath.Join(outputPath, ".download_progress.json")
	err := os.Remove(progressFile)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

func max(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}

