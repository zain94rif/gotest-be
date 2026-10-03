# IMS Finance — Loan Management Backend

Backend REST API untuk sistem peminjaman/kredit kendaraan, dibangun menggunakan **Go + Gin + PostgreSQL**.

---

## Requirement

- Go 1.22+
- Docker & Docker Compose
- PostgreSQL 16 (via Docker)

---

## Tech Stack

| Komponen | Library |
|----------|---------|
| Web Framework | [Gin](https://github.com/gin-gonic/gin) |
| Database Driver | [pgx v5](https://github.com/jackc/pgx) |
| Env Loading | [godotenv](https://github.com/joho/godotenv) |
| Database | PostgreSQL 16 |

---

## Cara Menjalankan

### 1. Clone & setup env

```bash
cp .env.example .env
# Edit .env sesuai konfigurasi lokal Anda
```

### 2. Jalankan PostgreSQL

```bash
docker compose up -d
```

### 3. Jalankan Migration (jika database kosong)

```bash
psql -h localhost -U postgres -d ims_finance -f migrations/001_init.sql
```

> ⚠️ Lewati langkah ini jika database sudah tersedia.

### 4. Jalankan Aplikasi

```bash
go run cmd/api/main.go
```

Server berjalan di `http://localhost:9531`

---

## API Endpoints

### Business Logic
| Method | Endpoint | Deskripsi |
|--------|----------|-----------|
| POST | `/api/v1/loans/calculate` | Hitung kredit (tanpa simpan) |

### Clients
| Method | Endpoint | Deskripsi |
|--------|----------|-----------|
| GET | `/api/v1/clients` | List semua client |
| GET | `/api/v1/clients/:id` | Get client by ID |
| POST | `/api/v1/clients` | Buat client baru |
| PUT | `/api/v1/clients/:id` | Update nama client |
| DELETE | `/api/v1/clients/:id` | Hapus client |

### Contracts
| Method | Endpoint | Deskripsi |
|--------|----------|-----------|
| GET | `/api/v1/contracts` | List semua kontrak |
| GET | `/api/v1/contracts/:contractNo` | Detail kontrak |
| POST | `/api/v1/contracts` | Buat kontrak kredit + generate angsuran |
| PUT | `/api/v1/contracts/:contractNo` | Update kontrak (client_id / tanggal) |
| DELETE | `/api/v1/contracts/:contractNo` | Hapus kontrak |
| GET | `/api/v1/contracts/:contractNo/installments?to=N` | Jadwal angsuran (opsional: filter sampai ke-N) |
| GET | `/api/v1/contracts/:contractNo/overdue?date=` | Overdue + denda per angsuran |
| GET | `/api/v1/contracts/:contractNo/overdue-summary?date=` | **Total** angsuran yang sudah jatuh tempo |
| GET | `/api/v1/contracts/:contractNo/penalties?date=` | **Detail denda** per angsuran yang belum dibayar |

### Installments
| Method | Endpoint | Deskripsi |
|--------|----------|-----------|
| GET | `/api/v1/installments/:id` | Get angsuran by ID |
| PUT | `/api/v1/installments/:id` | Update due_date & amount angsuran |
| DELETE | `/api/v1/installments/:id` | Hapus angsuran |

### Payments
| Method | Endpoint | Deskripsi |
|--------|----------|-----------|
| GET | `/api/v1/payments` | List semua payment |
| GET | `/api/v1/payments/:id` | Get payment by ID |
| POST | `/api/v1/payments` | Catat pembayaran angsuran |
| PUT | `/api/v1/payments/:id` | Update payment |
| DELETE | `/api/v1/payments/:id` | Hapus payment |

### Reports
| Method | Endpoint | Deskripsi |
|--------|----------|-----------|
| GET | `/api/v1/reports/client-overdue?client_code=&date=` | Total overdue seluruh kontrak satu client |

---

## Contoh Request & Response

### POST /api/v1/loans/calculate

**Request:**
```json
{
    "otr": 240000000,
    "dp_percentage": 20,
    "tenor_months": 18
}
```

**Response:**
```json
{
    "otr": 240000000,
    "dp": 48000000,
    "principal": 192000000,
    "tenor_months": 18,
    "interest_rate": 0.14,
    "interest_amount": 40320000,
    "total_amount": 232320000,
    "monthly_installment": 12907000
}
```

---

### POST /api/v1/contracts

**Request:**
```json
{
    "client_id": 1,
    "contract_no": "AGR00001",
    "otr": 240000000,
    "dp_percentage": 20,
    "tenor_months": 18,
    "contract_date": "2024-01-01",
    "first_due_date": "2024-01-25"
}
```

**Response:** `201 Created` — data kontrak lengkap

---

### GET /api/v1/contracts/AGR00001/installments

Filter opsional: `?to=12` untuk menampilkan angsuran 1 s/d 12 saja.

**Response:**
```json
{
    "contract_no": "AGR00001",
    "installments": [
        { "installment_no": 1, "amount": 12907000, "due_date": "2024-01-25" },
        { "installment_no": 2, "amount": 12907000, "due_date": "2024-02-25" }
    ]
}
```

---

### POST /api/v1/payments

**Request:**
```json
{
    "installment_id": 1,
    "payment_date": "2024-01-25",
    "amount": 12907000,
    "payment_reference": "PAY-000001"
}
```

**Response:** `201 Created`

---

### GET /api/v1/contracts/AGR00001/overdue?date=2024-08-14

**Response:**
```json
{
    "contract_no": "AGR00001",
    "as_of": "2024-08-14",
    "total_overdue": 25814000,
    "total_penalty": 903490,
    "installments": [
        {
            "installment_no": 6,
            "due_date": "2024-06-25",
            "amount": 12907000,
            "paid_amount": 0,
            "remaining_amount": 12907000,
            "days_late": 50,
            "penalty": 645350
        },
        {
            "installment_no": 7,
            "due_date": "2024-07-25",
            "amount": 12907000,
            "paid_amount": 0,
            "remaining_amount": 12907000,
            "days_late": 20,
            "penalty": 258140
        }
    ]
}
```

---

### GET /api/v1/contracts/AGR00001/overdue-summary?date=2024-08-25

SUM seluruh angsuran yang due_date <= tanggal input (termasuk yang sudah dibayar).
Jika `date` tidak diisi, default ke `2024-08-14`.

**Response:**
```json
{
    "contract_no": "AGR00001",
    "client_name": "SUGUS",
    "total_angsuran_jatuh_tempo": 103256000
}
```

---

### GET /api/v1/contracts/AGR00001/penalties?date=2024-08-14

Angsuran yang due_date <= tanggal input DAN belum ada pembayaran sama sekali, beserta kalkulasi denda.
Jika `date` tidak diisi, default ke `2024-08-14`.

**Response:**
```json
[
    {
        "contract_no": "AGR00001",
        "client_name": "SUGUS",
        "installment_no": 6,
        "hari_keterlambatan": 50,
        "total_denda": 645350
    },
    {
        "contract_no": "AGR00001",
        "client_name": "SUGUS",
        "installment_no": 7,
        "hari_keterlambatan": 20,
        "total_denda": 258140
    }
]
```

---

### GET /api/v1/reports/client-overdue?client_code=SUGUS&date=2024-08-14

Total agregat overdue seluruh kontrak milik satu client.

**Response:**
```json
{
    "client_code": "SUGUS",
    "name": "Pak Sugus",
    "as_of": "2024-08-14",
    "overdue_count": 2,
    "total_overdue": 25814000
}
```

---

## Cara Menjalankan Test

### Unit Test (tanpa DB)
```bash
go test ./internal/service/... -v
```

Test yang tersedia:
- `TestDetermineInterestRate` — validasi bunga berdasarkan tenor (12%, 14%, 16.5%)
- `TestCalculateLoan` — validasi lengkap kredit Pak Sugus (OTR 240jt, DP 20%, tenor 18)
- `TestCalculatePenalty` — validasi formula denda (0.1%/hari)

### Integration Test (butuh DB)
```bash
go test ./tests/... -v
```

Test yang tersedia:
- `TestCalculateLoanEndpoint` — POST /loans/calculate
- `TestGetInstallmentsEndpoint` — GET installments + filter `?to=12`
- `TestGetOverdueEndpoint` — GET overdue + kalkulasi denda
- `TestClientOverdueSummaryReport` — GET /reports/client-overdue
- `TestGetOverdueSummaryEndpoint` — GET /overdue-summary
- `TestGetPenaltiesEndpoint` — GET /penalties

### Jalankan Semua Test
```bash
go test ./... -v
```

---

## Struktur Project

```
.
├── cmd/
│   └── api/
│       └── main.go              # Entry point
├── internal/
│   ├── database/
│   │   └── postgres.go          # Connection pool pgxpool
│   ├── handler/                 # HTTP handler (Gin)
│   │   ├── loan_handler.go
│   │   ├── client_handler.go
│   │   ├── contract_handler.go
│   │   ├── installment_handler.go
│   │   └── payment_handler.go
│   ├── service/                 # Business logic
│   │   ├── loan_service.go
│   │   ├── loan_service_test.go
│   │   ├── client_service.go
│   │   ├── contract_service.go
│   │   ├── installment_service.go
│   │   └── payment_service.go
│   ├── repository/              # SQL queries
│   │   ├── client_repository.go
│   │   ├── contract_repository.go
│   │   ├── installment_repository.go
│   │   └── payment_repository.go
│   └── model/                   # Struct data & request/response types
│       ├── client.go
│       ├── contract.go
│       ├── installment.go
│       └── payment.go
├── tests/
│   └── api_integration_test.go  # Integration tests (butuh DB)
├── migrations/
│   └── 001_init.sql             # DDL schema + tabel + index
├── docker-compose.yml
├── .env.example
├── go.mod
└── README.md
```
