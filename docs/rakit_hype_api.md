# Rakit-Hype API Documentation

Module untuk urgency selling, FOMO, dan flash sale features.

## Table of Contents

1. [Get Active Tickets (Hype Display)](#1-get-active-tickets-hype-display)
2. [Check Ticket Availability](#2-check-ticket-availability)
3. [Set Flash Sale](#3-set-flash-sale)
4. [Disable Flash Sale](#4-disable-flash-sale)
5. [Set Countdown Timer](#5-set-countdown-timer)
6. [Set Stock Alert](#6-set-stock-alert)

---

## 1. Get Active Tickets (Hype Display)

Get semua ticket untuk satu event dengan computed display fields (hype info).

### Request

```
GET /api/v1/events/:event_id/hype
```

### Path Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| event_id | string (UUID) | Yes | Event ID |

### Response

```json
{
  "success": true,
  "data": [
    {
      "id": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
      "event_id": "882487e7-c3b5-44e4-aac5-7aa8d473ba8e",
      "type": "GOLD",
      "title": "Tiket Gold - Early Bird",
      "status": "AVAILABLE",
      "total": 100,
      "available_qty": 15,
      "sold_qty": 85,
      "original_price": 200000,
      "current_price": 150000,
      "is_flash_sale": true,
      "flash_sale_price": 150000,
      "flash_end_time": "2026-04-26T23:59:59Z",
      "is_on_flash_sale": true,
      "is_low_stock": true,
      "low_stock_message": "Sisa 15 tiket!",
      "stock_remaining": 15,
      "stock_percentage": 15,
      "is_available": true,
      "unavailable_reason": null,
      "countdown_seconds": 3600,
      "countdown_end": "2026-04-26T22:00:00Z",
      "show_countdown": true,
      "sale_start_time": "2026-04-01T00:00:00Z",
      "sale_end_time": "2026-04-26T22:00:00Z",
      "urgent_message": "Sisa 15 tiket!"
    },
    {
      "id": "b2c3d4e5-f6a7-8901-bcde-f23456789012",
      "event_id": "882487e7-c3b5-44e4-aac5-7aa8d473ba8e",
      "type": "VIP",
      "title": "Tiket VIP - Premium",
      "status": "AVAILABLE",
      "total": 50,
      "available_qty": 50,
      "sold_qty": 0,
      "original_price": 500000,
      "current_price": 500000,
      "is_flash_sale": false,
      "flash_sale_price": null,
      "flash_end_time": null,
      "is_on_flash_sale": false,
      "is_low_stock": false,
      "low_stock_message": null,
      "stock_remaining": 50,
      "stock_percentage": 100,
      "is_available": true,
      "unavailable_reason": null,
      "countdown_seconds": null,
      "countdown_end": null,
      "show_countdown": false,
      "sale_start_time": null,
      "sale_end_time": null,
      "urgent_message": null
    }
  ],
  "count": 2
}
```

### Field Description

| Field | Type | Description |
|-------|------|-------------|
| original_price | float64 | Harga normal ticket |
| current_price | float64 | Harga saat ini (regular atau flash sale) |
| is_flash_sale | bool | Apakah ticket memiliki config flash sale |
| is_on_flash_sale | bool | Apakah flash sale sedang aktif sekarang |
| flash_sale_price | float64 | Harga selama flash sale |
| flash_end_time | string | Waktu berakhir flash sale (ISO 8601) |
| is_low_stock | bool | Apakah stock < threshold (20% default) |
| low_stock_message | string | Message "Sisa X tiket!" |
| stock_remaining | int | Jumlah ticket tersisa |
| stock_percentage | int | Persentase stock tersisa (0-100) |
| is_available | bool | Apakah ticket bisa dibeli sekarang |
| unavailable_reason | string | Alasan tidak tersedia (SOLD_OUT, SALE_NOT_STARTED, SALE_ENDED) |
| countdown_seconds | int64 | Detik tersisa untuk countdown |
| countdown_end | string | Waktu akhir countdown (ISO 8601) |
| show_countdown | bool | Apakah menampilkan countdown |
| urgent_message | string | Message urgency (low stock atau flash sale) |

---

## 2. Check Ticket Availability

Check apakah ticket tersedia dan harga saat ini sebelum checkout.

### Request

```
GET /api/v1/hype/check/:ticket_id
```

### Path Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| ticket_id | string (UUID) | Yes | Ticket ID |

### Response

#### Success - Available

```json
{
  "success": true,
  "data": {
    "available": true,
    "reason": "",
    "current_price": 150000,
    "is_flash_sale": true,
    "flash_sale_price": 150000
  }
}
```

#### Success - Not Available (Flash Sale Ended)

```json
{
  "success": true,
  "data": {
    "available": false,
    "reason": "SALE_NOT_STARTED",
    "current_price": 200000,
    "is_flash_sale": false,
    "flash_sale_price": null
  }
}
```

#### Sold Out

```json
{
  "success": true,
  "data": {
    "available": false,
    "reason": "SOLD_OUT",
    "current_price": 200000,
    "is_flash_sale": false,
    "flash_sale_price": null
  }
}
```

### Availability Reasons

| Reason | Description |
|--------|-------------|
| (empty) | Ticket tersedia |
| SOLD_OUT | Stock habis |
| SALE_NOT_STARTED | Penjualan belum dimulai (sale_start_time di masa depan) |
| SALE_ENDED | Penjualan sudah berakhir (sale_end_time di masa lalu) |

---

## 3. Set Flash Sale

Aktifkan flash sale untuk satu ticket.

### Request

```
POST /api/v1/admin/hype/flash-sale
Authorization: Bearer <admin_token>
Content-Type: application/json
```

### Request Body

```json
{
  "ticket_id": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
  "flash_price": 150000,
  "start_time": "2026-04-26T18:00:00Z",
  "end_time": "2026-04-26T23:59:59Z"
}
```

### Request Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| ticket_id | string (UUID) | Yes | Ticket ID |
| flash_price | float64 | Yes | Harga selama flash sale |
| start_time | string (ISO 8601) | No | Waktu mulai flash sale. Jika null, langsung aktif |
| end_time | string (ISO 8601) | No | Waktu berakhir flash sale. Jika null, tidak terbatas |

### Response

```json
{
  "success": true,
  "message": "Flash sale berhasil diaktifkan"
}
```

### Scenarios

#### Immediate Flash Sale (start_time = null)
```json
{
  "ticket_id": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
  "flash_price": 150000,
  "start_time": null,
  "end_time": "2026-04-26T23:59:59Z"
}
```

#### Scheduled Flash Sale
```json
{
  "ticket_id": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
  "flash_price": 150000,
  "start_time": "2026-04-27T10:00:00Z",
  "end_time": "2026-04-27T12:00:00Z"
}
```

---

## 4. Disable Flash Sale

Nonaktifkan flash sale untuk satu ticket.

### Request

```
DELETE /api/v1/admin/hype/flash-sale/:ticket_id
Authorization: Bearer <admin_token>
```

### Path Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| ticket_id | string (UUID) | Yes | Ticket ID |

### Response

```json
{
  "success": true,
  "message": "Flash sale berhasil dinonaktifkan"
}
```

---

## 5. Set Countdown Timer

Set waktu akhir countdown untuk ticket.

### Request

```
PUT /api/v1/admin/hype/countdown
Authorization: Bearer <admin_token>
Content-Type: application/json
```

### Request Body

```json
{
  "ticket_id": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
  "end_time": "2026-04-26T22:00:00Z"
}
```

### Request Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| ticket_id | string (UUID) | Yes | Ticket ID |
| end_time | string (ISO 8601) | Yes | Waktu akhir countdown |

### Response

```json
{
  "success": true,
  "message": "Countdown berhasil disetting"
}
```

### Notes

- Countdown akan override `sale_end_time` untuk display purpose
- Jika `end_time` null, countdown akan disembunyikan
- Frontend akan menghitung `countdown_seconds` = `end_time - now`

---

## 6. Set Stock Alert

Set threshold dan visibility untuk stock alert.

### Request

```
PUT /api/v1/admin/hype/stock-alert
Authorization: Bearer <admin_token>
Content-Type: application/json
```

### Request Body

```json
{
  "ticket_id": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
  "low_stock_threshold": 20,
  "show_stock_alert": true
}
```

### Request Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| ticket_id | string (UUID) | Yes | Ticket ID |
| low_stock_threshold | int | No | Threshold percentage (0-100). Default: 20 |
| show_stock_alert | bool | No | Apakah menampilkan stock alert. Default: true |

### Response

```json
{
  "success": true,
  "message": "Stock alert berhasil disetting"
}
```

### Stock Alert Logic

```
is_low_stock = show_stock_alert AND (available_qty * 100 / total) <= low_stock_threshold
```

#### Example

- Total: 100
- Available: 15
- Threshold: 20%
- Calculation: (15 * 100 / 100) = 15% <= 20% → is_low_stock = true

---

## Error Responses

### 400 Bad Request

```json
{
  "message": "ticket_id is required"
}
```

### 401 Unauthorized

```json
{
  "message": "Invalid or missing authentication token"
}
```

### 403 Forbidden

```json
{
  "message": "Admin access required"
}
```

### 404 Not Found

```json
{
  "message": "ticket tidak ditemukan"
}
```

### 500 Internal Server Error

```json
{
  "message": "Internal server error message"
}
```

---

## Integration with Checkout

Sebelum membuat order, frontend/client harus:

1. Call `GET /api/v1/hype/check/:ticket_id`
2. Verify `available: true`
3. Use `current_price` sebagai harga final

### Checkout Validation Flow

```
Client -> POST /api/v1/hype/check/:ticket_id
       <- { available: true, current_price: 150000 }

Client -> POST /api/v1/checkout dengan price dari response di atas
       <- Order created dengan price validation
```

---

## Frontend Display Examples

### Low Stock Badge
```
┌─────────────────────────┐
│  🎉 Sisa 15 tiket!      │
│  Stok Terbatas!         │
└─────────────────────────┘
```

### Flash Sale Countdown
```
┌─────────────────────────┐
│  ⚡ FLASH SALE          │
│ Harga spesial: Rp150k   │
│  Berakhir dalam: 02:30  │
└─────────────────────────┘
```

### Combined Display
```
┌─────────────────────────┐
│  ⚡ FLASH SALE          │
│  🎉 Sisa 15 tiket!      │
│  Rp150.000 (Rp200.000)  │
│  Berakhir dalam: 02:30  │
└─────────────────────────┘
```

---

## Database Schema (tickets table extension)

```sql
-- Sales Timing
sale_start_time TIMESTAMPTZ NULL,
sale_end_time   TIMESTAMPTZ NULL,

-- Flash Sale
is_flash_sale    BOOL NOT NULL DEFAULT false,
flash_sale_price NUMERIC(12, 2) NULL,
flash_start_time TIMESTAMPTZ NULL,
flash_end_time   TIMESTAMPTZ NULL,

-- Stock Alert
low_stock_threshold INT NULL,   -- Default: 20%
show_stock_alert    BOOL NOT NULL DEFAULT true,

-- FOMO / Urgency
show_countdown  BOOL NOT NULL DEFAULT true,
countdown_end   TIMESTAMPTZ NULL,
```