# RakitTiket Payment Flow API Documentation

**Version:** 2.1.0  
**Last Updated:** 2026-04-11  
**Base URL:** `http://localhost:8000/api/v1`

---

## Table of Contents

1. [Overview](#overview)
2. [Payment Flow Diagram](#payment-flow-diagram)
3. [Endpoints](#endpoints)
4. [Payment Status Flow](#payment-status-flow)
5. [Error Handling](#error-handling)
6. [Examples](#examples)

---

## Overview

Sistem pembayaran RakitTiket mendukung dua metode pembayaran:

| Method | Code | Description |
|--------|------|-------------|
| **Midtrans (Gateway)** | `MIDTRANS` / `GATEWAY` | Pembayaran via Midtrans (Credit Card, E-Wallet, Bank Transfer via Midtrans) |
| **Manual Transfer** | `MANUAL` | Pembayaran via transfer bank langsung, verifikasi oleh admin |

### Key Concepts

- **Booking (Register)**: User memilih tiket → Order dibuat dengan status `pending`
- **Checkout**: User memilih metode pembayaran → Initiate pembayaran
- **Payment Completion**: 
  - Midtrans: via webhook otomatis setelah payment
  - Manual: admin manual verify setelah user upload bukti transfer

---

## Payment Flow Diagram

```
┌─────────────────────────────────────────────────────────────────────────────────┐
│                              PAYMENT FLOW                                          │
└─────────────────────────────────────────────────────────────────────────────────┘

    USER                    BACKEND                    MIDTRANS/ADMIN               │
      │                         │                            │                      │
      │  POST /register         │                            │                      │
      │  (pilih tiket)         │                            │                      │
      │ ─────────────────────> │                            │                      │
      │                         │                            │                      │
      │  Order created          │                            │                      │
      │  status: pending        │                            │                      │
      │ <───────────────────── │                            │                      │
      │                         │                            │                      │
      │  ┌─────────────────┐   │                            │                      │
      │  │ 15 MENIT EXPIRED │   │                            │                      │
      │  └─────────────────┘   │                            │                      │
      │                         │                            │                      │
      │                         │                            │                      │
      │  ═══════════════════════════════════════════════════════════════════════   │
      │                         │                            │                      │
      │                    MIDTRANS PATH                     │                      │
      │  ═══════════════════════════════════════════════════════════════════════   │
      │                         │                            │                      │
      │  POST /checkout/:id     │                            │                      │
      │  { payment_type:        │                            │                      │
      │    "MIDTRANS" }        │                            │                      │
      │ ─────────────────────> │                            │                      │
      │                         │                            │                      │
      │                         │  Create Snap Token         │                      │
      │                         │ ────────────────────────> │                      │
      │                         │                            │                      │
      │                         │  { token, redirect_url }  │                      │
      │                         │ <──────────────────────── │                      │
      │  { payment_url }        │                            │                      │
      │ <───────────────────── │                            │                      │
      │                         │                            │                      │
      │  Redirect ke            │                            │                      │
      │  payment_url           │                            │                      │
      │ ─────────────────────> │                            │                      │
      │                         │                            │                      │
      │                         │              ┌─────────────────────────┐         │
      │                         │              │ User bayar di Midtrans │         │
      │                         │              └─────────────────────────┘         │
      │                         │                            │                      │
      │                         │  POST /webhook/midtrans   │                      │
      │                         │ <─────────────────────────                      │
      │                         │                            │                      │
      │                         │  Update status: paid      │                      │
      │                         │  Generate e-ticket        │                      │
      │                         │  Send email               │                      │
      │                         │                            │                      │
      │  ═══════════════════════════════════════════════════════════════════════   │
      │                         │                            │                      │
      │                    MANUAL TRANSFER PATH                │                      │
      │  ═══════════════════════════════════════════════════════════════════════   │
      │                         │                            │                      │
      │  POST /checkout/:id     │                            │                      │
      │  { payment_type:        │                            │                      │
      │    "MANUAL" }          │                            │                      │
      │ ─────────────────────> │                            │                      │
      │                         │                            │                      │
      │  { bank_accounts[] }   │                            │                      │
      │ <───────────────────── │                            │                      │
      │                         │                            │                      │
      │  User transfer ke       │                            │                      │
      │  salah satu rekening    │                            │                      │
      │                         │                            │                      │
      │  POST /transfers/proof  │                            │                      │
      │  (multipart/form-data)  │                            │                      │
      │ ─────────────────────> │                            │                      │
      │                         │                            │                      │
      │  { transfer_id }        │                            │                      │
      │ <───────────────────── │                            │                      │
      │                         │                            │                      │
      │                         │                            │  ┌────────────────┐   │
      │                         │                            │  │ Admin login    │   │
      │                         │                            │  │ cek bukti      │   │
      │                         │                            │  │ transfer       │   │
      │                         │                            │  └───────┬────────┘   │
      │                         │                            │          │           │
      │                         │  GET /admin/transfers/pending          │           │
      │                         │ <──────────────────────────────────────            │
      │                         │                            │          │           │
      │                         │  [list transfers]        │          │           │
      │                         │ ───────────────────────────────────────>           │
      │                         │                            │          │           │
      │                         │                            │ POST /admin/       │
      │                         │                            │ transfers/:id/      │
      │                         │                            │ approve             │
      │                         │                            │ <─────────          │
      │                         │                            │          │           │
      │                         │  Update status: paid      │          │           │
      │                         │  Generate e-ticket        │          │           │
      │                         │  Send email               │          │           │
      │                         │                            │          │           │
      │  ✓ Payment verified!    │                            │          │           │
      │  (check /orders/:id/status)                            │          │           │
      │ ───────────────────────────────────────────────────────────────────────> │
      │                         │                            │          │           │
      │                         │                            │          │           │
      │  ┌─────────────────┐   │                            │          │           │
      │  │ EXPIRED 15 MENIT│   │                            │          │           │
      │  │ tanpa payment   │   │                            │          │           │
      │  └─────────────────┘   │                            │          │           │
      │                         │                            │          │           │
      │  Order expired         │                            │          │           │
      │  Tickets released      │                            │          │           │
      │  Email notification    │                            │          │           │
      │ ────────────────────────────────────────────────────────────────────────> │
      │                         │                            │          │           |
```

---

## Endpoints

### 1. Register (Booking)

**Purpose:** Membuat pesanan dan memesan tiket (stock reserved). Dengan Auto-Initiate, jika hanya ada 1 metode pembayaran aktif, checkout akan langsung diproses.

```
POST /register
```

#### Request

```json
{
  "registrant": {
    "ticket_id": "94a69693-af68-45f7-a80e-6b8f763fcdd8",
    "name": "John Doe",
    "email": "john@example.com",
    "phone": "081234567890",
    "gender": "male",
    "birthdate": "1990-01-15"
  },
  "attendees": [
    {
      "ticket_id": "94a69693-af68-45f7-a80e-6b8f763fcdd8",
      "name": "Jane Doe",
      "gender": "female"
    }
  ]
}
```

#### Response (Auto-Initiate: 1 Gateway Aktif)

```json
{
  "success": true,
  "message": "Registration successful",
  "data": {
    "order": {
      "order_id": "0676566f-45b1-11e7-8f91-de3d5d3d1f9f",
      "order_number": "TKT-2026-a1b2c3d4e5f6",
      "amount": 100000,
      "currency": "IDR",
      "payment_status": "pending",
      "expires_at": "2026-04-11T20:49:00Z"
    },
    "registrant": {
      "id": "8105a021-653d-1f72-a025-661fcec1856b",
      "unique_code": "TKT-2026-a1b2c3d4e5f6"
    },
    "payment_options": [
      {
        "type": "GATEWAY",
        "code": "MIDTRANS",
        "gateway_name": "Midtrans",
        "display_order": 1
      }
    ],
    "payment_info": {
      "payment_type": "GATEWAY",
      "payment_url": "https://app.sandbox.midtrans.com/snap/v2/vtweb/...",
      "payment_token": "token-xxx"
    }
  }
}
```

#### Response (Multiple Payment Options)

Jika lebih dari 1 metode pembayaran aktif, `payment_info` tidak akan di-return. Frontend harus menampilkan pilihan ke user.

```json
{
  "success": true,
  "message": "Registration successful",
  "data": {
    "order": {
      "order_id": "0676566f-45b1-11e7-8f91-de3d5d3d1f9f",
      "order_number": "TKT-2026-a1b2c3d4e5f6",
      "amount": 100000,
      "currency": "IDR",
      "payment_status": "pending",
      "expires_at": "2026-04-11T20:49:00Z"
    },
    "registrant": {
      "id": "8105a021-653d-1f72-a025-661fcec1856b",
      "unique_code": "TKT-2026-a1b2c3d4e5f6"
    },
    "payment_options": [
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
}
```

#### Notes

- **Auto-Initiate:** Jika hanya 1 gateway aktif, checkout langsung diproses dan `payment_info` di-return
- **Multiple Options:** Jika >1 metode aktif, frontend harus tampilkan pilihan dan panggil `/checkout`
- `expires_at` menunjukkan batas waktu untuk menyelesaikan pembayaran (15 menit)
- Status awal order adalah `pending`
- Jika `expires_at` tercapai tanpa pembayaran, order akan expire otomatis

---

### 2. Checkout - Initiate Payment

**Purpose:** Memilih metode pembayaran dan initiate transaksi

```
POST /checkout/:order_id
```

#### Request

```json
{
  "payment_type": "MIDTRANS"  // atau "MANUAL"
}
```

#### Response (MIDTRANS)

```json
{
  "success": true,
  "data": {
    "order_id": "0676566f-45b1-11e7-8f91-de3d5d3d1f9f",
    "order_number": "TKT-2026-a1b2c3d4e5f6",
    "amount": 100000,
    "payment_type": "GATEWAY",
    "payment_status": "pending",
    "expires_at": "2026-04-11T20:49:00Z",
    "payment_info": {
      "payment_url": "https://app.sandbox.midtrans.com/snap/v2/vtweb/a1b2c3d4...",
      "payment_token": "a1b2c3d4-..."
    }
  }
}
```

#### Response (MANUAL)

```json
{
  "success": true,
  "data": {
    "order_id": "0676566f-45b1-11e7-8f91-de3d5d3d1f9f",
    "order_number": "TKT-2026-a1b2c3d4e5f6",
    "amount": 100000,
    "payment_type": "MANUAL",
    "payment_status": "pending",
    "expires_at": "2026-04-11T20:49:00Z",
    "bank_accounts": [
      {
        "bank_name": "Bank Central Asia (BCA)",
        "bank_code": "BCA",
        "account_number": "1234567890",
        "account_holder": "PT Rakit Tiket Indonesia",
        "instruction_text": "Transfer tepat hingga 3 digit terakhir untuk加快了验证"
      }
    ]
  }
}
```

#### Notes

- **MIDTRANS**: User akan diarahkan ke `payment_info.payment_url` untuk melakukan pembayaran
- **MANUAL**: User transfer ke salah satu rekening yang disediakan, kemudian upload bukti transfer

---

### 3. Get Bank Accounts

**Purpose:** Mendapatkan daftar rekening untuk pembayaran manual

```
GET /bank-accounts
```

#### Response

```json
{
  "success": true,
  "data": [
    {
      "bank_name": "Bank Central Asia (BCA)",
      "bank_code": "BCA",
      "account_number": "1234567890",
      "account_holder": "PT Rakit Tiket Indonesia",
      "instruction_text": "Transfer tepat hingga 3 digit terakhir"
    },
    {
      "bank_name": "Bank Mandiri",
      "bank_code": "MANDIRI",
      "account_number": "1300098765432",
      "account_holder": "PT Rakit Tiket Indonesia",
      "instruction_text": null
    }
  ]
}
```

---

### 4. Submit Transfer Proof (Manual)

**Purpose:** Upload bukti transfer untuk verifikasi manual

```
POST /transfers/proof
Content-Type: multipart/form-data
```

#### Request (multipart/form-data)

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `data` | JSON String | Yes | Data transfer dalam format JSON |
| `transfer_proof` | File | Yes | Gambar bukti transfer (jpg, png, pdf) |

**JSON Data Structure:**

```json
{
  "order_id": "0676566f-45b1-11e7-8f91-de3d5d3d1f9f",
  "bank_account_id": "abc123-def456-...",
  "sender_name": "John Doe",
  "sender_account_number": "1234567890",
  "transfer_date": "2026-04-11T14:30:00Z"
}
```

#### Response

```json
{
  "success": true,
  "message": "Transfer proof submitted successfully",
  "data": {
    "id": "transfer-uuid-123",
    "order_id": "0676566f-45b1-11e7-8f91-de3d5d3d1f9f",
    "status": "pending",
    "transfer_amount": 100000,
    "transfer_date": "2026-04-11T14:30:00Z"
  }
}
```

#### Notes

- Satu order hanya bisa submit satu bukti transfer
- Jika bukti sudah pernah diupload, akan return error `409 Conflict`
- Admin akan memverifikasi dalam 1x24 jam (atau kurang)

---

### 5. Get Order Status

**Purpose:** Mengecek status order dan pembayaran

```
GET /orders/:order_number/status
```

#### Response

```json
{
  "success": true,
  "data": {
    "order_number": "TKT-2026-a1b2c3d4e5f6",
    "payment_method": "bank_transfer",
    "payment_channel": "bca",
    "payment_status": "paid",
    "amount": 100000,
    "payment_time": "2026-04-11T14:35:00Z",
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
        "birthdate": null,
        "ticket_title": "VIP Access",
        "ticket_type": "VIP"
      }
    ]
  }
}
```

---

## Payment Status Flow

```
┌─────────────────────────────────────────────────────────────────────────────────┐
│                          PAYMENT STATUS FLOW                                      │
└─────────────────────────────────────────────────────────────────────────────────┘

    ┌─────────────────┐
    │    pending      │  (Initial status setelah register)
    └────────┬────────┘
             │
    ┌────────┴────────────────────────────────────────┐
    │                                                 │
    │                                                 │
    │  ┌──────────────────────────────────────────┐  │
    │  │          MIDTRANS PATH                    │  │
    │  │  POST /checkout (MIDTRANS)                │  │
    │  │  → User diarahkan ke payment_url         │  │
    │  │  → Midtrans webhook → status: paid       │  │
    │  └──────────────────────────────────────────┘  │
    │                                                 │
    │  ┌──────────────────────────────────────────┐  │
    │  │         MANUAL TRANSFER PATH             │  │
    │  │  POST /checkout (MANUAL)                 │  │
    │  │  → User transfer + upload bukti         │  │
    │  │  → Admin approve → status: paid         │  │
    │  └──────────────────────────────────────────┘  │
    │                                                 │
    └─────────────────────────────────────────────────┘
                           │
                           │ (Tidak ada pembayaran dalam 15 menit)
                           ▼
                   ┌───────────┐
                   │  expired  │
                   └───────────┘
```

### Status Definitions

| Status | Description | Trigger |
|--------|-------------|---------|
| `pending` | Menunggu pembayaran | Register / Checkout initiated |
| `paid` | Pembayaran berhasil | Midtrans webhook / Admin approve |
| `failed` | Pembayaran gagal | Midtrans notification / Admin reject |
| `expired` | Batas waktu habis (15 menit) | Cron job |

---

## Error Handling

### HTTP Status Codes

| Code | Meaning | Common Causes |
|------|---------|---------------|
| `200` | Success | Request berhasil |
| `201` | Created | Resource berhasil dibuat |
| `400` | Bad Request | Invalid request body / parameters |
| `404` | Not Found | Order / resource tidak ditemukan |
| `409` | Conflict | Bukti transfer sudah diupload |
| `500` | Server Error | Internal server error |

### Error Response Format

```json
{
  "success": false,
  "message": "Order not found"
}
```

### Common Errors

| Error Message | HTTP Code | Description |
|---------------|-----------|-------------|
| `order not found` | 404 | Order dengan ID tersebut tidak ditemukan |
| `order has expired` | 400 | Batas waktu pembayaran sudah habis |
| `order cannot be checked out (status: paid)` | 400 | Order sudah dibayar |
| `Transfer proof already submitted for this order` | 409 | Sudah ada bukti transfer untuk order ini |

---

## Examples

### Complete Flow: Midtrans Payment

```bash
# 1. Register / Booking
curl -X POST http://localhost:8000/api/v1/register \
  -H "Content-Type: application/json" \
  -d '{
    "registrant": {
      "ticket_id": "94a69693-af68-45f7-a80e-6b8f763fcdd8",
      "name": "John Doe",
      "email": "john@example.com",
      "phone": "081234567890"
    }
  }'

# Response: { "data": { "order": { "order_id": "abc...", "expires_at": "..." } } }

# 2. Checkout dengan Midtrans
curl -X POST http://localhost:8000/api/v1/checkout/abc123... \
  -H "Content-Type: application/json" \
  -d '{ "payment_type": "MIDTRANS" }'

# Response: { "data": { "payment_info": { "payment_url": "https://..." } } }

# 3. Redirect user ke payment_url, user bayar di Midtrans

# 4. Check status
curl http://localhost:8000/api/v1/orders/TKT-2026-xxx/status
```

### Complete Flow: Manual Transfer

```bash
# 1. Register / Booking
curl -X POST http://localhost:8000/api/v1/register \
  -H "Content-Type: application/json" \
  -d '{ "registrant": { ... } }'

# 2. Get Bank Accounts
curl http://localhost:8000/api/v1/bank-accounts

# Response: { "data": [{ "account_number": "1234567890", ... }] }

# 3. Checkout dengan Manual Transfer
curl -X POST http://localhost:8000/api/v1/checkout/abc123... \
  -H "Content-Type: application/json" \
  -d '{ "payment_type": "MANUAL" }'

# Response: { "data": { "bank_accounts": [...] } }

# 4. User transfer ke rekening, kemudian upload bukti
curl -X POST http://localhost:8000/api/v1/transfers/proof \
  -F "data={\"order_id\":\"abc123...\",\"bank_account_id\":\"def456...\",\"sender_name\":\"John Doe\",\"transfer_date\":\"2026-04-11T14:30:00Z\"}" \
  -F "transfer_proof=@/path/to/proof.jpg"

# Response: { "success": true, "data": { "status": "pending" } }

# 5. Admin approve (via dashboard admin)

# 6. Check status
curl http://localhost:8000/api/v1/orders/TKT-2026-xxx/status
# Response: { "data": { "payment_status": "paid" } }
```

---

## Important Notes for Frontend

### 1. Expiration Handling

```javascript
// Selalu tampilkan countdown timer ke user
const expiresAt = new Date(order.expires_at);
const now = new Date();
const remainingMinutes = Math.floor((expiresAt - now) / 1000 / 60);

if (remainingMinutes <= 0) {
  // Order sudah expire, user harus pesan ulang
  showMessage("Order telah kadaluwarsa. Silakan pesan ulang.");
}
```

### 2. Midtrans Redirect

```javascript
// Setelah dapat payment_url, redirect user
window.location.href = response.data.payment_info.payment_url;

// Setelah user kembali dari Midtrans, check status
window.addEventListener('message', (event) => {
  if (event.data === 'success') {
    checkOrderStatus();
  }
});
```

### 3. Polling Status (for Manual Transfer)

```javascript
// Check status secara periodik sampai paid
async function waitForPayment() {
  while (true) {
    const response = await fetch(`/orders/${orderNumber}/status`);
    const data = await response.json();
    
    if (data.data.payment_status === 'paid') {
      showSuccessMessage("Pembayaran berhasil diverifikasi!");
      break;
    }
    
    if (data.data.payment_status === 'failed' || 
        data.data.payment_status === 'expired') {
      showErrorMessage("Pembayaran gagal atau kadaluwarsa");
      break;
    }
    
    await new Promise(r => setTimeout(r, 5000)); // Check setiap 5 detik
  }
}
```

---

## Simple State Machine

```
┌────────────────────────────────────────────────────────────────────────────────┐
│                         SIMPLIFIED PAYMENT STATES                                │
├────────────────────────────────────────────────────────────────────────────────┤
│                                                                                 │
│    Register                    Checkout                        Payment           │
│      │                            │                              │             │
│      ▼                            ▼                              ▼             │
│   pending ──────────────────► pending ◄──────────────────► paid             │
│                                  │                              │             │
│                                  │ (15 menit tanpa payment)      │             │
│                                  │                              │             │
│                                  ▼                              │             │
│                               expired ──────────────────────────┘             │
│                                  │                              │             │
│                                  │ (Payment gagal)              │             │
│                                  ▼                              │             │
│                                failed ──────────────────────────┘             │
│                                                                                 │
│  State Transitions:                                                             │
│  • pending → paid (sukses)                                                      │
│  • pending → expired (timeout 15 menit)                                         │
│  • pending → failed (payment gagal/ditolak)                                     │
│                                                                                 │
└────────────────────────────────────────────────────────────────────────────────┘
```

### 4. File Upload

```javascript
// Upload bukti transfer
const formData = new FormData();
formData.append('data', JSON.stringify({
  order_id: orderId,
  bank_account_id: selectedBank.id,
  sender_name: 'John Doe',
  transfer_date: new Date().toISOString()
}));
formData.append('transfer_proof', fileInput.files[0]);

const response = await fetch('/transfers/proof', {
  method: 'POST',
  body: formData
});
```

---

## Payment Options (Public)

### Get Active Payment Options

**Purpose:** Mendapatkan daftar metode pembayaran yang aktif untuk ditampilkan di halaman checkout. Endpoint ini PUBLIC (tidak memerlukan autentikasi).

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

#### Notes

- Hanya payment method yang AKTIF yang ditampilkan
- Hasil diurutkan berdasarkan `display_order` ascending
- Frontend WAJIB menggunakan endpoint ini untuk menampilkan opsi pembayaran
- Gateway codes: `MIDTRANS`, `XENDIT`, `DOKU`, `MANUAL`

---

## Admin Endpoints (for Reference)

> **Catatan:** Untuk dokumentasi lengkap Admin API (Gateway Management & Manual Transfer Configuration), lihat [PAYMENT_GATEWAY_ADMIN_API.md](./PAYMENT_GATEWAY_ADMIN_API.md)

### Quick Reference - Admin Payment Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/api/v1/admin/gateways` | List all payment gateways |
| `POST` | `/api/v1/admin/gateways/:code/activate` | Activate a gateway |
| `POST` | `/api/v1/admin/gateways/:code/deactivate` | Deactivate a gateway |
| `PUT` | `/api/v1/admin/gateways/:code/display-order` | Set gateway display order |
| `POST` | `/api/v1/admin/manual-transfer/enable` | Enable manual transfer |
| `POST` | `/api/v1/admin/manual-transfer/disable` | Disable manual transfer |
| `PUT` | `/api/v1/admin/manual-transfer/display-order` | Set manual transfer display order |

### Get Pending Transfers

```
GET /api/v1/admin/transfers/pending
Authorization: Bearer <admin_token>
```

### Approve Transfer

```
POST /api/v1/admin/transfers/:transfer_id/approve
Authorization: Bearer <admin_token>

{
  "notes": "Bukti transfer valid"
}
```

### Reject Transfer

```
POST /api/v1/admin/transfers/:transfer_id/reject
Authorization: Bearer <admin_token>

{
  "notes": "Jumlah tidak sesuai"
}
```

---

## Support

Jika ada pertanyaan tentang API ini, hubungi:

- **Backend Team**: backend@rakittiket.com
- **Technical Lead**: [Nama]

---

**Document Version:** 2.2.0  
**Last Updated:** 2026-04-12
