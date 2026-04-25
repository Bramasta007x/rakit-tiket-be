# Rakit-Checkin Gate System API Documentation

---

## Overview

Rakit-Checkin Gate System adalah modul untuk generate ticket fisik QR dan scan check-in/check-out di gate event.

### Base URL
```
http://localhost:8001/api/v1
```

### Authentication
- **Public**: Tidak memerlukan authentication
- **Admin**: Memerlukan JWT token dengan role `ADMIN`

---

## Endpoints

### 1. Scan Ticket (Public)

Scan ticket fisik untuk check-in atau check-out.

**Endpoint:** `POST /gate/scan`

**Request:**
```json
{
    "qr_code": "TKT2026-SILVER-001",
    "gate_name": "GATE-A"
}
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| qr_code | string | Yes | Kode QR ticket fisik |
| gate_name | string | No | Nama gate (misal: GATE-A, GATE-B) |

**Response Success (200):**
```json
{
    "success": true,
    "data": {
        "success": true,
        "action": "CHECK_IN",
        "qr_code": "TKT2026-SILVER-001",
        "ticket_type": "SILVER",
        "scan_count": 1,
        "message": "CHECK_IN berhasil"
    }
}
```

**Response Error / Duplicate (400):**
```json
{
    "success": false,
    "data": {
        "success": false,
        "action": "DUPLICATE",
        "qr_code": "TKT2026-SILVER-001",
        "ticket_type": "SILVER",
        "scan_count": 1,
        "message": "Tiket sudah di-scan"
    }
}
```

**Response Invalid (400):**
```json
{
    "success": false,
    "data": {
        "success": false,
        "action": "INVALID",
        "qr_code": "TKT2026-SILVER-999",
        "message": "Tiket tidak ditemukan"
    }
}
```

---

### 2. Create Gate Config (Admin)

Buat atau update konfigurasi gate untuk event.

**Endpoint:** `POST /admin/gate/config`

**Headers:**
```
Authorization: Bearer <JWT_TOKEN>
```

**Request:**
```json
{
    "event_id": "uuid-event-123",
    "mode": "CHECK_IN_OUT",
    "max_scan_per_ticket": 2,
    "max_scan_by_type": {
        "GOLD": 3,
        "SILVER": 1,
        "VIP": 10
    },
    "is_active": true
}
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| event_id | string | Yes | UUID event |
| mode | string | No | `CHECK_IN` atau `CHECK_IN_OUT` (default: `CHECK_IN`) |
| max_scan_per_ticket | int | No | Batas scan per ticket (1-10, default: 1) |
| max_scan_by_type | object | No | Override max scan per kategori ticket |
| is_active | bool | No | Status aktif (default: true) |

**Response Success (200):**
```json
{
    "success": true,
    "data": {
        "id": "uuid-config-123",
        "event_id": "uuid-event-123",
        "mode": "CHECK_IN_OUT",
        "max_scan_per_ticket": 2,
        "max_scan_by_type": {
            "GOLD": 3,
            "SILVER": 1,
            "VIP": 10
        },
        "is_active": true
    }
}
```

---

### 3. Get Gate Config (Admin)

Ambil konfigurasi gate untuk event.

**Endpoint:** `GET /admin/gate/config/:event_id`

**Headers:**
```
Authorization: Bearer <JWT_TOKEN>
```

**Response Success (200):**
```json
{
    "success": true,
    "data": {
        "id": "uuid-config-123",
        "event_id": "uuid-event-123",
        "mode": "CHECK_IN_OUT",
        "max_scan_per_ticket": 2,
        "max_scan_by_type": {
            "GOLD": 3,
            "SILVER": 1
        },
        "is_active": true
    }
}
```

**Response Not Found (404):**
```json
{
    "success": false,
    "message": "konfigurasi gate tidak ditemukan"
}
```

---

Generate Tiket se pengen nya admin, di batch per 100
### 4. Generate Physical Tickets (Admin)

Generate ticket fisik QR dengan jumlah bebas (tidak melebihi sold_qty).

**Endpoint:** `POST /admin/gate/generate-qr`

**Headers:**
```
Authorization: Bearer <JWT_TOKEN>
```

