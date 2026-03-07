package service

const ticketHTMLTemplate = `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <title>E-Voucher {{.OrderNumber}}</title>
    <style>
        body { font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif; background-color: #f4f4f4; padding: 20px; }
        .ticket-card { max-width: 600px; margin: 0 auto; background: #fff; border-radius: 10px; overflow: hidden; box-shadow: 0 4px 8px rgba(0,0,0,0.1); }
        .header { background: #2563eb; color: #fff; padding: 20px; text-align: center; }
        .header h1 { margin: 0; font-size: 24px; }
        .body { padding: 20px; text-align: center; }
        .qr-code { margin: 20px 0; }
        .qr-code img { border: 4px solid #f3f4f6; border-radius: 10px; width: 200px; height: 200px; }
        .details { text-align: left; margin-top: 20px; padding: 15px; background: #f9fafb; border-radius: 8px; }
        .details table { width: 100%; border-collapse: collapse; }
        .details td { padding: 8px 0; border-bottom: 1px solid #e5e7eb; }
        .details td:first-child { font-weight: bold; color: #4b5563; width: 40%; }
        .footer { background: #1e40af; color: #fff; padding: 15px; text-align: center; font-size: 14px; }
    </style>
</head>
<body>
    <div class="ticket-card">
        <div class="header">
            <h1>E-TICKET / VOUCHER</h1>
            <p>{{.EventName}}</p>
        </div>
        <div class="body">
            <p>Tunjukkan QR Code ini pada saat penukaran tiket / masuk ke area event.</p>
            <div class="qr-code">
                <img src="{{.QRCodeBase64}}" alt="QR Code" />
            </div>
            <h2>{{.OrderNumber}}</h2>
            
            <div class="details">
                <table>
                    <tr><td>Nama Pemilik</td><td>{{.OwnerName}}</td></tr>
                    <tr><td>Jenis Tiket</td><td>{{.TicketTitle}} ({{.TicketType}})</td></tr>
                    <tr><td>Status Pembayaran</td><td><strong style="color: #16a34a;">LUNAS</strong></td></tr>
                </table>
            </div>
        </div>
        <div class="footer">
            Mohon simpan E-Ticket ini dengan baik. Jangan bagikan QR Code kepada siapapun.
        </div>
    </div>
</body>
</html>
`
