package services

import (
	"archive/zip"
	"bytes"
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"
	"time"

	"contaduria/mvp/shared"
)

func (s *AccountingService) ExportWeeklyLedgerXLSX(weekOffset int) (string, error) {
	transactions, start, end, err := s.weekTransactions(weekOffset)
	if err != nil {
		return "", err
	}
	summary, err := s.GetWeeklySummary(weekOffset)
	if err != nil {
		return "", err
	}

	book, err := buildWeeklyLedgerXLSX(transactions, summary, start, end)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(book), nil
}

func buildWeeklyLedgerXLSX(transactions []shared.Transaction, summary *shared.WeeklySummary, start, end time.Time) ([]byte, error) {
	periodEnd := end.AddDate(0, 0, -1)
	journal := journalSheet(transactions, summary, start, periodEnd)
	budget := budgetSheet(summary, start, periodEnd)

	files := map[string]string{
		"[Content_Types].xml":        contentTypesXML(),
		"_rels/.rels":                rootRelsXML(),
		"xl/workbook.xml":            workbookXML(),
		"xl/_rels/workbook.xml.rels": workbookRelsXML(),
		"xl/styles.xml":              stylesXML(),
		"xl/worksheets/sheet1.xml":   journal,
		"xl/worksheets/sheet2.xml":   budget,
	}

	buffer := &bytes.Buffer{}
	writer := zip.NewWriter(buffer)
	for name, body := range files {
		part, err := writer.Create(name)
		if err != nil {
			return nil, fmt.Errorf("create xlsx part %s: %w", name, err)
		}
		if _, err := part.Write([]byte(body)); err != nil {
			return nil, fmt.Errorf("write xlsx part %s: %w", name, err)
		}
	}
	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("close xlsx: %w", err)
	}
	return buffer.Bytes(), nil
}

func journalSheet(transactions []shared.Transaction, summary *shared.WeeklySummary, start, periodEnd time.Time) string {
	rows := [][]cell{
		{textCell("Contaduria Personal", 1)},
		{textCell("Libro diario semanal (base de efectivo)", 1)},
		{textCell("Periodo", 2), textCell(start.Format("2006-01-02")+" a "+periodEnd.Format("2006-01-02"), 0)},
		{textCell("Moneda", 2), textCell("USD", 0)},
		{textCell("Emitido", 2), textCell(time.Now().In(start.Location()).Format("2006-01-02 15:04"), 0)},
		{},
		{
			textCell("Fecha", 2),
			textCell("Cuenta", 2),
			textCell("Tipo", 2),
			textCell("Concepto", 2),
			textCell("Debe", 2),
			textCell("Haber", 2),
			textCell("Saldo", 2),
		},
	}

	var saldo float64
	var totalDebe, totalHaber float64
	for _, tx := range transactions {
		debe, haber := 0.0, 0.0
		if tx.Type == shared.TransactionTypeExpense {
			debe = tx.Amount
			saldo -= tx.Amount
			totalDebe += tx.Amount
		} else {
			haber = tx.Amount
			saldo += tx.Amount
			totalHaber += tx.Amount
		}
		rows = append(rows, []cell{
			textCell(tx.CreatedAt.In(start.Location()).Format("2006-01-02 15:04"), 0),
			textCell(tx.Category, 0),
			textCell(movementLabel(tx.Type), 0),
			textCell(tx.Description, 0),
			moneyCell(debe),
			moneyCell(haber),
			moneyCell(saldo),
		})
	}

	rows = append(rows,
		[]cell{},
		[]cell{
			textCell("Totales del periodo", 2),
			textCell("", 0),
			textCell("", 0),
			textCell("", 0),
			moneyCell(totalDebe),
			moneyCell(totalHaber),
			moneyCell(saldo),
		},
		[]cell{},
		[]cell{textCell("Nota para revision", 2)},
		[]cell{textCell("Debe = salida de fondos (gasto). Haber = entrada de fondos (ingreso). El saldo es solo del periodo, no incluye saldo inicial.", 0)},
		[]cell{textCell(fmt.Sprintf("Fondos disponibles al emitir el reporte: %.2f. Presupuesto de la semana: %.2f.", summary.AvailableBalance, effectiveBudget(summary)), 0)},
	)
	return sheetXML(rows, []int{18, 22, 14, 36, 14, 14, 14})
}

