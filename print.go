package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"
	"sync"

	"github.com/phpdave11/gofpdf"
	qrcode "github.com/skip2/go-qrcode"
)

const receiptWidthMM = 80.0

type FlexString string

func (f *FlexString) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		*f = ""
		return nil
	}

	var text string
	if err := json.Unmarshal(data, &text); err == nil {
		*f = FlexString(text)
		return nil
	}

	var number json.Number
	if err := json.Unmarshal(data, &number); err == nil {
		*f = FlexString(number.String())
		return nil
	}

	var boolean bool
	if err := json.Unmarshal(data, &boolean); err == nil {
		*f = FlexString(strconv.FormatBool(boolean))
		return nil
	}

	return fmt.Errorf("unsupported string value: %s", string(data))
}

func (f FlexString) String() string {
	return string(f)
}

type FlexFloat float64

func (f *FlexFloat) UnmarshalJSON(data []byte) error {
	if string(data) == "null" || string(data) == `""` {
		*f = 0
		return nil
	}

	var number float64
	if err := json.Unmarshal(data, &number); err == nil {
		*f = FlexFloat(number)
		return nil
	}

	var text string
	if err := json.Unmarshal(data, &text); err == nil {
		if text == "" {
			*f = 0
			return nil
		}
		parsed, parseErr := strconv.ParseFloat(text, 64)
		if parseErr != nil {
			return parseErr
		}
		*f = FlexFloat(parsed)
		return nil
	}

	return fmt.Errorf("unsupported number value: %s", string(data))
}

func (f FlexFloat) String() string {
	return formatAmount(float64(f))
}

func (f FlexFloat) IsPositive() bool {
	return float64(f) > 0
}

type OrderItem struct {
	ItemName FlexString `json:"ItemName"`
	Qty      FlexString `json:"Qty"`
	Rate     FlexFloat  `json:"Rate"`
	Amount   FlexFloat  `json:"Amount"`
}

type Order struct {
	BillNo           FlexString  `json:"BillNo"`
	CafeName         FlexString  `json:"CafeName"`
	CanteenName      FlexString  `json:"CanteenName"`
	VendorName       FlexString  `json:"VendorName"`
	DateTime         FlexString  `json:"DateTime"`
	Address1         FlexString  `json:"Address1"`
	Address2         FlexString  `json:"Address2"`
	Address3         FlexString  `json:"Address3"`
	GSTN             FlexString  `json:"GSTN"`
	FssaiNo          FlexString  `json:"FssaiNo"`
	Discount         FlexFloat   `json:"Discount"`
	Note             FlexString  `json:"Note"`
	SalesDetList     []OrderItem `json:"SalesDetList"`
	SubTotal         FlexFloat   `json:"SubTotal"`
	ParcelCharges    FlexFloat   `json:"ParcelCharges"`
	CGSTAmt          FlexFloat   `json:"CGSTAmt"`
	SGSTAmt          FlexFloat   `json:"SGSTAmt"`
	CGSTPercentage   FlexFloat   `json:"CGSTPercentage"`
	SGSTPercentage   FlexFloat   `json:"SGSTPercentage"`
	GrandTotal       FlexFloat   `json:"GrandTotal"`
	TotQty           FlexFloat   `json:"TotQty"`
	GSTInfo          FlexString  `json:"GSTInfo"`
	EnableKOTPrint   bool        `json:"ENABLE_KOT_PRINT"`
	QRCode           FlexString  `json:"QRCode"`
	AmountInWords    FlexString  `json:"AmountInWords"`
	PaymentInfo      FlexString  `json:"PaymentInfo"`
	IsDuplicatePrint bool        `json:"isDuplicatePrint"`
}

type PrintOrderRequest struct {
	Order Order `json:"order"`
}

type PrintersResponse struct {
	Printers          []string `json:"printers"`
	SelectedPrinter   string   `json:"selectedPrinter"`
	SelectedPaperSize string   `json:"selectedPaperSize"`
}

