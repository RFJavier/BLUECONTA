package backend

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// EnsureDocsFiles writes the embedded legal documents and a user manual (PDF)
// into dataDir/docs on first run. Existing files are never overwritten.
func EnsureDocsFiles(dataDir, license, thirdParty string) error {
	docsDir := filepath.Join(dataDir, "docs")
	if err := os.MkdirAll(docsDir, 0o755); err != nil {
		return fmt.Errorf("create docs dir: %w", err)
	}

	// func EnsureDocsFiles(dataDir, license, thirdParty, manualMd string) error {
	// docsDir := filepath.Join(dataDir, "docs")
	// if err := os.MkdirAll(docsDir, 0o755); err != nil {
	// 	return fmt.Errorf("create docs dir: %w", err)
	// }

	writeIfMissing := func(path string, content []byte) error {
		if _, err := os.Stat(path); err == nil {
			return nil
		}
		return os.WriteFile(path, content, 0o644)
	}

	if err := writeIfMissing(filepath.Join(docsDir, "LICENSE.txt"), []byte(license)); err != nil {
		return fmt.Errorf("write LICENSE.txt: %w", err)
	}
	if err := writeIfMissing(filepath.Join(docsDir, "LICENCIAS_TERCEROS.md"), []byte(thirdParty)); err != nil {
		return fmt.Errorf("write LICENCIAS_TERCEROS.md: %w", err)
	}
	// if err := writeIfMissing(filepath.Join(docsDir, "MANUAL.pdf"), manualPDF(manualMd)); err != nil {
	// 	return fmt.Errorf("write MANUAL.pdf: %w", err)
	// }

	return nil
}

// manualPDF builds a minimal but valid PDF (Helvetica, WinAnsiEncoding) from
// the markdown manual text. It handles Spanish accents, line wrapping and
// pagination without external dependencies.
func manualPDF(md string) []byte {
	wrapped := wrapLines(markdownToLines(md), 88)

	const linesPerPage = 46
	var pages [][]string
	for i := 0; i < len(wrapped); i += linesPerPage {
		end := i + linesPerPage
		if end > len(wrapped) {
			end = len(wrapped)
		}
		pages = append(pages, wrapped[i:end])
	}
	if len(pages) == 0 {
		pages = [][]string{{"Manual de usuario"}}
	}

	var objects []string

	objects = append(objects, "1 0 obj\n<< /Type /Catalog /Pages 2 0 R >>\nendobj\n")

	kids := make([]string, len(pages))
	for i := range pages {
		kids[i] = fmt.Sprintf("%d 0 R", 5+2*i)
	}
	objects = append(objects, fmt.Sprintf("2 0 obj\n<< /Type /Pages /Kids [%s] /Count %d >>\nendobj\n", strings.Join(kids, " "), len(pages)))

	objects = append(objects, "3 0 obj\n<< /Type /Font /Subtype /Type1 /BaseFont /Helvetica /Encoding /WinAnsiEncoding >>\nendobj\n")

	for i, page := range pages {
		contentID := 4 + 2*i
		pageID := 5 + 2*i
		stream := pageContent(page)
		objects = append(objects, fmt.Sprintf("%d 0 obj\n<< /Length %d >>\nstream\n%s\nendstream\nendobj\n", contentID, len(stream), stream))
		objects = append(objects, fmt.Sprintf("%d 0 obj\n<< /Type /Page /Parent 2 0 R /MediaBox [0 0 612 792] /Resources << /Font << /F1 3 0 R >> >> /Contents %d 0 R >>\nendobj\n", pageID, contentID))
	}

	var b strings.Builder
	b.WriteString("%PDF-1.4\n")
	offsets := make([]int, 0, len(objects))
	for _, obj := range objects {
		offsets = append(offsets, b.Len())
		b.WriteString(obj)
	}
	xrefPos := b.Len()
	b.WriteString(fmt.Sprintf("xref\n0 %d\n", len(objects)+1))
	b.WriteString("0000000000 65535 f \n")
	for _, off := range offsets {
		b.WriteString(fmt.Sprintf("%010d 00000 n \n", off))
	}
	b.WriteString(fmt.Sprintf("trailer\n<< /Size %d /Root 1 0 R >>\nstartxref\n%d\n", len(objects)+1, xrefPos))
	b.WriteString("%%EOF\n")

	return []byte(b.String())
}

func pageContent(lines []string) string {
	var b strings.Builder
	b.WriteString("BT\n/F1 10 Tf\n14 TL\n50 742 Td\n")
	for _, line := range lines {
		b.WriteString("(" + pdfString(line) + ") Tj\nT*\n")
	}
	b.WriteString("ET")
	return b.String()
}

func markdownToLines(md string) []string {
	var out []string
	inCode := false
	for _, raw := range strings.Split(md, "\n") {
		line := strings.TrimRight(raw, " \t")
		if strings.HasPrefix(strings.TrimSpace(line), "```") {
			inCode = !inCode
			continue
		}
		if inCode {
			out = append(out, "  "+line)
			continue
		}
		line = strings.ReplaceAll(line, "**", "")
		line = strings.ReplaceAll(line, "`", "")
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "|") {
			line = strings.Trim(line, "|")
			line = strings.ReplaceAll(line, "|", "  ")
			if strings.Trim(line, "- ") == "" {
				continue
			}
		}
		line = strings.TrimLeft(line, "#")
		line = strings.TrimSpace(line)
		if line == "" {
			if len(out) > 0 && out[len(out)-1] != "" {
				out = append(out, "")
			}
			continue
		}
		out = append(out, line)
	}
	return out
}

func wrapLines(lines []string, max int) []string {
	var out []string
	for _, line := range lines {
		words := strings.Fields(line)
		if len(words) == 0 {
			out = append(out, "")
			continue
		}
		current := words[0]
		for _, w := range words[1:] {
			if len(current)+1+len(w) <= max {
				current += " " + w
			} else {
				out = append(out, current)
				current = w
			}
		}
		out = append(out, current)
	}
	return out
}

func pdfString(s string) string {
	var b strings.Builder
	for _, r := range s {
		switch r {
		case '\\':
			b.WriteString(`\\`)
		case '(':
			b.WriteString(`\(`)
		case ')':
			b.WriteString(`\)`)
		default:
			switch {
			case r >= 0xA0 && r <= 0xFF:
				b.WriteString(fmt.Sprintf("\\%03o", r&0xFF))
			case r == 0x2014 || r == 0x2013:
				b.WriteString(`\227`)
			case r < 128:
				b.WriteRune(r)
			default:
				b.WriteRune('?')
			}
		}
	}
	return b.String()
}