**Request:**
```json
{
    "event_id": "uuid-event-123",
    "ticket_types": ["SILVER", "GOLD"],
    "qty": {
        "SILVER": 50,
        "GOLD": 20
    }
}
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| event_id | string | Yes | UUID event |
| ticket_types | array | No | Filter kategori ticket (kosong = semua) |
| qty | object | No | Jumlah per kategori. Jika kosong, gunakan sold_qty. Tidak boleh melebihi sold_qty. |

**Catatan:**
- Jika `qty` tidak diisi, gunakan `sold_qty` dari ticket master
- Tidak boleh melebihi `sold_qty` dari ticket master
- Jika sudah ada ticket yang digenerate sebelumnya, akan generate ulang (tambah) dengan sisa quota

**Response Success (200):**
```json
{
    "success": true,
    "data": {
        "SILVER": {
            "generated": 50,
            "start_code": "TKT2026-SILVER-001",
            "end_code": "TKT2026-SILVER-050"
        },
        "GOLD": {
            "generated": 20,
            "start_code": "TKT2026-GOLD-001",
            "end_code": "TKT2026-GOLD-020"
        }
    }
}
```

**Response Error (400):**
```json
{
    "success": false,
    "message": "jumlah GOLD (100) melebihi sold_qty (50)"
}
```

---

### 5. Get Physical Tickets (Admin)

Ambil list ticket fisik untuk event.

**Endpoint:** `GET /admin/gate/qr/:event_id`

**Headers:**
```
Authorization: Bearer <JWT_TOKEN>
```

**Query Parameters:**
| Parameter | Type | Description |
|-----------|------|-------------|
| ticket_types | string | Filter kategori ticket |

**Response Success (200):**
```json
{
    "success": true,
    "data": [
        {
            "id": "uuid-ticket-1",
            "qr_code": "TKT2026-SILVER-001",
            "ticket_type": "SILVER",
            "registrant_id": "uuid-reg-1",
            "attendee_id": null,
            "status": "ACTIVE",
            "scan_count": 0
        },
        {
            "id": "uuid-ticket-2",
            "qr_code": "TKT2026-SILVER-002",
            "ticket_type": "SILVER",
            "registrant_id": "uuid-reg-1",
            "attendee_id": "uuid-att-1",
            "status": "CHECKED_IN",
            "scan_count": 1
        }
    ],
    "count": 2
}
```

---

### 6. Get Gate Stats (Admin)

Ambil statistik check-in untuk event.

**Endpoint:** `GET /admin/gate/stats/:event_id`

**Headers:**
```
Authorization: Bearer <JWT_TOKEN>
```

**Response Success (200):**
```json
{
    "success": true,
    "data": {
        "total_physical_tickets": 150,
        "checked_in": 120,
        "checked_out": 80,
        "active_now": 40,
        "by_type": {
            "SILVER": {
                "total": 100,
                "checked_in": 90,
                "checked_out": 60,
                "active_now": 30
            },
            "GOLD": {
                "total": 50,
                "checked_in": 30,
                "checked_out": 20,
                "active_now": 10
            }
        }
    }
}
```

---

### 7. Get Gate Logs (Admin)

Ambil audit trail scan ticket.

**Endpoint:** `GET /admin/gate/logs/:event_id`

**Headers:**
```
Authorization: Bearer <JWT_TOKEN>
```

**Response Success (200):**
```json
{
    "success": true,
    "data": [
        {
            "id": "uuid-log-1",
            "event_id": "uuid-event-123",
            "physical_ticket_id": "uuid-ticket-1",
            "action": "CHECK_IN",
            "success": true,
            "message": "CHECK_IN berhasil",
            "gate_name": "GATE-A",
            "ticket_type": "SILVER",
            "scan_sequence": 1,
            "created_at": "2026-04-23 10:30:00"
        }
    ],
    "count": 1
}
```

---

## Mode Configuration

### Mode: CHECK_IN

- **Max Scan**: Konfigurasi (1-10)
- **Use Case**: Festival tanpa re-entry, seminar

**Config Example:**
```json
{
    "event_id": "uuid-event-123",
    "mode": "CHECK_IN",
    "max_scan_per_ticket": 1,
    "max_scan_by_type": {
        "GOLD": 3,
        "SILVER": 1
    },
    "is_active": true
}
```

- `max_scan_per_ticket: 1` = Sekali masuk saja
- `max_scan_per_ticket: 3` = Boleh masuk 3x

### Mode: CHECK_IN_OUT

- **Max Scan**: Fixed 2x (1x Check-in, 1x Check-out)
- **Use Case**: Ruang terbatas, buffet, seminar with break

**Config Example:**
```json
{
    "event_id": "uuid-event-123",
    "mode": "CHECK_IN_OUT",
    "is_active": true
}
```

---

## QR Code Format

Format: `TKT{YYYY}-{TYPE}-{SEQUENCE:3DIGIT}`

| Contoh | Meaning |
|--------|---------|
| `TKT2026-SILVER-001` | Tahun 2026, Silver, urutan 1 |
| `TKT2026-GOLD-001` | Tahun 2026, Gold, urutan 1 |
| `TKT2026-VIP-123` | Tahun 2026, VIP, urutan 123 |

---

## Error Codes

| Code | Message |
|------|---------|
| 400 | Invalid request |
| 401 | Unauthorized |
| 403 | Forbidden ( bukan admin) |
| 404 | Resource not found |
| 500 | Internal server error |

---

## Notes

1. **Generate QR** dijalankan setelah pembayaran berhasil (PAID)
2. **Scan ticket** memerlukan Gate Config sudah diaktifkan
3. **Audit trail** disimpan di `gate_logs` untuk traceability
4. **Mode CHECK_IN** dapat dikonfigurasi max scan 1-10
5. **Mode CHECK_IN_OUT** fixed 2 scan (in + out)