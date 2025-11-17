package extraction

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"unicode"
)

// RTFExtractor handles RTF document extraction
type RTFExtractor struct{}

// NewRTFExtractor creates a new RTF extractor
func NewRTFExtractor() *RTFExtractor {
	return &RTFExtractor{}
}

// Extract extracts text from RTF files
func (e *RTFExtractor) Extract(ctx context.Context, data []byte) (string, error) {
	// Check for context cancellation
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	default:
	}

	// Validate RTF header
	if len(data) < 6 {
		return "", fmt.Errorf("%w: file too small to be a valid RTF", ErrCorruptedFile)
	}

	// Check for RTF magic string ({\rtf)
	header := string(data[:6])
	if !strings.HasPrefix(header, "{\\rtf") {
		return "", fmt.Errorf("%w: invalid RTF header - file may be corrupted or not an RTF", ErrCorruptedFile)
	}

	// Parse RTF and extract text
	text, err := e.parseRTF(ctx, data)
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrExtractionFailed, err)
	}

	// Normalize whitespace while preserving paragraph breaks
	text = normalizeWhitespace(text)

	return text, nil
}

// parseRTF parses RTF content and extracts plain text
func (e *RTFExtractor) parseRTF(ctx context.Context, data []byte) (string, error) {
	var result strings.Builder
	var controlWord strings.Builder
	var inControlWord bool
	var skipCount int
	var braceDepth int
	var inHeader bool

	reader := bytes.NewReader(data)

	for {
		// Check for context cancellation periodically
		select {
		case <-ctx.Done():
			return result.String(), ctx.Err()
		default:
		}

		b, err := reader.ReadByte()
		if err != nil {
			break // End of file
		}

		// Skip bytes if we're in a skip sequence
		if skipCount > 0 {
			skipCount--
			continue
		}

		switch b {
		case '{':
			braceDepth++
			// Check if we're entering a header group
			if braceDepth == 2 {
				// Peek ahead to see if this is a header group
				pos, _ := reader.Seek(0, 1)
				nextBytes := make([]byte, 10)
				n, _ := reader.Read(nextBytes)
				reader.Seek(pos, 0)

				nextStr := string(nextBytes[:n])
				if strings.HasPrefix(nextStr, "\\fonttbl") ||
					strings.HasPrefix(nextStr, "\\colortbl") ||
					strings.HasPrefix(nextStr, "\\stylesheet") ||
					strings.HasPrefix(nextStr, "\\info") {
					inHeader = true
				}
			}

		case '}':
			braceDepth--
			if braceDepth == 1 {
				inHeader = false
			}
			inControlWord = false
			controlWord.Reset()

		case '\\':
			// Start of control word or control symbol
			nextByte, err := reader.ReadByte()
			if err != nil {
				break
			}

			if nextByte == '\\' || nextByte == '{' || nextByte == '}' {
				// Escaped special character
				if !inHeader && braceDepth > 0 {
					result.WriteByte(nextByte)
				}
			} else if nextByte == '\'' {
				// Hex-encoded character \'XX
				hexBytes := make([]byte, 2)
				n, err := reader.Read(hexBytes)
				if err == nil && n == 2 {
					// Convert hex to character
					var charCode byte
					for _, hb := range hexBytes {
						charCode <<= 4
						if hb >= '0' && hb <= '9' {
							charCode |= hb - '0'
						} else if hb >= 'a' && hb <= 'f' {
							charCode |= hb - 'a' + 10
						} else if hb >= 'A' && hb <= 'F' {
							charCode |= hb - 'A' + 10
						}
					}
					if !inHeader && braceDepth > 0 && charCode >= 32 {
						result.WriteByte(charCode)
					}
				}
			} else if unicode.IsLetter(rune(nextByte)) {
				// Control word
				inControlWord = true
				controlWord.Reset()
				controlWord.WriteByte(nextByte)
			} else {
				// Control symbol
				e.handleControlSymbol(nextByte, &result, inHeader, braceDepth)
			}

		case ' ', '\r', '\n', '\t':
			if inControlWord {
				// Space terminates control word
				e.handleControlWord(controlWord.String(), &skipCount, &result, inHeader, braceDepth)
				inControlWord = false
				controlWord.Reset()
			} else if !inHeader && braceDepth > 0 {
				// Regular whitespace
				result.WriteByte(b)
			}

		case ';':
			if inControlWord {
				// Semicolon terminates control word
				e.handleControlWord(controlWord.String(), &skipCount, &result, inHeader, braceDepth)
				inControlWord = false
				controlWord.Reset()
			} else if !inHeader && braceDepth > 0 {
				result.WriteByte(b)
			}

		default:
			if inControlWord {
				if unicode.IsLetter(rune(b)) || unicode.IsDigit(rune(b)) || b == '-' {
					controlWord.WriteByte(b)
				} else {
					// Non-alphanumeric terminates control word
					e.handleControlWord(controlWord.String(), &skipCount, &result, inHeader, braceDepth)
					inControlWord = false
					controlWord.Reset()

					// Process the current byte as text
					if !inHeader && braceDepth > 0 {
						result.WriteByte(b)
					}
				}
			} else if !inHeader && braceDepth > 0 {
				// Regular text
				result.WriteByte(b)
			}
		}
	}

	return result.String(), nil
}

// handleControlWord processes RTF control words
func (e *RTFExtractor) handleControlWord(word string, skipCount *int, result *strings.Builder, inHeader bool, braceDepth int) {
	if word == "" {
		return
	}

	// Handle common control words that affect text output
	switch {
	case word == "par" || word == "line":
		// Paragraph or line break
		if !inHeader && braceDepth > 0 {
			result.WriteString("\n")
		}
	case word == "tab":
		// Tab character
		if !inHeader && braceDepth > 0 {
			result.WriteString("\t")
		}
	case strings.HasPrefix(word, "u"):
		// Unicode character \uN
		// The number after 'u' is the Unicode code point
		// Followed by a fallback character that we should skip
		*skipCount = 1
	case word == "~":
		// Non-breaking space
		if !inHeader && braceDepth > 0 {
			result.WriteString(" ")
		}
	case word == "-":
		// Optional hyphen
		if !inHeader && braceDepth > 0 {
			result.WriteString("-")
		}
	case word == "_":
		// Non-breaking hyphen
		if !inHeader && braceDepth > 0 {
			result.WriteString("-")
		}
	}
}

// handleControlSymbol processes RTF control symbols
func (e *RTFExtractor) handleControlSymbol(symbol byte, result *strings.Builder, inHeader bool, braceDepth int) {
	if inHeader || braceDepth == 0 {
		return
	}

	switch symbol {
	case '~':
		// Non-breaking space
		result.WriteString(" ")
	case '-':
		// Optional hyphen
		result.WriteString("-")
	case '_':
		// Non-breaking hyphen
		result.WriteString("-")
	}
}
