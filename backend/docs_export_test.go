package backend

import (
	"strings"
	"testing"
)

func TestManualPDF(t *testing.T) {
	md := `# Manual

Texto con acentos: presupuesto, ahorro, gestion.

- Lista uno
- Lista dos

Otra seccion de prueba.
`
	pdf := manualPDF(md)
	s := string(pdf)

	if !strings.HasPrefix(s, "%PDF-") {
		t.Fatalf("no empieza con header PDF: %q", s[:20])
	}
	if !strings.HasSuffix(s, "%%EOF\n") {
		t.Fatalf("no termina con %%EOF")
	}
	if !strings.Contains(s, "/Type /Catalog") {
		t.Fatalf("falta el catalogo")
	}
	if !strings.Contains(s, "/Type /Font") {
		t.Fatalf("falta la fuente")
	}
	if strings.Contains(s, "(? )") {
		t.Fatalf("caracter no soportado convertido a '?'")
	}
}

func TestMarkdownToLines(t *testing.T) {
	lines := markdownToLines("# Titulo\n\n- item **negrita**\n\n| a | b |\n| - | - |\n| 1 | 2 |\n")
	joined := strings.Join(lines, "\n")
	if strings.Contains(joined, "**") || strings.Contains(joined, "`") {
		t.Fatalf("no limpio marcado: %q", joined)
	}
}
