# RakitTiket Payment Gateway Admin API Documentation

**Version:** 1.0.0  
**Last Updated:** 2026-04-12  
**Base URL:** `http://localhost:8000/api/v1`

---

## Table of Contents

1. [Overview](#overview)
2. [Authentication](#authentication)
3. [Payment Gateway Management](#payment-gateway-management)
4. [Manual Transfer Configuration](#manual-transfer-configuration)
5. [Payment Options](#payment-options)
6. [Bank Account Management](#bank-account-management)
7. [Manual Transfer Verification](#manual-transfer-verification)
8. [Error Handling](#error-handling)
9. [Frontend Implementation Guide](#frontend-implementation-guide)

---

## Overview

Sistem Multi-Gateway Payment memungkinkan admin untuk mengkonfigurasi metode pembayaran yang tersedia di platform. Terdapat dua jenis metode pembayaran:

| Type | Code | Description |
|------|------|-------------|
| **Gateway** | `MIDTRANS`, `XENDIT`, `DOKU` | Pembayaran via third-party payment gateway |
| **Manual Transfer** | `MANUAL` | Pembayaran via transfer bank langsung, verifikasi oleh admin |

### Gateway Codes

| Code | Gateway Name | Status |
|------|-------------|--------|
| `MIDTRANS` | Midtrans | Implemented |
| `XENDIT` | Xendit | Placeholder (can be implemented) |
| `DOKU` | Doku | Placeholder (can be implemented) |

---

## Authentication

Semua endpoint admin memerlukan autentikasi JWT token dengan role admin.

### Required Headers

```
Authorization: Bearer <admin_jwt_token>
```

### Error Responses

| HTTP Code | Description |
|------------|-------------|
| `401` | Unauthorized - Token tidak valid atau tidak ada |
| `403` | Forbidden - User bukan admin |

---

## Payment Gateway Management

### 1. Get All Gateways

**Purpose:** Mendapatkan daftar semua payment gateway yang tersedia beserta statusnya.

```
GET /api/v1/admin/gateways
Authorization: Bearer <admin_token>
```

#### Response

```json
{
  "success": true,
  "data": [
    {
      "id": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
      "code": "MIDTRANS",
      "name": "Midtrans",
      "is_enabled": true,
      "is_active": true,
      "display_order": 1,
      "deleted": false,
      "data_hash": null,
      "created_at": "2026-04-12T00:00:00Z",
      "updated_at": "2026-04-12T00:00:00Z"
    },
    {
      "id": "b2c3d4e5-f6a7-8901-bcde-f23456789012",
      "code": "XENDIT",
      "name": "Xendit",
      "is_enabled": true,
      "is_active": false,
      "display_order": 2,
      "deleted": false,
      "data_hash": null,
      "created_at": "2026-04-12T00:00:00Z",
      "updated_at": "2026-04-12T00:00:00Z"
    },
    {
      "id": "c3d4e5f6-a7b8-9012-cdef-345678901234",
      "code": "DOKU",
      "name": "Doku",
      "is_enabled": false,
      "is_active": false,
      "display_order": 3,
      "deleted": false,
      "data_hash": null,
      "created_at": "2026-04-12T00:00:00Z",
      "updated_at": "2026-04-12T00:00:00Z"
    }
  ]
}
```

#### Field Descriptions

| Field | Type | Description |
|-------|------|-------------|
| `id` | string (UUID) | Unique identifier gateway |
| `code` | string | Kode gateway (MIDTRANS, XENDIT, DOKU) |
| `name` | string | Nama lengkap gateway |
| `is_enabled` | boolean | Apakah gateway bisa diaktifkan (konfigurasi sistem/ENV) |
| `is_active` | boolean | Apakah gateway sedang aktif digunakan |
| `display_order` | integer | Urutan tampil di frontend (1 = paling atas) |
| `deleted` | boolean | Soft delete flag |
| `created_at` | timestamp | Waktu pembuatan |
| `updated_at` | timestamp | Waktu update terakhir |

#### Notes

- `is_enabled` = false berarti gateway belum dikonfigurasi di sistem (belum ada credentials di ENV)
- Hanya satu gateway yang bisa `is_active` = true dalam satu waktu
- Gateway credentials (API keys) disimpan di ENV, bukan di database

---

### 2. Activate Gateway

**Purpose:** Mengaktifkan sebuah payment gateway. Gateway lain akan otomatis dinonaktifkan (hanya satu gateway aktif dalam satu waktu).

```
POST /api/v1/admin/gateways/:code/activate
Authorization: Bearer <admin_token>
```

#### URL Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `code` | string | Yes | Kode gateway (MIDTRANS, XENDIT, DOKU) |

#### Example Request

```bash
curl -X POST http://localhost:8000/api/v1/admin/gateways/MIDTRANS/activate \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

#### Response (Success - 200 OK)

```json
{
  "success": true,
  "message": "Gateway MIDTRANS activated successfully"
}
```

#### Response (Gateway Not Found - 404)

```json
{
  "success": false,
  "message": "Gateway not found"
}
```

#### Response (Gateway Not Enabled - 400)

```json
{
  "success": false,
  "message": "Gateway is not enabled"
}
```

#### Flow Diagram

```
┌─────────────────────────────────────────────────────────────────────┐
│                    ACTIVATION FLOW                                     │
├─────────────────────────────────────────────────────────────────────┤
│                                                                       │
│   Admin                     Backend                      Database      │
│     │                           │                            │        │
│     │  POST /gateways/         │                            │        │
│     │  MIDTRANS/activate       │                            │        │
│     │ ───────────────────────> │                            │        │
│     │                           │                            │        │
│     │                           │  1. Check gateway exists   │        │
│     │                           │ ───────────────────────> │        │
│     │                           │                            │        │
│     │                           │  2. Deactivate all        │        │
│     │                           │ ───────────────────────> │        │
│     │                           │    is_active = false       │        │
│     │                           │                            │        │
│     │                           │  3. Set MIDTRANS active   │        │
│     │                           │ ───────────────────────> │        │
│     │                           │    is_active = true       │        │
│     │                           │                            │        │
│     │  { success: true }        │                            │        │
│     │ <─────────────────────── │                            │        │
│     │                           │                            │        │
└─────────────────────────────────────────────────────────────────────┘
```

#### Notes

- Hanya satu gateway yang bisa aktif dalam satu waktu
- Jika gateway tidak ditemukan atau tidak di-enable di sistem, akan return error
- Credential gateway (API keys) harus sudah ada di ENV variable

---

### 3. Deactivate Gateway

**Purpose:** Menonaktifkan sebuah payment gateway.

```
POST /api/v1/admin/gateways/:code/deactivate
Authorization: Bearer <admin_token>
```

#### URL Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `code` | string | Yes | Kode gateway (MIDTRANS, XENDIT, DOKU) |

#### Example Request

```bash
curl -X POST http://localhost:8000/api/v1/admin/gateways/MIDTRANS/deactivate \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

#### Response (Success - 200 OK)

```json
{
  "success": true,
  "message": "Gateway MIDTRANS deactivated successfully"
}
```

#### Response (Gateway Not Found - 404)

```json
{
  "success": false,
  "message": "Gateway not found"
}
```

#### Notes

- Setelah dideactivate, user tidak bisa checkout menggunakan gateway tersebut
- Jika ada order yang sedang dalam proses dengan gateway tersebut, order tetap valid sampai expire

---

### 4. Set Gateway Display Order

**Purpose:** Mengatur urutan tampil gateway di halaman checkout.

```
PUT /api/v1/admin/gateways/:code/display-order
Authorization: Bearer <admin_token>
```

#### URL Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `code` | string | Yes | Kode gateway (MIDTRANS, XENDIT, DOKU) |

#### Example Request

```bash
curl -X PUT http://localhost:8000/api/v1/admin/gateways/XENDIT/display-order \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." \
  -H "Content-Type: application/json" \
  -d '{
    "display_order": 1
  }'
```

#### Request Body

```json
{
  "display_order": 1
}
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `display_order` | integer | Yes | Urutan tampil (1 = paling atas) |

#### Response (Success - 200 OK)

```json
{
  "success": true,
  "message": "Display order updated successfully"
}
```

#### Response (Bad Request - 400)

```json
{
  "success": false,
  "message": "invalid request body"
}
```

#### Notes

- Display order lebih rendah = tampil lebih dulu di frontend
- Ini juga mempengaruhi urutan di `GET /payment-options`

---

## Manual Transfer Configuration

### 5. Enable Manual Transfer

**Purpose:** Mengaktifkan opsi pembayaran manual transfer.

```
POST /api/v1/admin/manual-transfer/enable
Authorization: Bearer <admin_token>
```

#### Example Request

```bash
curl -X POST http://localhost:8000/api/v1/admin/manual-transfer/enable \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

#### Response (Success - 200 OK)

```json
{
  "success": true,
  "message": "Manual transfer enabled successfully"
}
```

#### Notes

- Manual transfer default DISABLED saat pertama kali sistem di-setup
- Setelah di-enable, user bisa memilih pembayaran via transfer manual
- Admin harus menambahkan bank accounts terlebih dahulu via endpoint `/bank-accounts`

---

### 6. Disable Manual Transfer

**Purpose:** Menonaktifkan opsi pembayaran manual transfer.

```
POST /api/v1/admin/manual-transfer/disable
Authorization: Bearer <admin_token>
```

#### Example Request

```bash
curl -X POST http://localhost:8000/api/v1/admin/manual-transfer/disable \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

#### Response (Success - 200 OK)

```json
{
  "success": true,
  "message": "Manual transfer disabled successfully"
}
```

#### Notes

- User tidak akan melihat opsi "Transfer Manual" di halaman checkout
- Order yang sudah menggunakan manual transfer tetap valid

---

### 7. Set Manual Transfer Display Order

**Purpose:** Mengatur urutan tampil manual transfer di halaman checkout.

```
PUT /api/v1/admin/manual-transfer/display-order
Authorization: Bearer <admin_token>
```

#### Example Request

```bash
curl -X PUT http://localhost:8000/api/v1/admin/manual-transfer/display-order \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." \
  -H "Content-Type: application/json" \
  -d '{
    "display_order": 2
  }'
```

#### Request Body

```json
{
  "display_order": 2
}
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `display_order` | integer | Yes | Urutan tampil (1 = paling atas) |

#### Response (Success - 200 OK)

```json
{
  "success": true,
  "message": "Manual transfer display order updated successfully"
}
```

---

## Payment Options

### 8. Get Active Payment Options (Public)

**Purpose:** Mendapatkan daftar metode pembayaran yang aktif untuk ditampilkan di frontend. Endpoint ini PUBLIC (tidak memerlukan autentikasi).

```
GET /api/v1/payment-options
```

#### Example Request

```bash
curl http://localhost:8000/api/v1/payment-options
```

#### Response (Both Gateway & Manual Enabled)

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

#### Response (Only Gateway Enabled)

```json
{
  "success": true,
  "data": [
    {
      "type": "GATEWAY",
      "code": "MIDTRANS",
      "gateway_name": "Midtrans",
      "display_order": 1
    }
  ]
}
```

#### Response (Only Manual Enabled)

```json
{
  "success": true,
  "data": [
    {
      "type": "MANUAL",
      "code": "MANUAL",
      "gateway_name": "Transfer Manual",
      "display_order": 1
    }
  ]
}
```

#### Response (No Payment Options)

```json
{
  "success": true,
  "data": []
}
```

#### Field Descriptions

| Field | Type | Description |
|-------|------|-------------|
| `type` | string | Jenis payment: `GATEWAY` atau `MANUAL` |
| `code` | string | Kode spesifik: MIDTRANS, XENDIT, DOKU, atau MANUAL |
| `gateway_name` | string | Nama yang ditampilkan ke user |
| `display_order` | integer | Urutan tampil |

#### Notes

- Hanya payment method yang AKTIF yang ditampilkan
- Hasil diurutkan berdasarkan `display_order` ascending
- Frontend WAJIB menggunakan endpoint ini untuk menampilkan opsi pembayaran

---

## Bank Account Management

### 9. Get Bank Accounts (Public)

**Purpose:** Mendapatkan daftar rekening bank aktif untuk pembayaran manual transfer.

```
GET /api/v1/bank-accounts
```

#### Example Request

```bash
curl http://localhost:8000/api/v1/bank-accounts
```

#### Response

```json
{
  "success": true,
  "data": [
    {
      "id": "bank-uuid-1234-5678-abcd-ef1234567890",
      "bank_name": "Bank Central Asia (BCA)",
      "bank_code": "BCA",
      "account_number": "1234567890",
      "account_holder": "PT Rakit Tiket Indonesia",
      "is_active": true,
      "is_default": true,
      "instruction_text": "Transfer tepat hingga 3 digit terakhir untuk加快了验证"
    },
    {
      "id": "bank-uuid-2345-6789-bcde-f12345678901",
      "bank_name": "Bank Mandiri",
      "bank_code": "MANDIRI",
      "account_number": "1300098765432",
      "account_holder": "PT Rakit Tiket Indonesia",
      "is_active": true,
      "is_default": false,
      "instruction_text": null
    }
  ]
}
```

#### Field Descriptions

| Field | Type | Description |
|-------|------|-------------|
| `id` | string (UUID) | Unique identifier rekening |
| `bank_name` | string | Nama bank lengkap |
| `bank_code` | string | Kode bank (BCA, MANDIRI, dll) |
| `account_number` | string | Nomor rekening |
| `account_holder` | string | Nama pemilik rekening |
| `is_active` | boolean | Apakah aktif |
| `is_default` | boolean | Apakah rekening default |
| `instruction_text` | string/null | Instruksi khusus untuk transfer |

---

### 10. Create Bank Account (Admin)

**Purpose:** Menambahkan rekening bank baru untuk manual transfer.

```
POST /api/v1/admin/bank-accounts
Authorization: Bearer <admin_token>
```

#### Example Request

```bash
curl -X POST http://localhost:8000/api/v1/admin/bank-accounts \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." \
  -H "Content-Type: application/json" \
  -d '{
    "bank_name": "Bank Central Asia (BCA)",
    "bank_code": "BCA",
    "account_number": "1234567890",
    "account_holder": "PT Rakit Tiket Indonesia",
    "is_default": true,
    "instruction_text": "Transfer tepat hingga 3 digit terakhir"
  }'
```

#### Request Body

```json
{
  "bank_name": "Bank Central Asia (BCA)",
  "bank_code": "BCA",
  "account_number": "1234567890",
  "account_holder": "PT Rakit Tiket Indonesia",
  "is_default": true,
  "instruction_text": "Transfer tepat hingga 3 digit terakhir"
}
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `bank_name` | string | Yes | Nama bank lengkap |
| `bank_code` | string | Yes | Kode bank (BCA, MANDIRI, dll) |
| `account_number` | string | Yes | Nomor rekening |
| `account_holder` | string | Yes | Nama pemilik rekening |
| `is_default` | boolean | No | Set sebagai default (default: false) |
| `instruction_text` | string | No | Instruksi khusus |

#### Response (Success - 201 Created)

```json
{
  "success": true,
  "message": "Bank account created successfully",
  "data": {
    "id": "new-bank-uuid-1234-5678",
    "bank_name": "Bank Central Asia (BCA)",
    "bank_code": "BCA",
    "account_number": "1234567890",
    "account_holder": "PT Rakit Tiket Indonesia",
    "is_active": true,
    "is_default": true,
    "instruction_text": "Transfer tepat hingga 3 digit terakhir"
  }
}
```

---

### 11. Update Bank Account (Admin)

**Purpose:** Mengupdate data rekening bank.

```
PUT /api/v1/admin/bank-accounts/:bank_account_id
Authorization: Bearer <admin_token>
```

#### Example Request

```bash
curl -X PUT http://localhost:8000/api/v1/admin/bank-accounts/bank-uuid-1234-5678 \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." \
  -H "Content-Type: application/json" \
  -d '{
    "bank_name": "Bank Central Asia (BCA)",
    "bank_code": "BCA",
    "account_number": "9876543210",
    "account_holder": "PT Rakit Tiket Indonesia",
    "instruction_text": "Updated instruction text"
  }'
```

#### Response (Success - 200 OK)

```json
{
  "success": true,
  "message": "Bank account updated successfully"
}
```

---

### 12. Delete Bank Account (Admin)

**Purpose:** Menghapus rekening bank.

```
DELETE /api/v1/admin/bank-accounts/:bank_account_id
Authorization: Bearer <admin_token>
```

#### Example Request

```bash
curl -X DELETE http://localhost:8000/api/v1/admin/bank-accounts/bank-uuid-1234-5678 \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

#### Response (Success - 200 OK)

```json
{
  "success": true,
  "message": "Bank account deleted successfully"
}
```

---

## Manual Transfer Verification

### 13. Get Pending Transfers (Admin)

**Purpose:** Mendapatkan daftar transfer yang menunggu verifikasi.

```
GET /api/v1/admin/transfers/pending
Authorization: Bearer <admin_token>
```

#### Example Request

```bash
curl http://localhost:8000/api/v1/admin/transfers/pending \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

#### Response

```json
{
  "success": true,
  "data": [
    {
      "id": "transfer-uuid-1234-5678",
      "order_id": "order-uuid-abcd-efgh",
      "bank_account_id": "bank-uuid-1234-5678",
      "transfer_amount": 250000,
      "transfer_proof_url": "/uploads/transfer_proof/xxx/ref",
      "transfer_proof_filename": "bukti-transfer.jpg",
      "sender_name": "John Doe",
      "sender_account_number": "1234567890",
      "transfer_date": "2026-04-12T14:30:00Z",
      "admin_notes": null,
      "reviewed_by": null,
      "reviewed_at": null,
      "status": "PENDING",
      "bank_account": {
        "id": "bank-uuid-1234-5678",
        "bank_name": "Bank Central Asia (BCA)",
        "bank_code": "BCA",
        "account_number": "1234567890",
        "account_holder": "PT Rakit Tiket Indonesia"
      },
      "order": {
        "id": "order-uuid-abcd-efgh",
        "order_number": "TKT-2026-a1b2c3d4",
        "amount": 250000,
        "payment_status": "pending"
      },
      "registrant": {
        "id": "registrant-uuid-1234",
        "name": "John Doe",
        "email": "john@example.com",
        "phone": "081234567890"
      }
    }
  ],
  "count": 1
}
```

#### Transfer Status Values

| Status | Description |
|--------|-------------|
| `PENDING` | Menunggu verifikasi admin |
| `APPROVED` | Sudah diverifikasi dan disetujui |
| `REJECTED` | Ditolak oleh admin |

---

### 14. Approve Transfer (Admin)

**Purpose:** Menyetujui transfer dan mengubah status order menjadi paid.

```
POST /api/v1/admin/transfers/:transfer_id/approve
Authorization: Bearer <admin_token>
```

#### Example Request

```bash
curl -X POST http://localhost:8000/api/v1/admin/transfers/transfer-uuid-1234-5678/approve \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." \
  -H "Content-Type: application/json" \
  -d '{
    "notes": "Bukti transfer valid, jumlah sesuai"
  }'
```

#### Request Body

```json
{
  "notes": "Bukti transfer valid, jumlah sesuai"
}
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `notes` | string | No | Catatan dari admin |

#### Response (Success - 200 OK)

```json
{
  "success": true,
  "message": "Transfer approved successfully"
}
```

#### What happens on approval:
1. Order status changed to `paid`
2. Tickets marked as sold (stock confirmed)
3. E-ticket PDF generated
4. Confirmation email sent to customer

#### Response (Not Found - 404)

```json
{
  "success": false,
  "message": "Transfer not found"
}
```

#### Response (Bad Request - 400)

```json
{
  "success": false,
  "message": "Transfer is not in pending status"
}
```

---

### 15. Reject Transfer (Admin)

**Purpose:** Menolak transfer.

```
POST /api/v1/admin/transfers/:transfer_id/reject
Authorization: Bearer <admin_token>
```

#### Example Request

```bash
curl -X POST http://localhost:8000/api/v1/admin/transfers/transfer-uuid-1234-5678/reject \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." \
  -H "Content-Type: application/json" \
  -d '{
    "notes": "Jumlah transfer tidak sesuai dengan tagihan"
  }'
```

#### Request Body

```json
{
  "notes": "Jumlah transfer tidak sesuai dengan tagihan"
}
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `notes` | string | No | Alasan penolakan |

#### Response (Success - 200 OK)

```json
{
  "success": true,
  "message": "Transfer rejected"
}
```

#### What happens on rejection:
1. Transfer status changed to `REJECTED`
2. Order remains `pending` (user can resubmit correct proof)
3. No email sent (user needs to resubmit)

---

## Error Handling

### HTTP Status Codes

| Code | Meaning | Common Causes |
|------|---------|---------------|
| `200` | Success | Request berhasil |
| `201` | Created | Resource berhasil dibuat |
| `400` | Bad Request | Invalid request body / parameters |
| `401` | Unauthorized | Token tidak valid atau tidak ada |
| `403` | Forbidden | User bukan admin |
| `404` | Not Found | Gateway / Transfer tidak ditemukan |
| `500` | Server Error | Internal server error |

### Error Response Format

```json
{
  "success": false,
  "message": "Error description here"
}
```

### Common Error Messages

| Error Message | HTTP Code | Description |
|---------------|-----------|-------------|
| `Gateway not found` | 404 | Gateway dengan code tersebut tidak ada |
| `Gateway is not enabled` | 400 | Gateway belum di-enable di sistem |
| `Transfer not found` | 404 | Transfer dengan ID tersebut tidak ada |
| `Transfer is not in pending status` | 400 | Transfer sudah di-approve/reject |
| `Failed to enable manual transfer` | 500 | Database error |
| `Failed to fetch gateways` | 500 | Database error |

---

## Frontend Implementation Guide

### 1. Fetching Payment Options

```javascript
// Fungsi untuk mendapatkan opsi pembayaran
async function getPaymentOptions() {
  try {
    const response = await fetch('/api/v1/payment-options');
    const result = await response.json();
    
    if (result.success) {
      return result.data; // Array of payment options
    }
    throw new Error(result.message);
  } catch (error) {
    console.error('Failed to fetch payment options:', error);
    return [];
  }
}

// Usage
const options = await getPaymentOptions();
// options = [
//   { type: "GATEWAY", code: "MIDTRANS", gateway_name: "Midtrans", display_order: 1 },
//   { type: "MANUAL", code: "MANUAL", gateway_name: "Transfer Manual", display_order: 2 }
// ]
```

### 2. Displaying Payment Options

```javascript
function renderPaymentOptions(options) {
  const container = document.getElementById('payment-options');
  
  // Sort berdasarkan display_order
  const sorted = [...options].sort((a, b) => a.display_order - b.display_order);
  
  sorted.forEach(option => {
    const div = document.createElement('div');
    div.className = 'payment-option';
    
    if (option.type === 'GATEWAY') {
      div.innerHTML = `
        <input type="radio" name="payment" value="${option.code}" id="${option.code}">
        <label for="${option.code}">
          <strong>${option.gateway_name}</strong>
          <small>Bayar dengan ${option.gateway_name}</small>
        </label>
      `;
    } else if (option.type === 'MANUAL') {
      div.innerHTML = `
        <input type="radio" name="payment" value="${option.code}" id="${option.code}">
        <label for="${option.code}">
          <strong>${option.gateway_name}</strong>
          <small>Transfer ke rekening bank yang tersedia</small>
        </label>
      `;
    }
    
    container.appendChild(div);
  });
}
```

### 3. Checkout Based on Selected Payment

```javascript
async function initiateCheckout(orderId, paymentType) {
  try {
    const response = await fetch(`/api/v1/checkout/${orderId}`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json'
      },
      body: JSON.stringify({
        payment_type: paymentType  // "MIDTRANS", "XENDIT", "DOKU", atau "MANUAL"
      })
    });
    
    const result = await response.json();
    
    if (!result.success) {
      throw new Error(result.message);
    }
    
    return result.data;
  } catch (error) {
    console.error('Checkout failed:', error);
    throw error;
  }
}

// Usage
const selectedPayment = document.querySelector('input[name="payment"]:checked').value;
const checkoutData = await initiateCheckout(orderId, selectedPayment);

// Handle berdasarkan payment_type
if (checkoutData.payment_type === 'GATEWAY') {
  // Redirect ke payment_url (Midtrans/Xendit/Doku)
  window.location.href = checkoutData.payment_info.payment_url;
} else if (checkoutData.payment_type === 'MANUAL') {
  // Tampilkan rekening bank
  showBankAccounts(checkoutData.bank_accounts);
}
```

### 4. Checkout Response Structures

#### Gateway Checkout Response

```javascript
// POST /api/v1/checkout/:order_id
// Request: { payment_type: "MIDTRANS" }

{
  "success": true,
  "data": {
    "order_id": "0676566f-45b1-11e7-8f91-de3d5d3d1f9f",
    "order_number": "TKT-2026-a1b2c3d4e5f6",
    "amount": 250000,
    "payment_type": "GATEWAY",
    "payment_status": "pending",
    "expires_at": "2026-04-12T15:30:00Z",
    "payment_info": {
      "payment_url": "https://app.sandbox.midtrans.com/snap/v2/vtweb/a1b2c3d4...",
      "payment_token": "a1b2c3d4-5678-90ef",
      "payment_method": "bank_transfer"
    }
  }
}
```

#### Manual Transfer Checkout Response

```javascript
// POST /api/v1/checkout/:order_id
// Request: { payment_type: "MANUAL" }

{
  "success": true,
  "data": {
    "order_id": "0676566f-45b1-11e7-8f91-de3d5d3d1f9f",
    "order_number": "TKT-2026-a1b2c3d4e5f6",
    "amount": 250000,
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

### 5. Admin Dashboard - Gateway Management

```javascript
// Fetch all gateways (admin only)
async function getAllGateways(token) {
  const response = await fetch('/api/v1/admin/gateways', {
    headers: {
      'Authorization': `Bearer ${token}`
    }
  });
  
  return response.json();
}

// Activate a gateway
async function activateGateway(code, token) {
  const response = await fetch(`/api/v1/admin/gateways/${code}/activate`, {
    method: 'POST',
    headers: {
      'Authorization': `Bearer ${token}`
    }
  });
  
  return response.json();
}

// Deactivate a gateway
async function deactivateGateway(code, token) {
  const response = await fetch(`/api/v1/admin/gateways/${code}/deactivate`, {
    method: 'POST',
    headers: {
      'Authorization': `Bearer ${token}`
    }
  });
  
  return response.json();
}

// Update display order
async function updateGatewayOrder(code, order, token) {
  const response = await fetch(`/api/v1/admin/gateways/${code}/display-order`, {
    method: 'PUT',
    headers: {
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${token}`
    },
    body: JSON.stringify({ display_order: order })
  });
  
  return response.json();
}
```

### 6. Admin Dashboard - Manual Transfer Management

```javascript
// Enable manual transfer
async function enableManualTransfer(token) {
  const response = await fetch('/api/v1/admin/manual-transfer/enable', {
    method: 'POST',
    headers: {
      'Authorization': `Bearer ${token}`
    }
  });
  
  return response.json();
}

// Disable manual transfer
async function disableManualTransfer(token) {
  const response = await fetch('/api/v1/admin/manual-transfer/disable', {
    method: 'POST',
    headers: {
      'Authorization': `Bearer ${token}`
    }
  });
  
  return response.json();
}

// Update display order
async function updateManualTransferOrder(order, token) {
  const response = await fetch('/api/v1/admin/manual-transfer/display-order', {
    method: 'PUT',
    headers: {
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${token}`
    },
    body: JSON.stringify({ display_order: order })
  });
  
  return response.json();
}
```

### 7. Admin - Bank Account Management

```javascript
// Get bank accounts
async function getBankAccounts() {
  const response = await fetch('/api/v1/bank-accounts');
  return response.json();
}

// Create bank account
async function createBankAccount(data, token) {
  const response = await fetch('/api/v1/admin/bank-accounts', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${token}`
    },
    body: JSON.stringify(data)
  });
  
  return response.json();
}

// Update bank account
async function updateBankAccount(id, data, token) {
  const response = await fetch(`/api/v1/admin/bank-accounts/${id}`, {
    method: 'PUT',
    headers: {
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${token}`
    },
    body: JSON.stringify(data)
  });
  
  return response.json();
}

// Delete bank account
async function deleteBankAccount(id, token) {
  const response = await fetch(`/api/v1/admin/bank-accounts/${id}`, {
    method: 'DELETE',
    headers: {
      'Authorization': `Bearer ${token}`
    }
  });
  
  return response.json();
}
```

### 8. Admin - Transfer Verification

```javascript
// Get pending transfers
async function getPendingTransfers(token) {
  const response = await fetch('/api/v1/admin/transfers/pending', {
    headers: {
      'Authorization': `Bearer ${token}`
    }
  });
  
  return response.json();
}

// Approve transfer
async function approveTransfer(transferId, notes, token) {
  const response = await fetch(`/api/v1/admin/transfers/${transferId}/approve`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${token}`
    },
    body: JSON.stringify({ notes })
  });
  
  return response.json();
}

// Reject transfer
async function rejectTransfer(transferId, notes, token) {
  const response = await fetch(`/api/v1/admin/transfers/${transferId}/reject`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${token}`
    },
    body: JSON.stringify({ notes })
  });
  
  return response.json();
}
```

### 9. React Admin Dashboard Example

```jsx
// GatewayManagement.jsx
import React, { useState, useEffect } from 'react';

function GatewayManagement() {
  const [gateways, setGateways] = useState([]);
  const [manualEnabled, setManualEnabled] = useState(false);
  const [loading, setLoading] = useState(true);
  
  useEffect(() => {
    loadData();
  }, []);
  
  async function loadData() {
    try {
      setLoading(true);
      
      // Load gateways
      const gatewayResult = await getAllGateways(token);
      setGateways(gatewayResult.data);
      
      // Check manual transfer status
      const options = await getPaymentOptions();
      const manual = options.find(o => o.type === 'MANUAL');
      setManualEnabled(!!manual);
    } catch (error) {
      console.error('Failed to load data:', error);
    } finally {
      setLoading(false);
    }
  }
  
  async function handleActivate(code) {
    try {
      await activateGateway(code, token);
      await loadData(); // Refresh
    } catch (error) {
      alert('Failed to activate: ' + error.message);
    }
  }
  
  async function handleDeactivate(code) {
    try {
      await deactivateGateway(code, token);
      await loadData(); // Refresh
    } catch (error) {
      alert('Failed to deactivate: ' + error.message);
    }
  }
  
  async function handleToggleManual() {
    try {
      if (manualEnabled) {
        await disableManualTransfer(token);
      } else {
        await enableManualTransfer(token);
      }
      await loadData(); // Refresh
    } catch (error) {
      alert('Failed to toggle manual transfer: ' + error.message);
    }
  }
  
  if (loading) return <div>Loading...</div>;
  
  return (
    <div className="gateway-management p-4">
      <h2 className="text-2xl font-bold mb-4">Payment Gateway Configuration</h2>
      
      {/* Gateway Table */}
      <table className="min-w-full bg-white border">
        <thead>
          <tr className="bg-gray-100">
            <th className="px-4 py-2 border">Gateway</th>
            <th className="px-4 py-2 border">Code</th>
            <th className="px-4 py-2 border">Enabled</th>
            <th className="px-4 py-2 border">Active</th>
            <th className="px-4 py-2 border">Order</th>
            <th className="px-4 py-2 border">Actions</th>
          </tr>
        </thead>
        <tbody>
          {gateways.map(gw => (
            <tr key={gw.id}>
              <td className="px-4 py-2 border">{gw.name}</td>
              <td className="px-4 py-2 border">{gw.code}</td>
              <td className="px-4 py-2 border">
                {gw.is_enabled ? '✓' : '✗'}
              </td>
              <td className="px-4 py-2 border">
                {gw.is_active ? '✓' : '-'}
              </td>
              <td className="px-4 py-2 border">{gw.display_order}</td>
              <td className="px-4 py-2 border">
                {gw.is_enabled && (
                  <>
                    {gw.is_active ? (
                      <button 
                        onClick={() => handleDeactivate(gw.code)}
                        className="bg-red-500 text-white px-2 py-1 rounded mr-2"
                      >
                        Deactivate
                      </button>
                    ) : (
                      <button 
                        onClick={() => handleActivate(gw.code)}
                        className="bg-green-500 text-white px-2 py-1 rounded mr-2"
                      >
                        Activate
                      </button>
                    )}
                  </>
                )}
              </td>
            </tr>
          ))}
        </tbody>
      </table>
      
      {/* Manual Transfer Section */}
      <div className="mt-6 p-4 bg-gray-50 rounded">
        <h3 className="text-lg font-semibold mb-2">Manual Transfer</h3>
        <p className="mb-2">Status: {manualEnabled ? 'Enabled' : 'Disabled'}</p>
        <button 
          onClick={handleToggleManual}
          className={`px-4 py-2 rounded ${
            manualEnabled 
              ? 'bg-red-500 text-white' 
              : 'bg-green-500 text-white'
          }`}
        >
          {manualEnabled ? 'Disable' : 'Enable'}
        </button>
      </div>
    </div>
  );
}

export default GatewayManagement;
```

---

## Complete Admin API Reference

| Method | Endpoint | Description | Auth Required |
|--------|----------|-------------|---------------|
| `GET` | `/api/v1/admin/gateways` | List all gateways | Yes |
| `POST` | `/api/v1/admin/gateways/:code/activate` | Activate a gateway | Yes |
| `POST` | `/api/v1/admin/gateways/:code/deactivate` | Deactivate a gateway | Yes |
| `PUT` | `/api/v1/admin/gateways/:code/display-order` | Set gateway display order | Yes |
| `POST` | `/api/v1/admin/manual-transfer/enable` | Enable manual transfer | Yes |
| `POST` | `/api/v1/admin/manual-transfer/disable` | Disable manual transfer | Yes |
| `PUT` | `/api/v1/admin/manual-transfer/display-order` | Set manual transfer display order | Yes |
| `GET` | `/api/v1/admin/transfers/pending` | List pending transfers | Yes |
| `POST` | `/api/v1/admin/transfers/:id/approve` | Approve transfer | Yes |
| `POST` | `/api/v1/admin/transfers/:id/reject` | Reject transfer | Yes |
| `GET` | `/api/v1/admin/bank-accounts` | List bank accounts | Yes |
| `POST` | `/api/v1/admin/bank-accounts` | Create bank account | Yes |
| `PUT` | `/api/v1/admin/bank-accounts/:id` | Update bank account | Yes |
| `DELETE` | `/api/v1/admin/bank-accounts/:id` | Delete bank account | Yes |

## Public API Reference

| Method | Endpoint | Description | Auth Required |
|--------|----------|-------------|---------------|
| `GET` | `/api/v1/payment-options` | Get active payment options | No |
| `GET` | `/api/v1/bank-accounts` | List active bank accounts | No |
| `POST` | `/api/v1/checkout/:order_id` | Initiate checkout | No |
| `POST` | `/api/v1/transfers/proof` | Submit transfer proof | No |

---

**Document Version:** 1.0.0  
**Last Updated:** 2026-04-12
