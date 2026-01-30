package utils

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/ledongthuc/pdf"
)

func ExtractPaymentData(content string, providerName string) map[string]string {
	data := make(map[string]string)

	// Check if content is a PDF (starts with %PDF)
	if strings.HasPrefix(content, "%PDF") {
		extractFromPDF([]byte(content), data)
		return data
	}

	// Otherwise treat as HTML
	reader := strings.NewReader(content)
	doc, err := goquery.NewDocumentFromReader(reader)
	if err != nil {
		return data
	}

	switch providerName {
	case "CBE":
		extractCBEData(doc, data)
	case "TeleBirr":
		extractTeleBirrData(doc, data)
	default:
		extractTeleBirrData(doc, data)
	}

	return data
}

func extractFromPDF(pdfBytes []byte, data map[string]string) {
	reader := bytes.NewReader(pdfBytes)
	pdfReader, err := pdf.NewReader(reader, int64(len(pdfBytes)))
	if err != nil {
		fmt.Println("Error reading PDF:", err)
		return
	}

	var textBuilder strings.Builder
	for pageNum := 1; pageNum <= pdfReader.NumPage(); pageNum++ {
		page := pdfReader.Page(pageNum)
		if page.V.IsNull() {
			continue
		}
		text, err := page.GetPlainText(nil)
		if err != nil {
			continue
		}
		textBuilder.WriteString(text)
		textBuilder.WriteString("\n")
	}

	pdfText := textBuilder.String()
	fmt.Println("Extracted PDF Text:", pdfText)

	// Parse the extracted text for CBE payment details
	lines := strings.Split(pdfText, "\n")

	// CBE PDF uses label on one line, value on next line format
	// Create a map to find values after specific labels
	for i := 0; i < len(lines)-1; i++ {
		label := strings.TrimSpace(lines[i])
		value := strings.TrimSpace(lines[i+1])

		switch label {
		case "Payer":
			if data["payerName"] == "" {
				data["payerName"] = value
			}
		case "Receiver":
			if data["creditedPartyName"] == "" {
				data["creditedPartyName"] = value
			}
		case "Payment Date & Time":
			if data["paymentDate"] == "" {
				data["paymentDate"] = value
			}
		case "Reference No. (VAT Invoice No)":
			if data["invoiceNo"] == "" {
				data["invoiceNo"] = value
			}
		case "Reason / Type of service":
			if data["paymentReason"] == "" {
				data["paymentReason"] = value
			}
		case "Transferred Amount":
			if data["totalPaidAmount"] == "" {
				// Remove "ETB" suffix
				value = strings.TrimSuffix(value, " ETB")
				data["totalPaidAmount"] = value
			}
		}

		// Handle "Account" lines after Payer/Receiver
		if label == "Account" && i >= 2 {
			prevLabel := strings.TrimSpace(lines[i-2])
			if prevLabel == "Payer" {
				data["payerAccountNo"] = value
			} else if prevLabel == "Receiver" {
				data["creditedPartyAccountNo"] = value
			}
		}
	}

	// If invoice number still empty, try to find FT reference pattern
	if data["invoiceNo"] == "" {
		refPattern := regexp.MustCompile(`FT\d{7}[A-Z0-9]+`)
		fullText := strings.Join(lines, " ")
		if match := refPattern.FindString(fullText); match != "" {
			data["invoiceNo"] = match
		}
	}

	// CBE receipts are typically successful transactions
	if data["status"] == "" {
		data["status"] = "Successful"
	}
}