var testOrder = Order{
	BillNo:         "3454",
	CafeName:       "Siemens Gamesa",
	CanteenName:    "Breakfast & Veg Meal",
	VendorName:     "Cheftalk Caterers",
	DateTime:       "22/05/2025 08:45 AM",
	Address1:       "Block A, Tech Park Campus, Tumkur Road",
	Address2:       "Bengaluru, Karnataka - 560073",
	Address3:       "",
	GSTN:           "29AAFCC7584D1ZE",
	FssaiNo:        "10014011000023",
	Discount:       50,
	Note:           "No onions or garlic, please. Deliver by 9 AM",
	SubTotal:       500.89,
	ParcelCharges:  10.67,
	CGSTAmt:        12.51,
	SGSTAmt:        12.51,
	CGSTPercentage: 12.53,
	SGSTPercentage: 12.53,
	GrandTotal:     485.18,
	TotQty:         7,
	GSTInfo:        "All rates are included GST",
	EnableKOTPrint: true,
	QRCode:         "https://ctuat.tricubeinnosoft.com/OrderFeedback?order_id=3454",
	AmountInWords:  "Rupees Four Hundred Eighty-Five Only",
	PaymentInfo:    "Pay at Counter",
	SalesDetList: []OrderItem{
		{ItemName: "Idli Vada Combo Idli Vada Combo", Qty: "2", Rate: 60.13, Amount: 120.12},
		{ItemName: "Masala Dosa", Qty: "3", Rate: 50.34, Amount: 150.45},
		{ItemName: "Upma + Coffee Set", Qty: "2", Rate: 115.67, Amount: 230.76},
	},
}

var printerConfig = struct {
	sync.RWMutex
	selectedPrinter   string
	selectedPaperSize string
}{
	selectedPaperSize: "80mm",
}

func setSelectedPrinter(printerName string) {
	printerConfig.Lock()
	defer printerConfig.Unlock()
	printerConfig.selectedPrinter = strings.TrimSpace(printerName)
	log.Printf("selected printer set to %q", printerConfig.selectedPrinter)
}

func getSelectedPrinter() string {
	printerConfig.RLock()
	defer printerConfig.RUnlock()
	return printerConfig.selectedPrinter
}

func setSelectedPaperSize(paperSize string) {
	printerConfig.Lock()
	defer printerConfig.Unlock()
	printerConfig.selectedPaperSize = strings.TrimSpace(paperSize)
	log.Printf("selected paper size set to %q", printerConfig.selectedPaperSize)
}

func getSelectedPaperSize() string {
	printerConfig.RLock()
	defer printerConfig.RUnlock()
	return printerConfig.selectedPaperSize
}

func printSimpleTestPage() error {
	pdf := newReceiptPDF(60)
	pdf.SetFont("Arial", "B", 14)
	pdf.CellFormat(0, 8, "Test print", "", 1, "C", false, 0, "")
	pdf.Ln(4)
	pdf.SetFont("Arial", "", 10)
	pdf.CellFormat(0, 6, "Cheftalk Go printer bridge", "", 1, "C", false, 0, "")
	return submitPDF(pdf, "test-print.pdf")
}

func printReceiptPDF(order Order) error {
	pdf := newReceiptPDF(estimateReceiptHeight(order))
	pdf.SetFont("Arial", "", 10)

	if order.IsDuplicatePrint {
		centerText(pdf, "Duplicate", "B", 11)
	}

	centerText(pdf, "Order ID: "+order.BillNo.String(), "B", 12)
	if order.VendorName.String() != "" {
		centerText(pdf, order.VendorName.String(), "B", 11)
	}
	centerText(pdf, order.CafeName.String(), "B", 11)
	if order.Address1.String() != "" {
		centerText(pdf, order.Address1.String(), "B", 10)
	}
	if order.Address2.String() != "" {
		centerText(pdf, order.Address2.String(), "B", 10)
	}
	if order.CanteenName.String() != "" {
		centerText(pdf, order.CanteenName.String(), "B", 11)
	}

	if info := receiptInfoLines(order); len(info) > 0 {
		for _, line := range info {
			leftText(pdf, line, "", 9)
		}
	}

	drawDivider(pdf)
	drawReceiptHeader(pdf)
	for _, item := range order.SalesDetList {
		drawReceiptItemRow(pdf, item)
	}
	drawDivider(pdf)

	drawAmountRowIfPositive(pdf, "SubTotal", order.SubTotal)
	drawAmountRowIfPositive(pdf, fmt.Sprintf("SGST(%s%%)", order.SGSTPercentage.String()), order.SGSTAmt)
	drawAmountRowIfPositive(pdf, fmt.Sprintf("CGST(%s%%)", order.CGSTPercentage.String()), order.CGSTAmt)
	drawAmountRowIfPositive(pdf, "Discount", order.Discount)
	drawAmountRowIfPositive(pdf, "Parcel Charges", order.ParcelCharges)
	drawAmountRowIfPositive(pdf, "Grand Total", order.GrandTotal)
	if order.PaymentInfo.String() != "" {
		drawTextValueRow(pdf, "Payment", order.PaymentInfo.String())
	}

	drawDivider(pdf)

	for _, text := range []string{
		order.AmountInWords.String(),
		order.GSTInfo.String(),
		noteText(order.Note.String()),
	} {
		if text == "" {
			continue
		}
		leftText(pdf, text, "", 9)
	}

	if order.QRCode.String() != "" {
		if err := addQRCode(pdf, order.QRCode.String()); err != nil {
			return err
		}
		centerText(pdf, "Kindly give your valuable feedback", "", 9)
	}

	centerText(pdf, "Thank you! Visit Again.", "B", 10)
	return submitPDF(pdf, "receipt-"+safeFilePart(order.BillNo.String())+".pdf")
}