func budgetSheet(summary *shared.WeeklySummary, start, periodEnd time.Time) string {
	budget := effectiveBudget(summary)
	rows := [][]cell{
		{textCell("Contaduria Personal", 1)},
		{textCell("Ejecucion presupuestaria semanal", 1)},
		{textCell("Periodo", 2), textCell(start.Format("2006-01-02")+" a "+periodEnd.Format("2006-01-02"), 0)},
		{},
		{
			textCell("Cuenta", 2),
			textCell("Presupuesto", 2),
			textCell("Ejecutado", 2),
			textCell("Variacion", 2),
			textCell("% ejecucion", 2),
			textCell("Observacion", 2),
		},
	}

	for _, item := range summary.Categories {
		variation := item.Budget - item.Spent
		note := "Sin presupuesto"
		if item.Budget > 0 {
			switch {
			case item.PercentUsed > 100:
				note = "Excedido"
			case item.PercentUsed > 80:
				note = "Cerca del tope"
			default:
				note = "Dentro del tope"
			}
		}
		rows = append(rows, []cell{
			textCell(item.CategoryName, 0),
			moneyCell(item.Budget),
			moneyCell(item.Spent),
			moneyCell(variation),
			textCell(fmt.Sprintf("%.0f%%", item.PercentUsed), 0),
			textCell(note, 0),
		})
	}

	rows = append(rows,
		[]cell{},
		[]cell{textCell("Resumen", 2)},
		[]cell{textCell("Gasto de la semana", 0), moneyCell(summary.TotalSpent)},
		[]cell{textCell("Presupuesto aplicado", 0), moneyCell(budget)},
		[]cell{textCell("Resultado vs semana anterior", 0), moneyCell(summary.DeltaVsPrev)},
		[]cell{textCell("Fondos disponibles", 0), moneyCell(summary.AvailableBalance)},
		[]cell{},
		[]cell{textCell("Uso sugerido", 2)},
		[]cell{textCell("Este libro permite revisar si el estudiante gasto dentro del tope, que cuentas se desviaron y si los fondos disponibles cubren el plan. No sustituye estados financieros formales.", 0)},
	)
	return sheetXML(rows, []int{24, 16, 16, 16, 14, 22})
}

func effectiveBudget(summary *shared.WeeklySummary) float64 {
	if summary.GlobalBudget > 0 {
		return summary.GlobalBudget
	}
	return summary.TotalBudget
}

type cell struct {
	value   string
	style   int
	number  bool
	numeric float64
}

func textCell(value string, style int) cell {
	return cell{value: value, style: style}
}

func moneyCell(value float64) cell {
	return cell{style: 3, number: true, numeric: value}
}

func sheetXML(rows [][]cell, widths []int) string {
	var body strings.Builder
	body.WriteString(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>`)
	body.WriteString(`<worksheet xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main">`)
	body.WriteString(`<cols>`)
	for i, width := range widths {
		body.WriteString(fmt.Sprintf(`<col min="%d" max="%d" width="%d" customWidth="1"/>`, i+1, i+1, width))
	}
	body.WriteString(`</cols><sheetData>`)
	for r, row := range rows {
		body.WriteString(fmt.Sprintf(`<row r="%d">`, r+1))
		for c, item := range row {
			ref := cellRef(c, r)
			style := ""
			if item.style > 0 {
				style = fmt.Sprintf(` s="%d"`, item.style)
			}
			if item.number {
				body.WriteString(fmt.Sprintf(`<c r="%s"%s><v>%s</v></c>`, ref, style, strconv.FormatFloat(item.numeric, 'f', 2, 64)))
				continue
			}
			if item.value == "" && item.style == 0 {
				continue
			}
			body.WriteString(fmt.Sprintf(`<c r="%s" t="inlineStr"%s><is><t>%s</t></is></c>`, ref, style, xmlEscape(item.value)))
		}
		body.WriteString(`</row>`)
	}
	body.WriteString(`</sheetData></worksheet>`)
	return body.String()
}

func cellRef(col, row int) string {
	name := ""
	n := col
	for {
		name = string(rune('A'+(n%26))) + name
		n = n/26 - 1
		if n < 0 {
			break
		}
	}
	return fmt.Sprintf("%s%d", name, row+1)
}

func xmlEscape(value string) string {
	replacer := strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;", `"`, "&quot;")
	return replacer.Replace(value)
}

func contentTypesXML() string {
	return `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types">
<Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/>
<Default Extension="xml" ContentType="application/xml"/>
<Override PartName="/xl/workbook.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.sheet.main+xml"/>
<Override PartName="/xl/worksheets/sheet1.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.worksheet+xml"/>
<Override PartName="/xl/worksheets/sheet2.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.worksheet+xml"/>
<Override PartName="/xl/styles.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.styles+xml"/>
</Types>`
}

func rootRelsXML() string {
	return `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
<Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument" Target="xl/workbook.xml"/>
</Relationships>`
}

func workbookXML() string {
	return `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<workbook xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main" xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships">
<sheets>
<sheet name="Libro diario" sheetId="1" r:id="rId1"/>
<sheet name="Presupuesto" sheetId="2" r:id="rId2"/>
</sheets>
</workbook>`
}

func workbookRelsXML() string {
	return `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
<Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/worksheet" Target="worksheets/sheet1.xml"/>
<Relationship Id="rId2" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/worksheet" Target="worksheets/sheet2.xml"/>
<Relationship Id="rId3" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/styles" Target="styles.xml"/>
</Relationships>`
}

func stylesXML() string {
	return `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<styleSheet xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main">
<fonts count="2"><font><sz val="11"/><name val="Calibri"/></font><font><b/><sz val="11"/><name val="Calibri"/></font></fonts>
<fills count="1"><fill><patternFill patternType="none"/></fill></fills>
<borders count="1"><border><left/><right/><top/><bottom/><diagonal/></border></borders>
<cellStyleXfs count="1"><xf numFmtId="0" fontId="0" fillId="0" borderId="0"/></cellStyleXfs>
<cellXfs count="4">
<xf numFmtId="0" fontId="0" fillId="0" borderId="0" xfId="0"/>
<xf numFmtId="0" fontId="1" fillId="0" borderId="0" xfId="0"/>
<xf numFmtId="0" fontId="1" fillId="0" borderId="0" xfId="0"/>
<xf numFmtId="4" fontId="0" fillId="0" borderId="0" xfId="0" applyNumberFormat="1"/>
</cellXfs>
</styleSheet>`
}
