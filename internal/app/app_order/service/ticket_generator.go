package service

import (
	"bytes"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"strings"
	"time"

	orderEntity "rakit-tiket-be/pkg/entity/app_order"
	regEntity "rakit-tiket-be/pkg/entity/app_registrant"
	ticketEntity "rakit-tiket-be/pkg/entity/app_ticket"
	"rakit-tiket-be/pkg/util"

	"github.com/SebastiaanKlippert/go-wkhtmltopdf"
	"gitlab.com/threetopia/envgo"
)

type TicketAttachment struct {
	FileName string
	Data     []byte
}

type EventDynamicData struct {
	EventName      string
	EventDate      string
	EventTimeStart string
	EventTimeEnd   string
	EventLocation  string
}

type TicketTemplateData struct {
	EventName      string
	OwnerName      string
	TicketTitle    string
	TicketPrice    string
	OrderNumber    string
	PaymentTime    string
	PaymentStatus  string
	Amount         string
	RegistrantName string
	EventDate      string
	EventTimeStart string
	EventTimeEnd   string
	EventLocation  string
	QRCodePath     string
	CurrentYear    string
}

// Helper Format Rupiah
func formatRupiah(amount float64) string {
	s := fmt.Sprintf("%.0f", amount)
	var res string
	for i, v := range s {
		if i > 0 && (len(s)-i)%3 == 0 {
			res += "."
		}
		res += string(v)
	}
	return "Rp " + res
}

func GenerateTicketsPDF(
	order orderEntity.Order,
	registrant regEntity.Registrant,
	attendees []regEntity.Attendee,
	ticketMap map[string]ticketEntity.Ticket,
	eventData EventDynamicData,
) ([]TicketAttachment, error) {

	var attachments []TicketAttachment

	tmpl, err := template.New("ticket_pdf").Parse(ticketPDFTemplate)
	if err != nil {
		return nil, fmt.Errorf("failed to parse template: %v", err)
	}

	type Owner struct {
		Name     string
		TicketID string
	}
	owners := []Owner{}

	if registrant.TicketID != nil {
		owners = append(owners, Owner{Name: registrant.Name, TicketID: string(*registrant.TicketID)})
	}
	for _, att := range attendees {
		owners = append(owners, Owner{Name: att.Name, TicketID: string(att.TicketID)})
	}

	// Persiapkan Folder Asset Storage (Supaya Admin Bisa Download)
	ticketDir := filepath.Join(envgo.GetString("APP_FILE_PATH", "./assets/app_file"), "tickets")
	_ = os.MkdirAll(ticketDir, 0755)

	for _, owner := range owners {
		ticketInfo, exists := ticketMap[owner.TicketID]
		if !exists {
			continue
		}

		// Generate QR Code file for PDF rendering
		safeName := strings.ReplaceAll(owner.Name, " ", "_")
		qrFileName := fmt.Sprintf("qr-%s-%s.png", order.OrderNumber, safeName)
		qrDir := filepath.Join(ticketDir, "qrcodes")
		qrFilePath := filepath.Join(qrDir, qrFileName)

		_, err = util.GenerateQRCodeFile(order.OrderNumber, 300, qrFilePath)
		if err != nil {
			return nil, fmt.Errorf("failed to generate QR: %v", err)
		}

		paymentTimeStr := "-"
		if order.PaymentTime != nil {
			paymentTimeStr = order.PaymentTime.Format("02 Jan 2006 15:04")
		}

		// Mapping Data ke Struct
		data := TicketTemplateData{
			EventName:      strings.ToUpper(eventData.EventName),
			OwnerName:      strings.ToUpper(owner.Name),
			TicketTitle:    strings.ToUpper(ticketInfo.Title),
			TicketPrice:    formatRupiah(ticketInfo.Price),
			OrderNumber:    strings.ToUpper(order.OrderNumber),
			PaymentTime:    paymentTimeStr,
			PaymentStatus:  strings.ToUpper(order.PaymentStatus),
			Amount:         formatRupiah(order.Amount),
			RegistrantName: strings.ToUpper(registrant.Name),
			EventDate:      eventData.EventDate,
			EventTimeStart: eventData.EventTimeStart,
			EventTimeEnd:   eventData.EventTimeEnd,
			EventLocation:  eventData.EventLocation,
			QRCodePath:     qrFilePath,
			CurrentYear:    time.Now().Format("2006"),
		}

		var renderedHTML bytes.Buffer
		if err := tmpl.Execute(&renderedHTML, data); err != nil {
			return nil, fmt.Errorf("failed to render HTML: %v", err)
		}

		// PROSES RENDER HTML KE PDF
		pdfg, err := wkhtmltopdf.NewPDFGenerator()
		if err != nil {
			return nil, fmt.Errorf("failed to init pdf generator: %v (pastikan wkhtmltopdf terinstal di OS)", err)
		}

		page := wkhtmltopdf.NewPageReader(bytes.NewReader(renderedHTML.Bytes()))
		page.EnableLocalFileAccess.Set(true)
		pdfg.AddPage(page)

		// Set layout PDF
		pdfg.MarginLeft.Set(0)
		pdfg.MarginRight.Set(0)
		pdfg.MarginTop.Set(0)
		pdfg.MarginBottom.Set(0)
		pdfg.PageSize.Set(wkhtmltopdf.PageSizeA4)

		if err := pdfg.Create(); err != nil {
			return nil, fmt.Errorf("failed to generate PDF: %v", err)
		}

		pdfBytes := pdfg.Bytes()

		// SIMPAN KE FILE ASSET
		fileName := fmt.Sprintf("E-Voucher-%s-%s.pdf", order.OrderNumber, safeName)
		filePath := filepath.Join(ticketDir, fileName)

		if err := os.WriteFile(filePath, pdfBytes, 0644); err != nil {
			fmt.Printf("Warning: failed to save PDF to asset folder: %v\n", err)
		}

		// Masukkan ke Lampiran Email
		attachments = append(attachments, TicketAttachment{
			FileName: fileName,
			Data:     pdfBytes,
		})
	}

	return attachments, nil
}
