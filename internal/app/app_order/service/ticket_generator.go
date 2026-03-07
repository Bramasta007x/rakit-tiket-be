package service

import (
	"bytes"
	"fmt"
	"html/template"
	"strings"

	orderEntity "rakit-tiket-be/pkg/entity/app_order"
	regEntity "rakit-tiket-be/pkg/entity/app_registrant"
	ticketEntity "rakit-tiket-be/pkg/entity/app_ticket"
	"rakit-tiket-be/pkg/util"
)

// TicketAttachment representasi file tiket yang siap dilampirkan ke email
type TicketAttachment struct {
	FileName string
	HTMLData []byte
}

// TicketData struct untuk disuntikkan ke HTML Template
type TicketData struct {
	OrderNumber  string
	EventName    string
	OwnerName    string
	TicketTitle  string
	TicketType   string
	QRCodeBase64 string
}

func GenerateTickets(
	order orderEntity.Order,
	registrant regEntity.Registrant,
	attendees []regEntity.Attendee,
	ticketMap map[string]ticketEntity.Ticket,
	eventName string,
) ([]TicketAttachment, error) {

	var attachments []TicketAttachment

	// Parsing HTML Template
	tmpl, err := template.New("ticket").Parse(ticketHTMLTemplate)
	if err != nil {
		return nil, fmt.Errorf("failed to parse template: %v", err)
	}

	// List pemilik tiket (Registrant + Attendees)
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

	// Generate Tiket untuk setiap orang
	for _, owner := range owners {
		ticketInfo, exists := ticketMap[owner.TicketID]
		if !exists {
			continue // Skip jika tiket tidak ditemukan di map
		}

		// Generate QR Code dari Order Number (Besar: 250px)
		qrBase64, err := util.GenerateQRCodeBase64(order.OrderNumber, 250)
		if err != nil {
			return nil, fmt.Errorf("failed to generate QR: %v", err)
		}

		// Siapkan data untuk template
		data := TicketData{
			OrderNumber:  order.OrderNumber,
			EventName:    eventName, // Ini bisa dibuat dinamis nanti dari Event DB
			OwnerName:    owner.Name,
			TicketTitle:  ticketInfo.Title,
			TicketType:   ticketInfo.Type,
			QRCodeBase64: qrBase64,
		}

		// Eksekusi Render Template HTML
		var renderedHTML bytes.Buffer
		if err := tmpl.Execute(&renderedHTML, data); err != nil {
			return nil, fmt.Errorf("failed to render HTML: %v", err)
		}

		// Buat Nama File (Contoh: E-Voucher-RF26-09833-Michael_Jack.html)
		safeName := strings.ReplaceAll(owner.Name, " ", "_")
		fileName := fmt.Sprintf("E-Voucher-%s-%s.html", order.OrderNumber, safeName)

		attachments = append(attachments, TicketAttachment{
			FileName: fileName,
			HTMLData: renderedHTML.Bytes(),
		})
	}

	return attachments, nil
}
