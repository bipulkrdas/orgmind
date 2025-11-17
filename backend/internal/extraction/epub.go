package extraction

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"strings"

	"golang.org/x/net/html"
)

// EPUBExtractor handles EPUB document extraction
type EPUBExtractor struct{}

// NewEPUBExtractor creates a new EPUB extractor
func NewEPUBExtractor() *EPUBExtractor {
	return &EPUBExtractor{}
}

// Extract extracts text from EPUB files
func (e *EPUBExtractor) Extract(ctx context.Context, data []byte) (string, error) {
	// Check for context cancellation
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	default:
	}

	// Validate EPUB header (EPUB files are ZIP archives)
	if len(data) < 4 {
		return "", fmt.Errorf("%w: file too small to be a valid EPUB", ErrCorruptedFile)
	}

	// Check for ZIP magic number (PK\x03\x04)
	if !bytes.HasPrefix(data, []byte("PK\x03\x04")) {
		return "", fmt.Errorf("%w: invalid EPUB header - file may be corrupted or not an EPUB", ErrCorruptedFile)
	}

	// Open the EPUB as a ZIP archive
	reader := bytes.NewReader(data)
	zipReader, err := zip.NewReader(reader, int64(len(data)))
	if err != nil {
		return "", fmt.Errorf("%w: failed to open EPUB archive - %v", ErrCorruptedFile, err)
	}

	// Find and parse the content.opf file to get reading order
	contentOPF, err := e.findContentOPF(zipReader)
	if err != nil {
		return "", fmt.Errorf("%w: failed to find content.opf - %v", ErrCorruptedFile, err)
	}

	// Parse the OPF to get spine (reading order)
	spine, err := e.parseSpine(contentOPF)
	if err != nil {
		return "", fmt.Errorf("%w: failed to parse spine - %v", ErrCorruptedFile, err)
	}

	// Extract text from chapters in order
	var result strings.Builder
	extractedChapters := 0

	for i, itemRef := range spine {
		// Check for context cancellation between chapters
		select {
		case <-ctx.Done():
			if result.Len() > 0 {
				return normalizeWhitespace(result.String()), fmt.Errorf("%w: extracted %d of %d chapters before timeout", ctx.Err(), extractedChapters, len(spine))
			}
			return "", ctx.Err()
		default:
		}

		// Extract text from this chapter
		chapterText, err := e.extractChapter(zipReader, itemRef)
		if err != nil {
			// Log error but continue with other chapters
			continue
		}

		if chapterText != "" {
			result.WriteString(chapterText)
			extractedChapters++

			// Add double newline between chapters
			if i < len(spine)-1 {
				result.WriteString("\n\n")
			}
		}
	}

	// If no text was extracted, return an error
	if result.Len() == 0 {
		return "", fmt.Errorf("%w: no text content found in EPUB", ErrExtractionFailed)
	}

	// Normalize whitespace while preserving paragraph breaks
	text := normalizeWhitespace(result.String())

	return text, nil
}

// findContentOPF locates the content.opf file in the EPUB
func (e *EPUBExtractor) findContentOPF(zipReader *zip.Reader) ([]byte, error) {
	// First, try to read container.xml to find the OPF location
	for _, file := range zipReader.File {
		if strings.HasSuffix(file.Name, "META-INF/container.xml") {
			rc, err := file.Open()
			if err != nil {
				continue
			}
			defer rc.Close()

			data, err := io.ReadAll(rc)
			if err != nil {
				continue
			}

			// Parse container.xml to find OPF path
			opfPath := e.parseContainerXML(data)
			if opfPath != "" {
				// Find and read the OPF file
				for _, f := range zipReader.File {
					if strings.HasSuffix(f.Name, opfPath) || f.Name == opfPath {
						rc, err := f.Open()
						if err != nil {
							continue
						}
						defer rc.Close()

						return io.ReadAll(rc)
					}
				}
			}
		}
	}

	// Fallback: search for any .opf file
	for _, file := range zipReader.File {
		if strings.HasSuffix(file.Name, ".opf") {
			rc, err := file.Open()
			if err != nil {
				continue
			}
			defer rc.Close()

			return io.ReadAll(rc)
		}
	}

	return nil, fmt.Errorf("content.opf not found")
}

// parseContainerXML extracts the OPF path from container.xml
func (e *EPUBExtractor) parseContainerXML(data []byte) string {
	type Container struct {
		Rootfiles struct {
			Rootfile []struct {
				FullPath string `xml:"full-path,attr"`
			} `xml:"rootfile"`
		} `xml:"rootfiles"`
	}

	var container Container
	if err := xml.Unmarshal(data, &container); err != nil {
		return ""
	}

	if len(container.Rootfiles.Rootfile) > 0 {
		return container.Rootfiles.Rootfile[0].FullPath
	}

	return ""
}

// parseSpine extracts the reading order from the OPF file
func (e *EPUBExtractor) parseSpine(opfData []byte) ([]string, error) {
	type Package struct {
		Manifest struct {
			Items []struct {
				ID   string `xml:"id,attr"`
				Href string `xml:"href,attr"`
			} `xml:"item"`
		} `xml:"manifest"`
		Spine struct {
			ItemRefs []struct {
				IDRef string `xml:"idref,attr"`
			} `xml:"itemref"`
		} `xml:"spine"`
	}

	var pkg Package
	if err := xml.Unmarshal(opfData, &pkg); err != nil {
		return nil, err
	}

	// Create a map of ID to href
	idToHref := make(map[string]string)
	for _, item := range pkg.Manifest.Items {
		idToHref[item.ID] = item.Href
	}

	// Build the spine in reading order
	var spine []string
	for _, itemRef := range pkg.Spine.ItemRefs {
		if href, exists := idToHref[itemRef.IDRef]; exists {
			spine = append(spine, href)
		}
	}

	return spine, nil
}

// extractChapter extracts text from a single chapter (XHTML file)
func (e *EPUBExtractor) extractChapter(zipReader *zip.Reader, chapterPath string) (string, error) {
	// Find the chapter file
	for _, file := range zipReader.File {
		if strings.HasSuffix(file.Name, chapterPath) || file.Name == chapterPath {
			rc, err := file.Open()
			if err != nil {
				return "", err
			}
			defer rc.Close()

			data, err := io.ReadAll(rc)
			if err != nil {
				return "", err
			}

			// Parse XHTML and extract text
			return e.extractTextFromHTML(data)
		}
	}

	return "", fmt.Errorf("chapter not found: %s", chapterPath)
}

// extractTextFromHTML extracts text from XHTML content
func (e *EPUBExtractor) extractTextFromHTML(htmlData []byte) (string, error) {
	doc, err := html.Parse(bytes.NewReader(htmlData))
	if err != nil {
		return "", err
	}

	var result strings.Builder
	var extractText func(*html.Node)

	extractText = func(n *html.Node) {
		if n.Type == html.TextNode {
			text := strings.TrimSpace(n.Data)
			if text != "" {
				result.WriteString(text)
				result.WriteString(" ")
			}
		}

		// Skip script and style tags
		if n.Type == html.ElementNode && (n.Data == "script" || n.Data == "style") {
			return
		}

		// Add newlines for block elements
		if n.Type == html.ElementNode {
			switch n.Data {
			case "p", "div", "br", "h1", "h2", "h3", "h4", "h5", "h6", "li":
				if result.Len() > 0 && !strings.HasSuffix(result.String(), "\n") {
					result.WriteString("\n")
				}
			}
		}

		for c := n.FirstChild; c != nil; c = c.NextSibling {
			extractText(c)
		}
	}

	extractText(doc)
	return result.String(), nil
}
