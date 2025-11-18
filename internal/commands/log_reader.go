package commands

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"
)

// LogReader handles log file reading and streaming
type LogReader struct {
	MateID string
}

// NewLogReader creates a new log reader
func NewLogReader(mateID string) *LogReader {
	return &LogReader{
		MateID: mateID,
	}
}

// ReadLogRequest represents the read_log command payload
type ReadLogRequest struct {
	SessionID string `json:"sessionId"` // Session ID from Navigator
	Path      string `json:"path"`
	Mode      string `json:"mode"`  // "smart", "full", "errors-only"
	Lines     int    `json:"lines"` // For future use
}

// LogDataMessage represents a log data chunk message
type LogDataMessage struct {
	SessionID   string  `json:"sessionId"`
	Chunk       string  `json:"chunk"`
	Progress    float64 `json:"progress"`
	CurrentLine int     `json:"currentLine"` // Current line number
	TotalLines  int     `json:"totalLines"`  // Total lines in file
	ChunkNumber int     `json:"chunkNumber"` // Which chunk (1-based)
	TotalChunks int     `json:"totalChunks"` // Total number of chunks
}

// LogCompleteMessage represents the log completion message
type LogCompleteMessage struct {
	SessionID string `json:"sessionId"`
	TotalSize int    `json:"totalSize"`
}

// HandleReadLogCommand processes the read_log command with line-based streaming
func (lr *LogReader) HandleReadLogCommand(request ReadLogRequest, sendMessage func(msgType string, data interface{})) error {
	log.Printf("Reading log file: %s (mode: %s)", request.Path, request.Mode)

	// Read log file
	content, err := os.ReadFile(request.Path)
	if err != nil {
		return fmt.Errorf("failed to read log file: %w", err)
	}

	// Use session ID from Navigator (or generate if not provided for backwards compatibility)
	sessionID := request.SessionID
	if sessionID == "" {
		sessionID = fmt.Sprintf("%s-%d", lr.MateID, time.Now().UnixMilli())
		log.Printf("Warning: No sessionId provided, generated: %s", sessionID)
	}

	// Split into lines
	allLines := strings.Split(string(content), "\n")
	totalLines := len(allLines)

	log.Printf("Log file loaded: %d lines, sessionId: %s", totalLines, sessionID)

	// Apply filtering based on mode
	var linesToProcess []string
	switch request.Mode {
	case "smart":
		linesToProcess = lr.filterRelevantLinesList(allLines)
	case "errors-only":
		linesToProcess = lr.filterErrorsOnlyList(allLines)
	default: // "full" mode
		linesToProcess = allLines
	}

	log.Printf("After filtering: %d lines", len(linesToProcess))

	// Stream in line-based chunks (1000 lines per chunk for LLM context)
	linesPerChunk := 1000
	totalChunks := (len(linesToProcess) + linesPerChunk - 1) / linesPerChunk

	for chunkNum := 0; chunkNum < totalChunks; chunkNum++ {
		start := chunkNum * linesPerChunk
		end := start + linesPerChunk
		if end > len(linesToProcess) {
			end = len(linesToProcess)
		}

		chunkLines := linesToProcess[start:end]
		chunk := strings.Join(chunkLines, "\n")
		progress := float64(end) / float64(len(linesToProcess)) * 100

		// Send chunk via WebSocket
		sendMessage("log_data", LogDataMessage{
			SessionID:   sessionID,
			Chunk:       chunk,
			Progress:    progress,
			CurrentLine: end,
			TotalLines:  len(linesToProcess),
			ChunkNumber: chunkNum + 1,
			TotalChunks: totalChunks,
		})

		log.Printf("Sent chunk %d/%d: lines %d-%d (%.1f%%)",
			chunkNum+1, totalChunks, start+1, end, progress)

		// Small delay between chunks to prevent overwhelming the connection
		time.Sleep(10 * time.Millisecond)
	}

	// Send completion message
	sendMessage("log_complete", LogCompleteMessage{
		SessionID: sessionID,
		TotalSize: len(linesToProcess),
	})

	log.Printf("Log transfer completed: session=%s, %d lines in %d chunks",
		sessionID, len(linesToProcess), totalChunks)
	return nil
}

