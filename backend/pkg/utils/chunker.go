package utils

import (
	"strings"
	"unicode"
)

const (
	MaxChunkSize     = 10000 // Maximum characters per chunk
	OverlapSize      = 200   // Character overlap between chunks
	SentenceEndChars = ".!?。！？" // Punctuation marks that end sentences
)

// ChunkDocument splits a document into chunks with a maximum size while preserving sentence boundaries
// and maintaining context overlap between chunks.
func ChunkDocument(content string) []string {
	if len(content) == 0 {
		return []string{}
	}

	// If content is smaller than max chunk size, return as single chunk
	if len(content) <= MaxChunkSize {
		return []string{content}
	}

	var chunks []string
	start := 0

	for start < len(content) {
		end := start + MaxChunkSize

		// If this is the last chunk, take everything remaining
		if end >= len(content) {
			chunks = append(chunks, content[start:])
			break
		}

		// Find the best sentence boundary before the max chunk size
		chunkEnd := findSentenceBoundary(content, start, end)

		// If no sentence boundary found, split at word boundary
		if chunkEnd == start {
			chunkEnd = findWordBoundary(content, start, end)
		}

		// If still no good boundary, just split at max size
		if chunkEnd == start {
			chunkEnd = end
		}

		chunks = append(chunks, content[start:chunkEnd])

		// Move start position back by overlap size to maintain context
		start = chunkEnd - OverlapSize
		if start < chunkEnd {
			start = chunkEnd
		}
	}

	return chunks
}

// findSentenceBoundary finds the last sentence-ending punctuation before the end position
func findSentenceBoundary(content string, start, end int) int {
	// Search backwards from end to find sentence boundary
	for i := end - 1; i > start; i-- {
		if isSentenceEnd(content, i) {
			// Include the punctuation and any trailing whitespace
			j := i + 1
			for j < len(content) && unicode.IsSpace(rune(content[j])) {
				j++
			}
			return j
		}
	}
	return start
}

// isSentenceEnd checks if the character at position i is a sentence-ending punctuation
func isSentenceEnd(content string, i int) bool {
	if i >= len(content) {
		return false
	}

	char := rune(content[i])
	
	// Check if it's a sentence-ending character
	if !strings.ContainsRune(SentenceEndChars, char) {
		return false
	}

	// Make sure it's followed by whitespace or end of string
	if i+1 < len(content) {
		nextChar := rune(content[i+1])
		return unicode.IsSpace(nextChar) || i+1 == len(content)
	}

	return true
}

// findWordBoundary finds the last word boundary (whitespace) before the end position
func findWordBoundary(content string, start, end int) int {
	// Search backwards from end to find whitespace
	for i := end - 1; i > start; i-- {
		if unicode.IsSpace(rune(content[i])) {
			// Skip consecutive whitespace
			j := i + 1
			for j < len(content) && unicode.IsSpace(rune(content[j])) {
				j++
			}
			return j
		}
	}
	return start
}
