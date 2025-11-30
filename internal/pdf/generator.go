package pdf

import (
	"bytes"

	"github.com/jung-kurt/gofpdf"
	"servicehealthchecker/internal/models"
)

type Generator struct{}

func NewGenerator() *Generator {
	return &Generator{}
}

func (g *Generator) Generate(linkSets []models.LinkSet) ([]byte, error) {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()

	pdf.SetFont("Arial", "B", 16)
	pdf.Cell(190, 10, "Link Status Report")
	pdf.Ln(15)

	pdf.SetFont("Arial", "B", 10)
	pdf.SetFillColor(200, 200, 200)
	pdf.CellFormat(20, 8, "Set ID", "1", 0, "C", true, 0, "")
	pdf.CellFormat(120, 8, "URL", "1", 0, "C", true, 0, "")
	pdf.CellFormat(50, 8, "Status", "1", 0, "C", true, 0, "")
	pdf.Ln(-1)

	pdf.SetFont("Arial", "", 10)
	for _, ls := range linkSets {
		for _, link := range ls.Links {
			pdf.CellFormat(20, 7, intToString(ls.ID), "1", 0, "C", false, 0, "")
			pdf.CellFormat(120, 7, truncateString(link.URL, 50), "1", 0, "L", false, 0, "")
			pdf.CellFormat(50, 7, link.Status, "1", 0, "C", false, 0, "")
			pdf.Ln(-1)
		}
	}

	var buf bytes.Buffer
	err := pdf.Output(&buf)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func intToString(n int) string {
	if n == 0 {
		return "0"
	}

	var result []byte
	negative := n < 0
	if negative {
		n = -n
	}

	for n > 0 {
		result = append([]byte{byte('0' + n%10)}, result...)
		n /= 10
	}

	if negative {
		result = append([]byte{'-'}, result...)
	}

	return string(result)
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

