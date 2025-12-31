# 📘 ZTA-CS API README (Zero Trust Customer Service)


---

## 1. Gambaran Umum Project

**Nama Project:** ZTA-CS (Zero Trust Customer Service)
**Tipe:** Ticketing & Chat System dengan Zero Trust Architecture
**Tujuan:** Menyediakan layanan Customer Service yang **aman**, **terkontrol**, dan **ter-audit** untuk aksi sensitif (contoh: reset password).

### Prinsip Keamanan Utama

* **Zero Trust** → CS *tidak pernah* dipercaya secara default.
* **Just-In-Time Access (JIT)** → Akses sensitif hanya aktif **sementara & berbasis verifikasi user**.
* **Least Privilege** → CS hanya bisa melakukan aksi sesuai konteks tiket.
* **Full Audit Trail** → Semua aksi tercatat dan immutable.

---

## 2. Role & Hak Akses

| Role    | Deskripsi         | Hak Utama                                            |
| ------- | ----------------- | ---------------------------------------------------- |
| USER    | Pengguna aplikasi | Buat tiket, chat, verifikasi identitas               |
| CS      | Customer Support  | Klaim tiket, chat, trigger verifikasi, aksi sensitif |
| AUDITOR | Pengawas          | Baca audit log (read-only)                           |

---

## 3. Standar Teknis

* **Base URL (Dev):** `http://localhost:8080`
* **Format Data:** JSON
* **Authentication:** JWT Bearer Token
* **Header Wajib (Protected API):**

  ```http
  Authorization: Bearer <token>
  ```
* **Date Format:** ISO 8601 (UTC)

  ```json
  "2025-12-31T15:04:05Z"
  ```

---

## 4. Data Dictionary (Enum & Status)

### Role

* `USER`
* `CS`
* `AUDITOR`

### Ticket Status

* `OPEN` → Tiket baru, belum di-claim
* `IN_PROGRESS` → Sedang ditangani CS
* `CLOSED` → Tiket selesai & terkunci

### Verification Status

* `PENDING` → Menunggu jawaban user
* `PASSED` → Verifikasi sukses (JIT aktif)
* `FAILED` → Jawaban salah
* `EXPIRED` → Sesi verifikasi kadaluarsa

---

## 5. Authentication Module (Public)

### Login

Digunakan oleh semua role.

**Endpoint**

```
POST /login
```

**Request Body**

```json
{
  "email": "user@example.com",
  "password": "secretpassword"
}
```

**Response 200**

```json
{
  "token": "jwt_token_string",
  "role": "CS"
}
```

**Response 401**

```json
{
  "error": "Invalid email or password"
}
```

---

## 6. Verification Module (Public – Via Email Link)

### Get Verification Questions

**Endpoint**

```
GET /verify/:token
```

**Response 200**

```json
{
  "questions": [
    { "id": 1, "category": "STATIC", "question": "Apa nama ibu kandung anda?" },
    { "id": 5, "category": "USAGE", "question": "Berapa transaksi terakhir anda?" }
  ]
}
```

---

### Submit Verification Answers

**Endpoint**

```
POST /verify/:token
```

**Request Body**

```json
{
  "answers": {
    "1": "Siti Aminah",
    "5": "50000"
  }
}
```

**Response 200 (PASSED)**

```json
{
  "status": "PASSED",
  "message": "Identity verified. Support agent has been notified."
}
```

**Response 403 (FAILED)**

```json
{
  "status": "FAILED",
  "message": "Verification failed. Access denied."
}
```

---

## 7. User Dashboard API (Role: USER)

### Create Ticket

```
POST /api/user/tickets
```

```json
{ "subject": "Saya lupa password akun saya" }
```

---

### Chat (User)

* **Send:** `POST /api/user/tickets/:id/chat`
* **History:** `GET /api/user/tickets/:id/chat`

---

## 8. CS Workspace API (Role: CS)

### Get Open Tickets

```
GET /api/cs/tickets/open
```

---

### Claim Ticket

```
POST /api/cs/tickets/:id/claim
```

**Rule:** CS hanya boleh punya **1 tiket IN_PROGRESS**.

---

### Start Verification (Zero Trust Trigger)

```
POST /api/cs/tickets/:id/start-verification
```

**Catatan FE:**

* Disable tombol jika `RiskScore >= 80`

---

### Reset Password (JIT Required)

```
POST /api/cs/tickets/:id/reset-password
```

**Syarat:**

* Verification Status = `PASSED`
* JIT Token masih aktif

---

## 9. Auditor API (Role: AUDITOR)

### Get Audit Logs

```
GET /api/auditor/logs
```

Audit bersifat **immutable** dan **anonim (hash)**.

---

## 10. End-to-End Flow (Ringkas)

1. User buat tiket
2. CS claim tiket
3. CS trigger verifikasi
4. User lolos verifikasi
5. CS dapat JIT access
6. CS eksekusi aksi sensitif
7. Semua tercatat di audit log

---

## 11. Catatan Penting untuk Frontend

* Jangan hardcode role → selalu pakai value dari JWT
* Tombol sensitif **harus state-aware** (disabled/enabled)
* Anggap semua error `403` sebagai **policy violation**, bukan bug

---

Dokumen ini adalah **single source of truth** untuk integrasi API ZTA-CS.
Jika FE mengikuti README ini, **tidak ada logic backend yang ambigu**.
