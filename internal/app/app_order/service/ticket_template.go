package service

const ticketPDFTemplate = `
<!DOCTYPE html>
<html lang="id">
<head>
  <meta charset="UTF-8">
  <title>E-Ticket {{ .OrderNumber }}</title>
  <style>
    @page { margin: 0; }

    body {
      margin: 0;
      padding: 22px;
      background: #f1f5f9;
      color: #0f172a;
      font-family: "DejaVu Sans", "Helvetica Neue", Helvetica, Arial, sans-serif;
      font-size: 12px;
      line-height: 1.45;
    }

    .ticket {
      max-width: 760px;
      margin: 0 auto;
      background: #ffffff;
      border: 1px solid #dbe4f0;
      border-radius: 16px;
      overflow: hidden;
      box-shadow: 0 12px 28px rgba(15, 23, 42, 0.08);
    }

    .hero {
      background: #2563eb;
      background-image: linear-gradient(120deg, #1d4ed8 0%, #2563eb 55%, #3b82f6 100%);
      padding: 20px 24px;
      color: #ffffff;
    }

    .hero-top {
      width: 100%;
      border-collapse: collapse;
      margin-bottom: 10px;
    }

    .hero-top td {
      vertical-align: top;
      padding: 0;
    }

    .hero-badge {
      display: inline-block;
      padding: 4px 10px;
      font-size: 10px;
      letter-spacing: 0.6px;
      border-radius: 999px;
      background: rgba(255, 255, 255, 0.18);
      border: 1px solid rgba(255, 255, 255, 0.35);
      text-transform: uppercase;
      font-weight: 700;
    }

    .hero-title {
      margin: 0;
      font-size: 22px;
      font-weight: 700;
      letter-spacing: 0.2px;
    }

    .hero-event {
      margin: 6px 0 0 0;
      font-size: 13px;
      opacity: 0.95;
      font-weight: 500;
    }

    .hero-order {
      text-align: right;
      font-size: 11px;
      opacity: 0.95;
      font-weight: 600;
    }

    .hero-meta {
      margin: 0;
      font-size: 11px;
      opacity: 0.9;
    }

    .content {
      padding: 20px 22px 22px;
    }

    .split {
      width: 100%;
      border-collapse: separate;
      border-spacing: 10px;
      margin: 0 0 4px 0;
    }

    .split td {
      vertical-align: top;
      width: 50%;
      padding: 0;
    }

    .panel {
      border: 1px solid #e2e8f0;
      border-radius: 12px;
      padding: 12px 14px;
      background: #ffffff;
    }

    .panel-soft {
      background: #f8fafc;
    }

    .panel-title {
      margin: 0 0 8px 0;
      font-size: 11px;
      color: #475569;
      text-transform: uppercase;
      letter-spacing: 0.5px;
      font-weight: 700;
    }

    .kv {
      width: 100%;
      border-collapse: collapse;
    }

    .kv tr + tr td {
      border-top: 1px dashed #e2e8f0;
    }

    .kv td {
      padding: 6px 0;
      font-size: 12px;
      vertical-align: top;
    }

    .kv td:first-child {
      color: #64748b;
      width: 46%;
    }

    .kv td:last-child {
      color: #0f172a;
      font-weight: 700;
      text-align: right;
    }

    .status-paid {
      display: inline-block;
      padding: 2px 8px;
      border-radius: 999px;
      background: #dcfce7;
      color: #166534;
      border: 1px solid #bbf7d0;
      font-size: 10px;
      font-weight: 700;
      letter-spacing: 0.3px;
    }

    .qr-panel {
      margin-top: 10px;
      border: 1px solid #dbe4f0;
      border-radius: 12px;
      padding: 14px;
      text-align: center;
      background: #ffffff;
    }

    .qr-panel img {
      width: 162px;
      height: 162px;
      border: 1px solid #d1d5db;
      border-radius: 10px;
      padding: 6px;
      background: #ffffff;
    }

    .qr-title {
      margin: 10px 0 3px 0;
      font-size: 12px;
      color: #1e293b;
      font-weight: 700;
    }

    .qr-sub {
      margin: 0;
      font-size: 11px;
      color: #64748b;
    }

    .terms {
      margin-top: 10px;
      border: 1px solid #e2e8f0;
      border-radius: 12px;
      background: #f8fafc;
      padding: 12px 14px;
    }

    .terms p {
      margin: 0;
      color: #334155;
      font-size: 11px;
      line-height: 1.55;
      text-align: justify;
    }

    .footer {
      border-top: 1px solid #e2e8f0;
      margin-top: 14px;
      padding-top: 10px;
      text-align: center;
      font-size: 10px;
      color: #64748b;
      line-height: 1.5;
    }
  </style>
</head>
<body>
  <div class="ticket">
    <div class="hero">
      <table class="hero-top">
        <tr>
          <td>
            <span class="hero-badge">E-Ticket</span>
          </td>
          <td class="hero-order">
            ORDER: {{ .OrderNumber }}
          </td>
        </tr>
      </table>
      <h1 class="hero-title">{{ .EventName }}</h1>
      <p class="hero-event">Akses masuk resmi untuk 1 pemegang tiket</p>
      <p class="hero-meta">{{ .EventDate }} | {{ .EventTimeStart }} - {{ .EventTimeEnd }}</p>
    </div>

    <div class="content">
      <table class="split">
        <tr>
          <td>
            <div class="panel">
              <h3 class="panel-title">Ticket Holder</h3>
              <table class="kv">
                <tr><td>Nama</td><td>{{ .OwnerName }}</td></tr>
                <tr><td>Tipe Tiket</td><td>{{ .TicketTitle }}</td></tr>
                <tr><td>Harga Tiket</td><td>{{ .TicketPrice }}</td></tr>
              </table>
            </div>
          </td>
          <td>
            <div class="panel panel-soft">
              <h3 class="panel-title">Informasi Pesanan</h3>
              <table class="kv">
                <tr><td>Kode Tagihan</td><td>{{ .OrderNumber }}</td></tr>
                <tr><td>Tanggal Bayar</td><td>{{ .PaymentTime }}</td></tr>
                <tr><td>Status</td><td><span class="status-paid">{{ .PaymentStatus }}</span></td></tr>
                <tr><td>Total Bayar</td><td>{{ .Amount }}</td></tr>
                <tr><td>Pemesan</td><td>{{ .RegistrantName }}</td></tr>
              </table>
            </div>
          </td>
        </tr>
      </table>

      <table class="split">
        <tr>
          <td>
            <div class="panel">
              <h3 class="panel-title">Detail Event</h3>
              <table class="kv">
                <tr><td>Nama Event</td><td>{{ .EventName }}</td></tr>
                <tr><td>Tanggal & Jam</td><td>{{ .EventDate }} | {{ .EventTimeStart }} - {{ .EventTimeEnd }}</td></tr>
                <tr><td>Lokasi</td><td>{{ .EventLocation }}</td></tr>
              </table>
            </div>
          </td>
          <td>
            <div class="qr-panel">
              <img src="{{ .QRCodePath }}" alt="QR Code">
              <p class="qr-title">Scan QR saat check-in</p>
              <p class="qr-sub">{{ .OrderNumber }}</p>
            </div>
          </td>
        </tr>
      </table>

      <div class="terms">
        <h3 class="panel-title" style="margin-top:0;">Syarat & Ketentuan</h3>
        <p>
          E-ticket ini berlaku untuk <strong>1 (satu) orang</strong> sesuai nama pemegang tiket.
          Harap tunjukkan e-ticket (digital/cetak) dan identitas diri saat check-in. Tiket tidak dapat dipindahtangankan
          tanpa persetujuan penyelenggara.
        </p>
      </div>

      <div class="footer">
        {{ .EventName }} Official &copy; {{ .CurrentYear }}<br>
        Powered by Rakit Tiket
      </div>
    </div>
  </div>
</body>
</html>
`
