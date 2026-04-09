package email

import (
	"context"
	"fmt"
	"io"

	"rakit-tiket-be/pkg/util"

	"go.uber.org/zap"
	"gopkg.in/gomail.v2"
)

type EmailService interface {
	SendTicketEmail(ctx context.Context, toEmail, orderNumber, eventName, ownerName string, attachments []Attachment) error
	SendTransferApprovalEmail(ctx context.Context, toEmail, orderNumber, eventName, ownerName string, attachments []Attachment) error
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

func (s emailService) SendTicketEmail(ctx context.Context, toEmail, orderNumber, eventName, ownerName string, attachments []Attachment) error {
	m := gomail.NewMessage()

	m.SetHeader("From", m.FormatAddress(s.senderEmail, s.senderName))
	m.SetHeader("To", toEmail)
	m.SetHeader("Subject", "E-Ticket Anda - Pembayaran Berhasil ["+orderNumber+"]")

	// Email HTML
	htmlBody := fmt.Sprintf(`
	<!DOCTYPE html>
	<html>
	<body style="font-family: Arial, sans-serif; color: #333; line-height: 1.6; padding: 20px;">
		<div style="max-width: 600px; margin: 0 auto; border: 1px solid #ddd; border-radius: 10px; padding: 20px; background-color: #f9f9f9;">
			<h2 style="color: #b20000; text-align: center;">Pembayaran Berhasil! 🎉</h2>
			<p>Halo <b>%s</b>,</p>
			<p>Terima kasih telah melakukan pembelian tiket untuk event <strong>%s</strong>.</p>
			<p>Pembayaran Anda dengan nomor pesanan <b style="color:#1e40af;">%s</b> telah berhasil kami terima.</p>
			<div style="background-color: #fff; padding: 15px; border-left: 4px solid #b20000; margin: 20px 0;">
				<p style="margin: 0;">Terlampir E-Ticket (PDF) Anda pada email ini. Silakan unduh dan tunjukkan <b>QR Code</b> pada tiket saat memasuki area acara untuk di-scan oleh petugas.</p>
			</div>
			<br>
			<p>Salam Hangat,<br><b>Tim %s</b></p>
		</div>
	</body>
	</html>
	`, ownerName, eventName, orderNumber, s.senderName)

	m.SetBody("text/html", htmlBody) // Set sebagai text/html

	// Attach File PDF
	for _, att := range attachments {
		fileData := att.Data
		m.Attach(att.FileName, gomail.SetCopyFunc(func(w io.Writer) error {
			_, err := w.Write(fileData)
			return err
		}))
	}

	d := gomail.NewDialer(s.host, s.port, s.user, s.password)

	s.log.Info(ctx, "Mencoba mengirim email HTML PDF tiket...", zap.String("to", toEmail))
	if err := d.DialAndSend(m); err != nil {
		s.log.Error(ctx, "Gagal mengirim email", zap.Error(err))
		return err
	}

	return nil
}

func (s emailService) SendTransferApprovalEmail(ctx context.Context, toEmail, orderNumber, eventName, ownerName string, attachments []Attachment) error {
	m := gomail.NewMessage()

	m.SetHeader("From", m.FormatAddress(s.senderEmail, s.senderName))
	m.SetHeader("To", toEmail)
	m.SetHeader("Subject", "Pembayaran Transfer Manual Anda Telah Diverifikasi - "+orderNumber)

	htmlBody := fmt.Sprintf(`
	<!DOCTYPE html>
	<html>
	<body style="font-family: Arial, sans-serif; color: #333; line-height: 1.6; padding: 20px;">
		<div style="max-width: 600px; margin: 0 auto; border: 1px solid #ddd; border-radius: 10px; padding: 20px; background-color: #f9f9f9;">
			<h2 style="color: #059669; text-align: center;">Pembayaran Telah Diverifikasi! ✅</h2>
			<p>Halo <b>%s</b>,</p>
			<p>Pembayaran transfer manual Anda untuk event <strong>%s</strong> telah berhasil kami verifikasi.</p>
			<p>Nomor pesanan: <b style="color:#1e40af;">%s</b></p>
			<div style="background-color: #fff; padding: 15px; border-left: 4px solid #059669; margin: 20px 0;">
				<p style="margin: 0;">Terlampir E-Ticket (PDF) Anda pada email ini. Silakan unduh dan tunjukkan <b>QR Code</b> pada tiket saat memasuki area acara.</p>
			</div>
			<br>
			<p>Jika Anda memiliki pertanyaan, jangan ragu untuk menghubungi kami.</p>
			<p>Salam Hangat,<br><b>Tim %s</b></p>
		</div>
	</body>
	</html>
	`, ownerName, eventName, orderNumber, s.senderName)

	m.SetBody("text/html", htmlBody)

	for _, att := range attachments {
		fileData := att.Data
		m.Attach(att.FileName, gomail.SetCopyFunc(func(w io.Writer) error {
			_, err := w.Write(fileData)
			return err
		}))
	}

	d := gomail.NewDialer(s.host, s.port, s.user, s.password)

	s.log.Info(ctx, "Mencoba mengirim email approval transfer...", zap.String("to", toEmail))
	if err := d.DialAndSend(m); err != nil {
		s.log.Error(ctx, "Gagal mengirim email approval", zap.Error(err))
		return err
	}

	return nil
}