// filterRelevantLines filters log content for relevant entries (smart mode)
func (lr *LogReader) filterRelevantLines(content string) string {
	lines := strings.Split(content, "\n")
	relevant := []string{}

	// Keywords that indicate important log entries
	keywords := []string{
		"error", "ERROR", "Error",
		"warn", "WARN", "warning", "Warning",
		"fail", "FAIL", "failed", "Failed",
		"critical", "CRITICAL", "Critical",
		"panic", "Panic", "PANIC",
		"segfault", "segmentation fault",
		"out of memory", "OOM", "oom",
		"authentication failure", "auth failed",
		"denied", "Denied", "DENIED",
		"timeout", "Timeout", "TIMEOUT",
		"refused", "Refused", "REFUSED",
		"exception", "Exception", "EXCEPTION",
	}

	for _, line := range lines {
		for _, keyword := range keywords {
			if strings.Contains(line, keyword) {
				relevant = append(relevant, line)
				break
			}
		}
	}

	// If filtering resulted in empty output, return at least some context
	if len(relevant) == 0 {
		// Return last 50 lines as fallback
		startIdx := len(lines) - 50
		if startIdx < 0 {
			startIdx = 0
		}
		return strings.Join(lines[startIdx:], "\n")
	}

	return strings.Join(relevant, "\n")
}

// filterErrorsOnly filters only error-level entries
func (lr *LogReader) filterErrorsOnly(content string) string {
	lines := strings.Split(content, "\n")
	errors := []string{}

	// Only critical keywords
	keywords := []string{
		"error", "ERROR", "Error",
		"critical", "CRITICAL", "Critical",
		"panic", "Panic", "PANIC",
		"fail", "FAIL", "failed", "Failed",
		"segfault", "segmentation fault",
		"exception", "Exception", "EXCEPTION",
	}

	for _, line := range lines {
		for _, keyword := range keywords {
			if strings.Contains(line, keyword) {
				errors = append(errors, line)
				break
			}
		}
	}

	// If no errors found, return empty with notice
	if len(errors) == 0 {
		return "No errors found in log file."
	}

	return strings.Join(errors, "\n")
}

// filterRelevantLinesList filters relevant lines (list-based for streaming)
func (lr *LogReader) filterRelevantLinesList(lines []string) []string {
	relevant := []string{}

	// Keywords that indicate important log entries
	keywords := []string{
		"error", "ERROR", "Error",
		"warn", "WARN", "warning", "Warning",
		"fail", "FAIL", "failed", "Failed",
		"critical", "CRITICAL", "Critical",
		"panic", "Panic", "PANIC",
		"segfault", "segmentation fault",
		"out of memory", "OOM", "oom",
		"authentication failure", "auth failed",
		"denied", "Denied", "DENIED",
		"timeout", "Timeout", "TIMEOUT",
		"refused", "Refused", "REFUSED",
		"exception", "Exception", "EXCEPTION",
	}

	for _, line := range lines {
		for _, keyword := range keywords {
			if strings.Contains(line, keyword) {
				relevant = append(relevant, line)
				break
			}
		}
	}

	// If filtering resulted in empty output, return at least some context
	if len(relevant) == 0 {
		// Return last 50 lines as fallback
		startIdx := len(lines) - 50
		if startIdx < 0 {
			startIdx = 0
		}
		return lines[startIdx:]
	}

	return relevant
}

// filterErrorsOnlyList filters only error-level entries (list-based for streaming)
func (lr *LogReader) filterErrorsOnlyList(lines []string) []string {
	errors := []string{}

	// Only critical keywords
	keywords := []string{
		"error", "ERROR", "Error",
		"critical", "CRITICAL", "Critical",
		"panic", "Panic", "PANIC",
		"fail", "FAIL", "failed", "Failed",
		"segfault", "segmentation fault",
		"exception", "Exception", "EXCEPTION",
	}

	for _, line := range lines {
		for _, keyword := range keywords {
			if strings.Contains(line, keyword) {
				errors = append(errors, line)
				break
			}
		}
	}

	// If no errors found, return notice
	if len(errors) == 0 {
		return []string{"No errors found in log file."}
	}

	return errors
}