func printKOTPDF(order Order) error {
	pdf := newReceiptPDF(estimateKOTHeight(order))
	pdf.SetFont("Arial", "", 10)

	if order.IsDuplicatePrint {
		centerText(pdf, "Duplicate", "B", 11)
	}

	centerText(pdf, "Order ID: "+order.BillNo.String(), "B", 12)
	centerText(pdf, order.CanteenName.String(), "B", 11)
	centerText(pdf, order.DateTime.String(), "", 10)
	drawDivider(pdf)

	drawKOTHeader(pdf)
	for _, item := range order.SalesDetList {
		drawKOTItemRow(pdf, item)
	}
	drawDivider(pdf)
	centerText(pdf, "Thank you! Visit Again.", "B", 10)

	return submitPDF(pdf, "kot-"+safeFilePart(order.BillNo.String())+".pdf")
}

func newReceiptPDF(heightMM float64) *gofpdf.Fpdf {
	pdf := gofpdf.NewCustom(&gofpdf.InitType{
		UnitStr: "mm",
		Size:    gofpdf.SizeType{Wd: receiptWidthMM, Ht: heightMM},
	})
	pdf.SetMargins(4, 4, 4)
	pdf.SetAutoPageBreak(false, 0)
	pdf.AddPage()
	pdf.SetFont("Arial", "", 10)
	return pdf
}

func estimateReceiptHeight(order Order) float64 {
	height := 115.0 + float64(len(order.SalesDetList))*8.0
	if order.QRCode.String() != "" {
		height += 38
	}
	for _, value := range []string{
		order.Address1.String(),
		order.Address2.String(),
		order.Address3.String(),
		order.AmountInWords.String(),
		order.GSTInfo.String(),
		order.Note.String(),
	} {
		if value != "" {
			height += 6
		}
	}
	if height < 180 {
		height = 180
	}
	return height
}

func estimateKOTHeight(order Order) float64 {
	height := 55.0 + float64(len(order.SalesDetList))*8.0
	if height < 110 {
		height = 110
	}
	return height
}

func addQRCode(pdf *gofpdf.Fpdf, qrValue string) error {
	png, err := qrcode.Encode(qrValue, qrcode.Medium, 128)
	if err != nil {
		return err
	}

	imageName := "qr-code"
	options := gofpdf.ImageOptions{ImageType: "PNG"}
	pdf.RegisterImageOptionsReader(imageName, options, bytes.NewReader(png))

	x := (receiptWidthMM - 22) / 2
	y := pdf.GetY() + 2
	pdf.ImageOptions(imageName, x, y, 22, 22, false, options, 0, "")
	pdf.SetY(y + 24)
	return nil
}

func receiptInfoLines(order Order) []string {
	lines := make([]string, 0, 3)
	if order.DateTime.String() != "" {
		lines = append(lines, "Date: "+order.DateTime.String())
	}
	if order.GSTN.String() != "" {
		lines = append(lines, "GSTIN: "+order.GSTN.String())
	}
	if order.FssaiNo.String() != "" {
		lines = append(lines, "FSSAI: "+order.FssaiNo.String())
	}
	return lines
}

