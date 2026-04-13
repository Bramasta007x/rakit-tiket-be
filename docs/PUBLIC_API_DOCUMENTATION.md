# RakitTiket Public API Documentation

**Version:** 2.0.0  
**Last Updated:** 2026-04-12  
**Base URL:** `http://localhost:8000/api/v1`

---

## Table of Contents

1. [Overview](#overview)
2. [Payment Flow](#payment-flow)
3. [Public Endpoints](#public-endpoints)
   - [1. Get Events](#1-get-events)
   - [2. Get Tickets](#2-get-tickets)
   - [3. Register (Booking)](#3-register-booking)
   - [4. Get Payment Options](#4-get-payment-options)
   - [5. Get Bank Accounts](#5-get-bank-accounts)
   - [6. Initiate Checkout](#6-initiate-checkout)
   - [7. Submit Transfer Proof](#7-submit-transfer-proof)
   - [8. Get Order Status](#8-get-order-status)
9. [Webhooks](#webhooks)
10. [Error Handling](#error-handling)
11. [Frontend Implementation](#frontend-implementation)

---

## Overview

Dokumentasi ini menjelaskan seluruh Public API yang digunakan oleh Frontend untuk flow pembayaran dari registrasi hingga mendapatkan e-ticket.

### Key Concepts

| Concept | Description |
|---------|-------------|
| **Event** | Event yang memiliki tiket |
| **Ticket** | Tiket yang bisa dipesan user |
| **Register** | Pemesanan tiket (booking) |
| **Checkout** | Inisiasi pembayaran |
| **Order** | Transaksi pembayaran |
| **Payment Status** | Status: `pending`, `paid`, `expired`, `failed` |

---

## Payment Flow

```
┌─────────────────────────────────────────────────────────────────────────────────┐
│                         USER PAYMENT FLOW                                          │
├─────────────────────────────────────────────────────────────────────────────────┤
│                                                                                   │
│   USER              FRONTEND              BACKEND              PAYMENT            │
│     │                   │                    │                    │                 │
│     │  1. Get Events   │                    │                    │                 │
│     │ ────────────────> │                    │                    │                 │
│     │                   │                    │                    │                 │
│     │  2. Get Tickets  │                    │                    │                 │
│     │ ────────────────> │                    │                    │                 │
│     │                   │                    │                    │                 │
│     │  3. Register     │                    │                    │                 │
│     │ ────────────────────────────────────> │                    │                 │
│     │                   │                    │                    │                 │
│     │  Order created   │                    │                    │                 │
│     │  status: pending │                    │                    │                 │
│     │ <────────────────────────────────────── │                    │                 │
│     │                   │                    │                    │                 │
│     │  4. Get Payment  │                    │                    │                 │
│     │     Options     │                    │                    │                 │
│     │ ────────────────> │                    │                    │                 │
│     │                   │                    │                    │                 │
│     │  5. Checkout     │                    │                    │                 │
│     │ ────────────────────────────────────> │                    │                 │
│     │                   │                    │                    │                 │
│     │              ┌────────────────────────────────────────┐   │                 │
│     │              │         GATEWAY PATH (MIDTRANS)          │   │                 │
│     │              └────────────────────────────────────────┘   │                 │
│     │                   │                    │  6. Create Token   │                 │
│     │                   │                    │ ────────────────> │                 │
│     │                   │                    │ <─────────────── │                 │
│     │  payment_url      │                    │                    │                 │
│     │ <────────────────────────────────────── │                    │                 │
│     │                   │                    │                    │                 │
│     │  7. Pay at       │                    │                    │                 │
│     │  Gateway         │                    │                    │                 │
│     │ ──────────────────────────────────────────────────────────────> │          │
│     │                   │                    │                    │                 │
│     │                   │  8. Webhook        │                    │                 │
│     │                   │ <────────────────── │                    │                 │
│     │                   │                    │                    │                 │
│     │              ┌────────────────────────────────────────┐   │                 │
│     │              │       MANUAL TRANSFER PATH             │   │                 │
│     │              └────────────────────────────────────────┘   │                 │
│     │                   │                    │                    │                 │
│     │  Get Bank Accts  │                    │                    │                 │
│     │ ────────────────> │                    │                    │                 │
│     │                   │                    │                    │                 │
│     │  User transfers   │                    │                    │                 │
│     │ ────────────────────────────────────> │                    │                 │
│     │  Submit Proof     │                    │                    │                 │
│     │ ────────────────────────────────────> │                    │                 │
│     │                   │                    │                    │                 │
│     │                   │  Admin verifies    │                    │                 │
│     │                   │ <────────────────── │                    │                 │
│     │                   │                    │                    │                 │
│     │  9. Get Status   │                    │                    │                 │
│     │ ────────────────> │                    │                    │                 │
│     │                   │                    │                    │                 │
│     │  status: paid    │                    │                    │                 │
│     │  E-Ticket sent   │                    │                    │                 │
│     │ <────────────────────────────────────── │                    │                 │
│     │                   │                    │                    │                 │
└─────────────────────────────────────────────────────────────────────────────────┘
```

---

## Public Endpoints

### 1. Get Events

**Purpose:** Mendapatkan daftar event yang tersedia untuk dipesan.

```
GET /api/v1/events
```

#### Query Parameters (Optional)

| Parameter | Type | Description |
|-----------|------|-------------|
| `status` | string | Filter by status (e.g., "published", "draft") |
| `date_start` | string | Filter events starting from date (YYYY-MM-DD) |
| `date_end` | string | Filter events ending before date (YYYY-MM-DD) |

#### Response

```json
{
  "data": [
    {
      "id": "evt-uuid-123",
      "name": "Music Festival 2026",
      "description": "Annual music festival...",
      "status": "published",
      "max_ticket_per_tx": 10,
      "ticket_prefix_code": "MF",
      "created_at": "2026-04-01T00:00:00Z",
      "updated_at": "2026-04-01T00:00:00Z"
    }
  ],
  "count": 1
}
```

#### Notes

- Events dengan status `published` akan ditampilkan di landing page
- Filter `status` opsional, default menampilkan semua event

---

### 2. Get Tickets

**Purpose:** Mendapatkan daftar tiket yang tersedia untuk sebuah event.

```
GET /api/v1/tickets
```

#### Query Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `event_id` | string (UUID) | Yes | ID event untuk filter tiket |

#### Additional Query Parameters (Optional)

| Parameter | Type | Description |
|-----------|------|-------------|
| `type` | string | Filter by ticket type (e.g., "VIP", "Regular") |
| `status` | string | Filter by status (e.g., "published") |

#### Request Example

```
GET /api/v1/tickets?event_id=evt-uuid-123
```

#### Response

```json
{
  "data": [
    {
      "id": "tkt-uuid-456",
      "event_id": "evt-uuid-123",
      "title": "VIP Access",
      "type": "VIP",
      "price": 500000,
      "total": 100,
      "sold": 25,
      "booked": 10,
      "available": 65,
      "status": "published",
      "description": "VIP access with exclusive perks",
      "created_at": "2026-04-01T00:00:00Z",
      "updated_at": "2026-04-01T00:00:00Z"
    },
    {
      "id": "tkt-uuid-789",
      "event_id": "evt-uuid-123",
      "title": "Regular Admission",
      "type": "Regular",
      "price": 150000,
      "total": 500,
      "sold": 100,
      "booked": 20,
      "available": 380,
      "status": "published",
      "description": "Regular admission ticket",
      "created_at": "2026-04-01T00:00:00Z",
      "updated_at": "2026-04-01T00:00:00Z"
    }
  ],
  "count": 2
}
```

#### Field Descriptions

| Field | Type | Description |
|-------|------|-------------|
| `id` | string (UUID) | Unique ticket ID |
| `title` | string | Nama tiket |
| `type` | string | Jenis tiket (VIP, Regular, dll) |
| `price` | number | Harga tiket dalam IDR |
| `total` | integer | Total kapasitas tiket |
| `sold` | integer | Jumlah tiket yang sudah lunas |
| `booked` | integer | Jumlah tiket yang sedang dipesan (pending payment) |
| `available` | integer | Jumlah tiket yang masih bisa dipesan |
| `status` | string | Status tiket |

#### Notes

- Frontend harus menggunakan `event_id` dari hasil Get Events
- Cek `available > 0` sebelum menampilkan opsi pembelian
- `booked` menunjukkan tiket yang sedang dalam proses checkout (belum lunas)

---

### 3. Register (Booking)

**Purpose:** Membuat pemesanan tiket. Fungsi ini akan:
1. Reserve stock tiket (atomic booking)
2. Membuat Order dengan status `pending`
3. Generate unique code untuk registrant

```
POST /api/v1/register
Content-Type: application/json
```

#### Request Body

```json
{
  "registrant": {
    "ticket_id": "tkt-uuid-456",
    "name": "John Doe",
    "email": "john@example.com",
    "phone": "081234567890",
    "gender": "male",
    "birthdate": "1990-01-15"
  },
  "attendees": [
    {
      "ticket_id": "tkt-uuid-456",
      "name": "Jane Doe",
      "gender": "female",
      "birthdate": "1992-05-20"
    }
  ]
}
```

#### Field Descriptions

**Registrant (Required)**

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `ticket_id` | string (UUID) | Yes | ID tiket yang dipesan |
| `name` | string | Yes | Nama lengkap pemesan |
| `email` | string | Yes | Email valid untuk receive e-ticket |
| `phone` | string | Yes | Nomor telepon aktif |
| `gender` | string | No | Jenis kelamin (male/female) |
| `birthdate` | string | No | Tanggal lahir (YYYY-MM-DD) |

**Attendees (Optional)**

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `ticket_id` | string (UUID) | Yes | ID tiket untuk attendee |
| `name` | string | Yes | Nama lengkap attendee |
| `gender` | string | No | Jenis kelamin |
| `birthdate` | string | No | Tanggal lahir (YYYY-MM-DD) |

#### Response (Success)

```json
{
  "success": true,
  "data": {
    "order": {
      "order_id": "ord-uuid-abc",
      "order_number": "MF-2026-a1b2c3d4e5f6",
      "amount": 1000000,
      "currency": "IDR",
      "payment_status": "pending",
      "expires_at": "2026-04-12T15:30:00Z"
    },
    "registrant": {
      "id": "reg-uuid-xyz",
      "unique_code": "MF-2026-a1b2c3d4e5f6"
    }
  }
}
```

#### Response (Error - Stock Habis)

```json
{
  "success": false,
  "data": null
}
```

#### Response (Error - Max Ticket Exceeded)

```json
{
  "success": false,
  "data": null
}
```

#### Notes

- **Atomic Booking:** Stock otomatis di-reserve saat register
- **15 Menit Expiry:** Order akan expire dalam 15 menit jika tidak dibayar
- **Attendees:** Jika memesan 1 tiket untuk diri sendiri + 1 tiket untuk attendee, total 2 tiket
- **Same Event:** Semua tiket harus dari event yang sama
- **Max Per Tx:** Terbatas oleh konfigurasi event (`max_ticket_per_tx`)

#### Common Error Messages

| Message | Description |
|---------|-------------|
| `stok tiket tidak mencukupi` | Tiket yang dipesan tidak tersedia |
| `maksimal X tiket per registrasi` | Melebihi batas maksimal tiket per transaksi |
| `semua tiket dalam satu transaksi harus berasal dari event yang sama` | Tiket harus dari event yang sama |
| `tiket tidak ditemukan` | Ticket ID tidak valid |

---

### 4. Get Payment Options

**Purpose:** Mendapatkan metode pembayaran yang aktif untuk ditampilkan di halaman checkout.

```
GET /api/v1/payment-options
```

#### Response

```json
{
  "success": true,
  "data": [
    {
      "type": "GATEWAY",
      "code": "MIDTRANS",
      "gateway_name": "Midtrans",
      "display_order": 1
    },
    {
      "type": "MANUAL",
      "code": "MANUAL",
      "gateway_name": "Transfer Manual",
      "display_order": 2
    }
  ]
}
```

#### Field Descriptions

| Field | Type | Description |
|-------|------|-------------|
| `type` | string | Jenis: `GATEWAY` atau `MANUAL` |
| `code` | string | Kode spesifik: `MIDTRANS`, `XENDIT`, `DOKU`, `MANUAL` |
| `gateway_name` | string | Nama untuk ditampilkan ke user |
| `display_order` | integer | Urutan tampil (1 = paling atas) |

#### Notes

- Hanya payment method yang AKTIF yang ditampilkan
- Hasil diurutkan berdasarkan `display_order` ascending
- Frontend WAJIB menggunakan endpoint ini untuk menampilkan opsi pembayaran
- Gateway codes: `MIDTRANS`, `XENDIT`, `DOKU`, `MANUAL`

---

### 5. Get Bank Accounts

**Purpose:** Mendapatkan daftar rekening bank untuk pembayaran manual transfer.

```
GET /api/v1/bank-accounts
```

#### Response

```json
{
  "success": true,
  "data": [
    {
      "id": "bank-uuid-1",
      "bank_name": "Bank Central Asia (BCA)",
      "bank_code": "BCA",
      "account_number": "1234567890",
      "account_holder": "PT Rakit Tiket Indonesia",
      "instruction_text": "Transfer tepat hingga 3 digit terakhir"
    },
    {
      "id": "bank-uuid-2",
      "bank_name": "Bank Mandiri",
      "bank_code": "MANDIRI",
      "account_number": "1300098765432",
      "account_holder": "PT Rakit Tiket Indonesia",
      "instruction_text": null
    }
  ]
}
```

#### Field Descriptions

| Field | Type | Description |
|-------|------|-------------|
| `id` | string (UUID) | Unique ID rekening |
| `bank_name` | string | Nama bank lengkap |
| `bank_code` | string | Kode bank (BCA, MANDIRI, dll) |
| `account_number` | string | Nomor rekening |
| `account_holder` | string | Nama pemilik rekening |
| `instruction_text` | string/null | Instruksi transfer khusus |

#### Notes

- Endpoint ini PUBLIC (tidak perlu autentikasi)
- Bank accounts dikonfigurasi oleh admin
- `instruction_text` mungkin null, tampilkan instruksi default jika null

---

### 6. Initiate Checkout

**Purpose:** Memilih metode pembayaran dan memulai proses pembayaran.

```
POST /api/v1/checkout/:order_id
Content-Type: application/json
```

#### URL Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `order_id` | string (UUID) | Yes | ID order dari response Register |

#### Request Body

```json
{
  "payment_type": "MIDTRANS"
}
```

#### Valid Payment Types

| Payment Type | Description |
|-------------|-------------|
| `MIDTRANS` | Bayar via Midtrans (Credit Card, E-Wallet, Bank Transfer) |
| `XENDIT` | Bayar via Xendit |
| `DOKU` | Bayar via Doku |
| `MANUAL` | Bayar via transfer bank manual |

#### Response (Gateway - MIDTRANS/XENDIT/DOKU)

```json
{
  "success": true,
  "data": {
    "order_id": "ord-uuid-abc",
    "order_number": "MF-2026-a1b2c3d4e5f6",
    "amount": 1000000,
    "payment_type": "GATEWAY",
    "payment_status": "pending",
    "expires_at": "2026-04-12T15:30:00Z",
    "payment_info": {
      "payment_url": "https://app.sandbox.midtrans.com/snap/v2/vtweb/...",
      "payment_token": "token-xxx",
      "payment_method": "midtrans"
    }
  }
}
```

#### Response (Manual Transfer)

```json
{
  "success": true,
  "data": {
    "order_id": "ord-uuid-abc",
    "order_number": "MF-2026-a1b2c3d4e5f6",
    "amount": 1000000,
    "payment_type": "MANUAL",
    "payment_status": "pending",
    "expires_at": "2026-04-12T15:30:00Z",
    "bank_accounts": [
      {
        "bank_name": "Bank Central Asia (BCA)",
        "bank_code": "BCA",
        "account_number": "1234567890",
        "account_holder": "PT Rakit Tiket Indonesia",
        "instruction_text": "Transfer tepat hingga 3 digit terakhir"
      }
    ]
  }
}
```

#### Notes

- **Gateway:** User akan diarahkan ke `payment_info.payment_url`
- **Manual:** User transfer ke rekening yang ditampilkan, lalu submit bukti transfer
- Payment type harus uppercase (MIDTRANS, MANUAL)

#### Error Responses

| HTTP Code | Message | Description |
|-----------|---------|-------------|
| 400 | `order not found` | Order ID tidak valid |
| 400 | `order has expired` | Batas waktu pembayaran sudah habis |
| 400 | `order cannot be checked out (status: paid)` | Order sudah lunas |
| 500 | `no active payment gateway configured` | Tidak ada gateway aktif |

---

### 7. Submit Transfer Proof

**Purpose:** Upload bukti transfer untuk verifikasi manual.

```
POST /api/v1/transfers/proof
Content-Type: multipart/form-data
```

#### Request (multipart/form-data)

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `data` | JSON String | Yes | Data transfer dalam format JSON |
| `transfer_proof` | File | Yes | Gambar/PDF bukti transfer |

**JSON Data Structure:**

```json
{
  "order_id": "ord-uuid-abc",
  "bank_account_id": "bank-uuid-1",
  "sender_name": "John Doe",
  "sender_account_number": "1234567890",
  "transfer_date": "2026-04-12T14:30:00Z"
}
```

#### Request Example (cURL)

```bash
curl -X POST http://localhost:8000/api/v1/transfers/proof \
  -F "data={\"order_id\":\"ord-uuid-abc\",\"bank_account_id\":\"bank-uuid-1\",\"sender_name\":\"John Doe\",\"transfer_date\":\"2026-04-12T14:30:00Z\"}" \
  -F "transfer_proof=@/path/to/proof.jpg"
```

#### Request Example (JavaScript)

```javascript
const formData = new FormData();
formData.append('data', JSON.stringify({
  order_id: 'ord-uuid-abc',
  bank_account_id: 'bank-uuid-1',
  sender_name: 'John Doe',
  sender_account_number: '1234567890',
  transfer_date: new Date().toISOString()
}));
formData.append('transfer_proof', fileInput.files[0]);

const response = await fetch('/api/v1/transfers/proof', {
  method: 'POST',
  body: formData
});
```

#### Response (Success)

```json
{
  "success": true,
  "message": "Transfer proof submitted successfully",
  "data": {
    "id": "transfer-uuid-123",
    "order_id": "ord-uuid-abc",
    "status": "pending",
    "transfer_amount": 1000000,
    "transfer_date": "2026-04-12T14:30:00Z",
    "sender_name": "John Doe",
    "sender_account_number": "1234567890",
    "bank_account": {
      "bank_name": "Bank Central Asia (BCA)",
      "account_number": "1234567890"
    },
    "created_at": "2026-04-12T14:35:00Z"
  }
}
```

#### Error Responses

| HTTP Code | Message | Description |
|-----------|---------|-------------|
| 400 | `order_id is required` | Missing order_id |
| 400 | `bank_account_id is required` | Missing bank_account_id |
| 400 | `sender_name is required` | Missing sender_name |
| 400 | `transfer_proof file is required` | Missing file upload |
| 409 | `Transfer proof already submitted for this order` | Bukti sudah diupload |
| 404 | `Order not found` | Order tidak valid |

#### Notes

- Satu order hanya bisa submit SATU bukti transfer
- File didukung: JPG, PNG, PDF
- Max file size: sesuai konfigurasi server
- Admin akan memverifikasi dalam 1x24 jam

---

### 8. Get Order Status

**Purpose:** Mengecek status order dan pembayaran. Juga untuk mendapatkan e-ticket setelah payment berhasil.

```
GET /api/v1/orders/:order_number/status
```

#### URL Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `order_number` | string | Yes | Nomor order (dari response Register) |

#### Response (Pending)

```json
{
  "success": true,
  "data": {
    "order_number": "MF-2026-a1b2c3d4e5f6",
    "payment_method": "",
    "payment_channel": "",
    "payment_status": "pending",
    "amount": 1000000,
    "payment_time": null,
    "registrant": {
      "name": "John Doe",
      "email": "john@example.com",
      "phone": "081234567890",
      "gender": "male",
      "birthdate": "1990-01-15",
      "ticket_title": "VIP Access",
      "ticket_type": "VIP"
    },
    "attendees": [
      {
        "name": "Jane Doe",
        "gender": "female",
        "birthdate": "1992-05-20",
        "ticket_title": "VIP Access",
        "ticket_type": "VIP"
      }
    ]
  }
}
```

#### Response (Paid - With E-Ticket)

```json
{
  "success": true,
  "data": {
    "order_number": "MF-2026-a1b2c3d4e5f6",
    "payment_method": "bank_transfer",
    "payment_channel": "bca",
    "payment_status": "paid",
    "amount": 1000000,
    "payment_time": "2026-04-12T14:35:00Z",
    "registrant": {
      "name": "John Doe",
      "email": "john@example.com",
      "phone": "081234567890",
      "gender": "male",
      "birthdate": "1990-01-15",
      "ticket_title": "VIP Access",
      "ticket_type": "VIP"
    },
    "attendees": [
      {
        "name": "Jane Doe",
        "gender": "female",
        "birthdate": "1992-05-20",
        "ticket_title": "VIP Access",
        "ticket_type": "VIP"
      }
    ]
  }
}
```

#### Response (Expired)

```json
{
  "success": true,
  "data": {
    "order_number": "MF-2026-a1b2c3d4e5f6",
    "payment_method": "",
    "payment_channel": "",
    "payment_status": "expired",
    "amount": 1000000,
    "payment_time": null,
    "registrant": {
      "name": "John Doe",
      "email": "john@example.com",
      "phone": "081234567890",
      "gender": "male",
      "birthdate": "1990-01-15",
      "ticket_title": "VIP Access",
      "ticket_type": "VIP"
    },
    "attendees": []
  }
}
```

#### Payment Status Definitions

| Status | Description | Next Action |
|--------|-------------|-------------|
| `pending` | Menunggu pembayaran | Lakukan pembayaran |
| `paid` | Pembayaran berhasil | E-ticket dikirim via email |
| `expired` | Batas waktu habis | Harus pesan ulang |
| `failed` | Pembayaran gagal/ditolak | Hubungi support |

#### Notes

- Polling endpoint ini untuk cek status setelah checkout
- Interval polling: setiap 3-5 detik saat menunggu
- E-ticket akan dikirim via email setelah status `paid`
- Jika expired, user harus pesan ulang karena stock sudah di-release

---

## Webhooks

### Midtrans Webhook

**Purpose:** Backend menerima notification dari Midtrans untuk update payment status.

```
POST /api/v1/webhook/payment/midtrans
Content-Type: application/json
```

#### Supported Events

| Transaction Status | Action |
|-------------------|--------|
| `capture` | Payment captured (settlement) |
| `settlement` | Payment settled |
| `pending` | Payment pending |
| `deny` | Payment denied |
| `cancel` | Payment cancelled |
| `expire` | Payment expired |

#### Notes

- Webhook dipanggil secara otomatis oleh Midtrans
- Tidak perlu frontend intervention
- Backend akan update status dan kirim e-ticket

---

## Error Handling

### HTTP Status Codes

| Code | Meaning | Common Causes |
|------|---------|---------------|
| `200` | Success | Request berhasil |
| `201` | Created | Resource berhasil dibuat |
| `400` | Bad Request | Invalid request body / parameters |
| `404` | Not Found | Resource tidak ditemukan |
| `409` | Conflict | Data sudah ada (duplicate) |
| `500` | Server Error | Internal server error |

### Error Response Format

```json
{
  "success": false,
  "message": "Error description here"
}
```

### Common Error Messages

| Error | HTTP Code | Description |
|-------|-----------|-------------|
| `order not found` | 404 | Order tidak ditemukan |
| `order has expired` | 400 | Batas waktu pembayaran habis |
| `order cannot be checked out (status: paid)` | 400 | Order sudah lunas |
| `Transfer proof already submitted` | 409 | Bukti transfer sudah ada |
| `stok tiket tidak mencukupi` | 400 | Tiket habis |
| `maksimal X tiket per registrasi` | 400 | Exceeds max ticket limit |

---

## Frontend Implementation

### 1. Complete Registration Flow

```javascript
async function registerForEvent(registrantData, attendees = []) {
  // Step 1: Register
  const registerResponse = await fetch('/api/v1/register', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      registrant: registrantData,
      attendees: attendees
    })
  });
  
  const registerResult = await registerResponse.json();
  
  if (!registerResult.success) {
    throw new Error('Registration failed');
  }
  
  const { order, registrant } = registerResult.data;
  
  return { order, registrant };
}

async function checkoutAndPay(orderId, paymentType) {
  // Step 2: Get Payment Options
  const optionsResponse = await fetch('/api/v1/payment-options');
  const options = await optionsResponse.json();
  
  // Step 3: Initiate Checkout
  const checkoutResponse = await fetch(`/api/v1/checkout/${orderId}`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ payment_type: paymentType })
  });
  
  const checkoutData = await checkoutResponse.json();
  
  if (paymentType === 'MANUAL') {
    // For manual transfer, get bank accounts and show to user
    const bankResponse = await fetch('/api/v1/bank-accounts');
    const bankAccounts = await bankResponse.json();
    
    return {
      type: 'manual',
      data: checkoutData.data,
      bankAccounts: bankAccounts.data
    };
  } else {
    // For gateway, redirect to payment URL
    return {
      type: 'gateway',
      paymentUrl: checkoutData.data.payment_info.payment_url
    };
  }
}

async function submitTransferProof(orderId, bankAccountId, file) {
  const formData = new FormData();
  formData.append('data', JSON.stringify({
    order_id: orderId,
    bank_account_id: bankAccountId,
    sender_name: 'John Doe',
    transfer_date: new Date().toISOString()
  }));
  formData.append('transfer_proof', file);
  
  const response = await fetch('/api/v1/transfers/proof', {
    method: 'POST',
    body: formData
  });
  
  return response.json();
}

async function pollOrderStatus(orderNumber) {
  while (true) {
    const response = await fetch(`/api/v1/orders/${orderNumber}/status`);
    const result = await response.json();
    
    if (result.data.payment_status === 'paid') {
      return { status: 'paid', data: result.data };
    }
    
    if (result.data.payment_status === 'expired' || 
        result.data.payment_status === 'failed') {
      return { status: result.data.payment_status, data: result.data };
    }
    
    // Wait 5 seconds before next poll
    await new Promise(resolve => setTimeout(resolve, 5000));
  }
}
```

### 2. UI State Management

```javascript
// State machine for payment flow
const PaymentState = {
  INITIAL: 'initial',
  REGISTERED: 'registered',
  CHECKOUT_INITIATED: 'checkout_initiated',
  PAYMENT_PENDING: 'payment_pending',
  PAYMENT_COMPLETED: 'payment_completed',
  PAYMENT_FAILED: 'payment_failed',
  ORDER_EXPIRED: 'order_expired'
};

function handlePaymentFlow(state) {
  switch (state) {
    case PaymentState.INITIAL:
      showEventAndTicketSelection();
      break;
      
    case PaymentState.REGISTERED:
      showPaymentOptions();
      showCountdownTimer(order.expires_at);
      break;
      
    case PaymentState.CHECKOUT_INITIATED:
      if (paymentType === 'GATEWAY') {
        redirectToPaymentUrl();
      } else {
        showBankAccounts();
      }
      break;
      
    case PaymentState.PAYMENT_PENDING:
      if (paymentType === 'MANUAL') {
        showTransferInstructions();
        showUploadProofForm();
      }
      startPolling();
      break;
      
    case PaymentState.PAYMENT_COMPLETED:
      showSuccessMessage();
      showETicketInfo();
      break;
      
    case PaymentState.ORDER_EXPIRED:
      showExpiredMessage();
      showRebookButton();
      break;
      
    case PaymentState.PAYMENT_FAILED:
      showFailedMessage();
      showRetryButton();
      break;
  }
}
```

### 3. Countdown Timer

```javascript
function startCountdownTimer(expiresAt) {
  const expiresDate = new Date(expiresAt);
  
  const timer = setInterval(() => {
    const now = new Date();
    const diff = expiresDate - now;
    
    if (diff <= 0) {
      clearInterval(timer);
      handleOrderExpired();
      return;
    }
    
    const minutes = Math.floor(diff / 60000);
    const seconds = Math.floor((diff % 60000) / 1000);
    
    updateTimerDisplay(`${minutes}:${seconds.toString().padStart(2, '0')}`);
  }, 1000);
  
  return timer;
}
```

---

## Complete API Reference

| Method | Endpoint | Description | Auth |
|--------|----------|-------------|------|
| `GET` | `/api/v1/events` | List all events | No |
| `GET` | `/api/v1/tickets?event_id=X` | Get tickets for event | No |
| `POST` | `/api/v1/register` | Book tickets | No |
| `GET` | `/api/v1/payment-options` | Get active payment options | No |
| `GET` | `/api/v1/bank-accounts` | Get bank accounts for manual transfer | No |
| `POST` | `/api/v1/checkout/:order_id` | Initiate checkout | No |
| `POST` | `/api/v1/transfers/proof` | Submit transfer proof | No |
| `GET` | `/api/v1/orders/:order_number/status` | Get order status | No |
| `POST` | `/api/v1/webhook/payment/midtrans` | Midtrans webhook | No |

---

## Support

Jika ada pertanyaan tentang API ini, hubungi:

- **Backend Team**: backend@rakittiket.com
- **Technical Lead**: [Nama]

---

**Document Version:** 2.0.0  
**Last Updated:** 2026-04-12
