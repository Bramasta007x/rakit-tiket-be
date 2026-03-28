package service

const ticketPDFTemplate = `
<!DOCTYPE html>
<html lang="id">
<head>
    <meta charset="UTF-8">
    <title>E-Voucher {{ .OrderNumber }}</title>
    <style>
        @page { margin: 0; }
        body { font-family: 'Helvetica Neue', Helvetica, Arial, sans-serif; background: #f3f3f3; color: #222; margin: 0; padding: 20px; }
        .voucher { background: #fff; border: 1px solid #ccc; border-radius: 10px; padding: 25px 35px; width: 100%; max-width: 700px; margin: 0 auto; box-shadow: 0 2px 6px rgba(0, 0, 0, 0.15); }
        .header { text-align: center; margin-bottom: 20px; }
        .header h1 { font-size: 22px; color: #b20000; margin-bottom: 5px; text-transform: uppercase; }
        .header h2 { font-size: 16px; font-weight: normal; color: #444; }
        .section-title { font-size: 14px; font-weight: bold; border-bottom: 2px solid #e0e0e0; margin-top: 25px; margin-bottom: 10px; text-transform: uppercase; color: #555; }
        table { width: 100%; border-collapse: collapse; }
        td { padding: 6px 0; font-size: 13px; vertical-align: top; }
        td:first-child { width: 45%; color: #555; }
        td:last-child { color: #111; font-weight: bold; }
        .qr { text-align: center; margin-top: 30px; }
        .qr img { width: 160px; height: 160px; border: 1px solid #ccc; border-radius: 8px; padding: 6px; background: #fff; }
        .footer { margin-top: 35px; font-size: 11px; color: #666; text-align: center; border-top: 1px solid #ddd; padding-top: 10px; line-height: 1.4; }
    </style>
</head>
<body>
    <div class="voucher">
        <div class="header">
            <h1>E-Voucher</h1>
            <h2>{{ .EventName }}</h2>
        </div>

        <div class="section-title">Informasi Pemegang Tiket / Ticket Holder</div>
        <table>
            <tr><td>Nama</td><td>: {{ .OwnerName }}</td></tr>
            <tr><td>Tipe Tiket</td><td>: {{ .TicketTitle }}</td></tr>
            <tr><td>Harga Tiket</td><td>: {{ .TicketPrice }}</td></tr>
        </table>

        <div class="section-title">Informasi Pesanan</div>
        <table>
            <tr><td>Kode Tagihan</td><td>: {{ .OrderNumber }}</td></tr>
            <tr><td>Tanggal Pembelian</td><td>: {{ .PaymentTime }}</td></tr>
            <tr><td>Status Pembayaran</td><td>: <span style="color: green;">{{ .PaymentStatus }}</span></td></tr>
            <tr><td>Jumlah Pembayaran</td><td>: {{ .Amount }}</td></tr>
            <tr><td>Pemesan Utama</td><td>: {{ .RegistrantName }}</td></tr>
        </table>

        <div class="qr">
            <img src="{{ .QRCodePath }}" alt="QR Code">
            <p style="font-size: 12px; color: #666; margin-top: 8px;">Tunjukkan QR Code ini saat check-in</p>
            <p style="font-size: 11px; color: #999; margin:0;">{{ .OrderNumber }}</p>
        </div>

        <div class="section-title">Detail Event</div>
        <table>
            <tr><td>Nama Event</td><td>: {{ .EventName }}</td></tr>
            <tr><td>Tanggal & Waktu</td><td>: {{ .EventDate }} | {{ .EventTimeStart }} - {{ .EventTimeEnd }}</td></tr>
            <tr><td>Lokasi</td><td>: {{ .EventLocation }}</td></tr>
        </table>

        <div class="section-title">Syarat & Ketentuan</div>
        <p style="font-size: 12px; line-height: 1.5; text-align: justify;">
            E-voucher ini berlaku untuk <strong>1 (satu) orang</strong> dan hanya dapat digunakan untuk acara yang tercantum di atas. Harap tunjukkan e-voucher (digital atau cetak) saat memasuki area acara. Nama pada voucher harus sesuai dengan identitas diri.
        </p>

        <div class="footer">
            {{ .EventName }} Official © {{ .CurrentYear }}<br>
            Tersistem otomatis oleh Rakit Tiket
        </div>
    </div>
</body>
</html>
`