func extractCBEData(doc *goquery.Document, data map[string]string) {
	// CBE label mappings (adjust based on actual CBE HTML)
	labelMap := map[string]string{
		"Transaction Reference":   "invoiceNo",
		"Txn Reference":           "invoiceNo",
		"Reference":               "invoiceNo",
		"Amount":                  "totalPaidAmount",
		"Transaction Amount":      "totalPaidAmount",
		"Payer":                   "payerName",
		"Payer Name":              "payerName",
		"Sender Name":             "payerName",
		"Sender":                  "payerName",
		"Beneficiary":             "creditedPartyName",
		"Beneficiary Name":        "creditedPartyName",
		"Receiver Name":           "creditedPartyName",
		"Receiver":                "creditedPartyName",
		"Beneficiary Account":     "creditedPartyAccountNo",
		"Account Number":          "creditedPartyAccountNo",
		"Date":                    "paymentDate",
		"Transaction Date":        "paymentDate",
		"Value Date":              "paymentDate",
		"Status":                  "status",
		"Transaction Status":      "status",
		"Payment Type":            "paymentMode",
		"Transaction Type":        "paymentMode",
		"Reason":                  "paymentReason",
		"Narration":               "paymentReason",
		"Remark":                  "paymentReason",
	}

	// Try to extract from table rows (label: value format)
	doc.Find("tr").Each(func(i int, tr *goquery.Selection) {
		tds := tr.Find("td")
		if tds.Length() >= 2 {
			label := strings.TrimSpace(tds.Eq(0).Text())
			label = strings.TrimSuffix(label, ":")
			value := strings.TrimSpace(tds.Eq(1).Text())

			for mapLabel, key := range labelMap {
				if strings.EqualFold(label, mapLabel) || strings.Contains(strings.ToLower(label), strings.ToLower(mapLabel)) {
					if data[key] == "" {
						data[key] = value
					}
				}
			}
		}
	})

	// Try to extract from th/td pairs
	doc.Find("tr").Each(func(i int, tr *goquery.Selection) {
		th := tr.Find("th")
		td := tr.Find("td")
		if th.Length() > 0 && td.Length() > 0 {
			label := strings.TrimSpace(th.First().Text())
			label = strings.TrimSuffix(label, ":")
			value := strings.TrimSpace(td.First().Text())

			for mapLabel, key := range labelMap {
				if strings.EqualFold(label, mapLabel) || strings.Contains(strings.ToLower(label), strings.ToLower(mapLabel)) {
					if data[key] == "" {
						data[key] = value
					}
				}
			}
		}
	})

	// Try to extract from div/span with class patterns common in CBE
	doc.Find("div, span, p").Each(func(i int, s *goquery.Selection) {
		text := strings.TrimSpace(s.Text())
		for mapLabel, key := range labelMap {
			pattern := fmt.Sprintf(`(?i)%s\s*[:\-]?\s*(.+)`, regexp.QuoteMeta(mapLabel))
			re := regexp.MustCompile(pattern)
			if matches := re.FindStringSubmatch(text); len(matches) > 1 {
				value := strings.TrimSpace(matches[1])
				if data[key] == "" && value != "" {
					data[key] = value
				}
			}
		}
	})

	// Try to find status from common patterns
	if data["status"] == "" {
		htmlLower := strings.ToLower(doc.Text())
		if strings.Contains(htmlLower, "successful") || strings.Contains(htmlLower, "success") {
			data["status"] = "Successful"
		} else if strings.Contains(htmlLower, "pending") {
			data["status"] = "Pending"
		} else if strings.Contains(htmlLower, "failed") {
			data["status"] = "Failed"
		}
	}
}

func extractTeleBirrData(doc *goquery.Document, data map[string]string) {
	labelMap := map[string]string{
		"Payer Name":                "payerName",
		"Payment Mode":              "paymentMode",
		"Payment Reason":            "paymentReason",
		"Payment date":              "paymentDate",
		"Invoice No.":               "invoiceNo",
		"Total Paid Amount":         "totalPaidAmount",
		"transaction status":        "status",
		"Credited Party name":       "creditedPartyName",
		"Credited party account no": "creditedPartyAccountNo",
	}

	// Special: Invoice table (3-column header + data row)
	doc.Find("tr").Each(func(i int, tr *goquery.Selection) {
		tds := tr.Find("td")
		if tds.Length() == 3 {
			label := strings.TrimSpace(tds.Eq(0).Text())
			if strings.Contains(label, "Invoice No.") {
				nextTr := tr.Next()
				if nextTr.Length() > 0 {
					valTds := nextTr.Find("td")
					if valTds.Length() >= 3 {
						data["invoiceNo"] = strings.TrimSpace(valTds.Eq(0).Text())
						data["paymentDate"] = strings.TrimSpace(valTds.Eq(1).Text())
						data["totalPaidAmount"] = strings.TrimSpace(valTds.Eq(2).Text())
					}
				}
				return
			}
		}
	})

	// Regular 2-column rows
	doc.Find("tr").Each(func(i int, tr *goquery.Selection) {
		tds := tr.Find("td")
		if tds.Length() < 2 {
			return
		}

		raw := strings.TrimSpace(tds.First().Text())
		parts := strings.Split(raw, "/")
		label := strings.TrimSpace(parts[len(parts)-1])

		if key, ok := labelMap[label]; ok {
			value := strings.TrimSpace(tds.Eq(1).Text())
			data[key] = value
		}
	})
}
func ParsePaymentDate(dateStr string) time.Time {
	// Try multiple date formats
	formats := []string{
		"02-01-2006 15:04:05",         // Telebirr: 06-11-2025 11:40:12
		"1/2/2006, 3:04:05 PM",        // CBE: 1/6/2026, 12:44:00 PM
		"2/1/2006, 3:04:05 PM",        // CBE alternate
		"01/02/2006, 3:04:05 PM",      // CBE with leading zeros
		"2006-01-02 15:04:05",         // ISO format
		"02/01/2006 15:04:05",         // Another common format
	}

	for _, format := range formats {
		if t, err := time.Parse(format, dateStr); err == nil {
			return t
		}
	}

	return time.Time{}
}