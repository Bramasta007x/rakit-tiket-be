package email

import (
	"context"
	"io"

	"rakit-tiket-be/pkg/util"

	"go.uber.org/zap"
	"gopkg.in/gomail.v2"
)

type EmailService interface {
	SendTicketEmail(ctx context.Context, toEmail, orderNumber, eventName string, attachments []Attachment) error
}

type Attachment struct {
	FileName string
	Data     []byte
}

type emailService struct {
	log         util.LogUtil
	host        string
	port        int
	user        string
	password    string
	senderName  string
	senderEmail string
}

func MakeEmailService(log util.LogUtil, host string, port int, user, password, senderName, senderEmail string) EmailService {
	return emailService{
		log:         log,
		host:        host,
		port:        port,
		user:        user,
		password:    password,
		senderName:  senderName,
		senderEmail: senderEmail,
	}
}

func (s emailService) SendTicketEmail(ctx context.Context, toEmail, orderNumber, eventName string, attachments []Attachment) error {
	m := gomail.NewMessage()

	// Set Header
	m.SetHeader("From", m.FormatAddress(s.senderEmail, s.senderName))
	m.SetHeader("To", toEmail)
	m.SetHeader("Subject", "E-Ticket Anda - Pembayaran Berhasil ["+orderNumber+"]")

	// Set Body Email
	body := "Halo,\n\n" +
		"Terima kasih telah melakukan pembelian tiket untuk event **" + eventName + "**.\n\n" +
		"Pembayaran Anda dengan nomor pesanan " + orderNumber + " telah berhasil kami terima.\n\n" +
		"Terlampir E-Ticket Anda pada email ini. Silakan unduh dan tunjukkan QR Code pada tiket saat memasuki area acara.\n\n" +
		"Salam Hangat,\nTim " + s.senderName

	m.SetBody("text/plain", body)

	// Attach File (Membaca langsung dari memory RAM tanpa membuat file fisik)
	for _, att := range attachments {
		// Menghindari pointer issue di dalam loop
		fileData := att.Data

		m.Attach(att.FileName, gomail.SetCopyFunc(func(w io.Writer) error {
			_, err := w.Write(fileData)
			return err
		}))
	}

	// Proses Dial & Send
	d := gomail.NewDialer(s.host, s.port, s.user, s.password)

	s.log.Info(ctx, "Mencoba mengirim email tiket...", zap.String("to", toEmail))
	if err := d.DialAndSend(m); err != nil {
		s.log.Error(ctx, "Gagal mengirim email", zap.Error(err))
		return err
	}

	return nil
}