func drawReceiptHeader(pdf *gofpdf.Fpdf) {
	pdf.SetFont("Arial", "B", 9)
	pdf.CellFormat(36, 6, "Item Name", "", 0, "L", false, 0, "")
	pdf.CellFormat(10, 6, "Qty", "", 0, "C", false, 0, "")
	pdf.CellFormat(14, 6, "Rate", "", 0, "C", false, 0, "")
	pdf.CellFormat(16, 6, "Amount", "", 1, "R", false, 0, "")
	pdf.SetFont("Arial", "", 9)
}

func drawKOTHeader(pdf *gofpdf.Fpdf) {
	pdf.SetFont("Arial", "B", 9)
	pdf.CellFormat(56, 6, "Item Name", "", 0, "L", false, 0, "")
	pdf.CellFormat(16, 6, "Qty", "", 1, "R", false, 0, "")
	pdf.SetFont("Arial", "", 9)
}

func drawReceiptItemRow(pdf *gofpdf.Fpdf, item OrderItem) {
	const lineHeight = 4.5
	const nameWidth = 36.0
	const qtyWidth = 10.0
	const rateWidth = 14.0
	const amountWidth = 16.0

	x, y := pdf.GetXY()
	lines := pdf.SplitLines([]byte(item.ItemName.String()), nameWidth)
	rowHeight := float64(len(lines)) * lineHeight
	if rowHeight < lineHeight {
		rowHeight = lineHeight
	}

	pdf.MultiCell(nameWidth, lineHeight, item.ItemName.String(), "", "L", false)
	pdf.SetXY(x+nameWidth, y)
	pdf.CellFormat(qtyWidth, rowHeight, item.Qty.String(), "", 0, "C", false, 0, "")
	pdf.CellFormat(rateWidth, rowHeight, item.Rate.String(), "", 0, "C", false, 0, "")
	pdf.CellFormat(amountWidth, rowHeight, item.Amount.String(), "", 0, "R", false, 0, "")
	pdf.SetXY(x, y+rowHeight)
}

func drawKOTItemRow(pdf *gofpdf.Fpdf, item OrderItem) {
	const lineHeight = 4.5
	const nameWidth = 56.0
	const qtyWidth = 16.0

	x, y := pdf.GetXY()
	lines := pdf.SplitLines([]byte(item.ItemName.String()), nameWidth)
	rowHeight := float64(len(lines)) * lineHeight
	if rowHeight < lineHeight {
		rowHeight = lineHeight
	}

	pdf.MultiCell(nameWidth, lineHeight, item.ItemName.String(), "", "L", false)
	pdf.SetXY(x+nameWidth, y)
	pdf.CellFormat(qtyWidth, rowHeight, item.Qty.String(), "", 0, "R", false, 0, "")
	pdf.SetXY(x, y+rowHeight)
}

func drawAmountRowIfPositive(pdf *gofpdf.Fpdf, label string, amount FlexFloat) {
	if !amount.IsPositive() {
		return
	}
	drawTextValueRow(pdf, label, amount.String())
}

func drawTextValueRow(pdf *gofpdf.Fpdf, label string, value string) {
	pdf.SetFont("Arial", "B", 9)
	pdf.CellFormat(40, 5, label, "", 0, "L", false, 0, "")
	pdf.CellFormat(32, 5, value, "", 1, "R", false, 0, "")
	pdf.SetFont("Arial", "", 9)
}

func drawDivider(pdf *gofpdf.Fpdf) {
	y := pdf.GetY() + 1
	pdf.Line(4, y, receiptWidthMM-4, y)
	pdf.SetY(y + 2)
}

func centerText(pdf *gofpdf.Fpdf, text, style string, size float64) {
	pdf.SetFont("Arial", style, size)
	pdf.MultiCell(0, 5, text, "", "C", false)
}

func leftText(pdf *gofpdf.Fpdf, text, style string, size float64) {
	pdf.SetFont("Arial", style, size)
	pdf.MultiCell(0, 5, text, "", "L", false)
}

func noteText(note string) string {
	if note == "" {
		return ""
	}
	return "Note: " + note
}

func safeFilePart(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return "print-job"
	}
	value = strings.ReplaceAll(value, "/", "-")
	value = strings.ReplaceAll(value, " ", "-")
	return value
}

func formatAmount(value float64) string {
	text := strconv.FormatFloat(value, 'f', 2, 64)
	text = strings.TrimRight(text, "0")
	text = strings.TrimRight(text, ".")
	if text == "" {
		return "0"
	}
	return text
}
