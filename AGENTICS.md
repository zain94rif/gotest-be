# AGENTICS.md

## Project: IMS Finance - Loan Management Backend

## 1. Tujuan Project

Buat sebuah **Backend REST API menggunakan Golang** untuk sistem peminjaman/kredit kendaraan sederhana.

Backend harus dapat:

1. Menghitung kredit berdasarkan:
   - OTR
   - DP
   - Jangka waktu/tenor
   - Bunga berdasarkan tenor
2. Membuat kontrak kredit.
3. Membuat jadwal angsuran.
4. Mencatat pembayaran angsuran.
5. Menampilkan angsuran yang sudah jatuh tempo.
6. Menghitung keterlambatan.
7. Menghitung denda keterlambatan.

Project ini dibuat untuk **technical test**, sehingga prioritaskan:
- kode sederhana
- mudah dibaca
- mudah dijalankan
- struktur BE yang jelas
- tidak over-engineering

---

# 2. Tech Stack

Gunakan:

- Go
- Gin
- PostgreSQL
- pgx
- REST API
- SQL migration
- Docker Compose untuk PostgreSQL

Jangan menggunakan:
- ORM
- framework tambahan yang tidak diperlukan
- microservices
- message broker
- Redis
- Kubernetes

Gunakan SQL langsung melalui `pgx`.

---

# 3. Struktur Project

Gunakan struktur berikut:

```text
.
├── cmd/
│   └── api/
│       └── main.go
│
├── internal/
│   ├── handler/
│   │   ├── loan_handler.go
│   │   ├── contract_handler.go
│   │   ├── installment_handler.go
│   │   └── payment_handler.go
│   │
│   ├── service/
│   │   ├── loan_service.go
│   │   ├── contract_service.go
│   │   ├── installment_service.go
│   │   └── payment_service.go
│   │
│   ├── repository/
│   │   ├── contract_repository.go
│   │   ├── installment_repository.go
│   │   └── payment_repository.go
│   │
│   ├── model/
│   │   ├── client.go
│   │   ├── contract.go
│   │   ├── installment.go
│   │   └── payment.go
│   │
│   └── database/
│       └── postgres.go
│
├── migrations/
│   └── 001_init.sql
│
├── .env.example
├── docker-compose.yml
├── go.mod
├── go.sum
└── README.md
```

Jangan membuat folder tambahan kecuali memang diperlukan.

---

# 4. Database

Gunakan PostgreSQL.

Gunakan schema:

```text
ujicoba
```

Database memiliki 4 tabel:

```text
clients
contracts
installments
payments
```

Relasi:

```text
clients
   |
   | 1:N
   v
contracts
   |
   | 1:N
   v
installments
   |
   | 1:N
   v
payments
```

---

# 5. Table: clients

Menyimpan data customer.

```sql
CREATE TABLE ujicoba.clients (
    id BIGSERIAL PRIMARY KEY,
    client_code VARCHAR(50) NOT NULL UNIQUE,
    name VARCHAR(150) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

---

# 6. Table: contracts

Menyimpan data kontrak kredit.

```sql
CREATE TABLE ujicoba.contracts (
    id BIGSERIAL PRIMARY KEY,

    contract_no VARCHAR(50) NOT NULL UNIQUE,

    client_id BIGINT NOT NULL
        REFERENCES ujicoba.clients(id),

    otr NUMERIC(18,2) NOT NULL,

    dp_percentage NUMERIC(5,2) NOT NULL,

    dp_amount NUMERIC(18,2) NOT NULL,

    principal_amount NUMERIC(18,2) NOT NULL,

    tenor_months INTEGER NOT NULL,

    interest_rate NUMERIC(7,4) NOT NULL,

    interest_amount NUMERIC(18,2) NOT NULL,

    total_amount NUMERIC(18,2) NOT NULL,

    monthly_installment NUMERIC(18,2) NOT NULL,

    contract_date DATE NOT NULL,

    first_due_date DATE NOT NULL,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CHECK (otr > 0),
    CHECK (dp_percentage >= 0 AND dp_percentage <= 100),
    CHECK (tenor_months > 0),
    CHECK (interest_rate >= 0)
);
```

---

# 7. Table: installments

Menyimpan jadwal angsuran.

```sql
CREATE TABLE ujicoba.installments (
    id BIGSERIAL PRIMARY KEY,

    contract_id BIGINT NOT NULL
        REFERENCES ujicoba.contracts(id),

    installment_no INTEGER NOT NULL,

    due_date DATE NOT NULL,

    amount NUMERIC(18,2) NOT NULL,

    principal_amount NUMERIC(18,2),

    interest_amount NUMERIC(18,2),

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CHECK (installment_no > 0),
    CHECK (amount > 0),

    UNIQUE (contract_id, installment_no)
);
```

---

# 8. Table: payments

Menyimpan pembayaran angsuran.

```sql
CREATE TABLE ujicoba.payments (
    id BIGSERIAL PRIMARY KEY,

    installment_id BIGINT NOT NULL
        REFERENCES ujicoba.installments(id),

    payment_date DATE NOT NULL,

    amount NUMERIC(18,2) NOT NULL,

    payment_reference VARCHAR(100),

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CHECK (amount > 0)
);
```

---

# 9. Index

Buat index berikut:

```sql
CREATE INDEX idx_contracts_client_id
ON contracts(client_id);

CREATE INDEX idx_installments_contract_due_date
ON installments(contract_id, due_date);

CREATE INDEX idx_payments_installment_id
ON payments(installment_id);
```

---

# 10. Business Rules

## 10.1 DP

DP dihitung berdasarkan persentase OTR:

```text
DP = OTR × DP Percentage
```

Contoh:

```text
OTR = 240.000.000
DP = 20%

DP = 240.000.000 × 20%
   = 48.000.000
```

---

# 11. Pokok Utang

```text
Pokok Utang = OTR - DP
```

Contoh:

```text
240.000.000 - 48.000.000
= 192.000.000
```

---

# 12. Bunga

Gunakan aturan:

```text
Tenor <= 12 bulan
    bunga = 12%

Tenor > 12 dan <= 24 bulan
    bunga = 14%

Tenor > 24 bulan
    bunga = 16.5%
```

Untuk bunga total gunakan:

```text
Total Bunga =
Pokok Utang × Bunga × (Tenor / 12)
```

Contoh tenor 18 bulan:

```text
192.000.000 × 14% × 1,5
= 40.320.000
```

---

# 13. Total Kredit

```text
Total Kredit =
Pokok Utang + Total Bunga
```

Contoh:

```text
192.000.000 + 40.320.000
= 232.320.000
```

---

# 14. Angsuran Per Bulan

```text
Angsuran Per Bulan =
Total Kredit / Tenor
```

Contoh:

```text
232.320.000 / 18
= 12.906.666,67
```

Untuk menghasilkan angka seperti contoh soal:

```text
Rp12.907.000
```

gunakan pembulatan ke ribuan.

Jangan menggunakan `float64` untuk menyimpan nilai uang pada database.

Database menggunakan:

```text
NUMERIC
```

Di Go, gunakan pendekatan yang aman untuk monetary calculation. Jika implementasi sederhana menggunakan integer rupiah, gunakan `int64`.

---

# 15. Jadwal Angsuran

Ketika contract dibuat, generate installment sebanyak:

```text
tenor_months
```

Contoh:

```text
tenor = 18
```

maka:

```text
installment 1
installment 2
...
installment 18
```

Tanggal jatuh tempo berdasarkan `first_due_date`.

Contoh:

```text
first_due_date = 2024-01-25

1  = 2024-01-25
2  = 2024-02-25
3  = 2024-03-25
...
18 = 2025-06-25
```

Gunakan `time.AddDate()` di Go.

---

# 16. Pembayaran

Pembayaran dicatat pada tabel:

```text
payments
```

Jangan menyimpan status:

```text
PAID
UNPAID
```

di tabel installments.

Status pembayaran harus dihitung berdasarkan total payment.

Contoh:

```text
installment amount = 12.907.000

payment total = 12.907.000

status = PAID
```

Jika:

```text
payment total = 0
```

maka:

```text
status = UNPAID
```

Jika:

```text
payment total < installment amount
```

maka:

```text
status = PARTIAL
```

---

# 17. Overdue

Sebuah installment dianggap overdue jika:

```text
due_date < as_of_date
AND
total_payment < installment_amount
```

Gunakan tanggal yang diberikan API.

Contoh:

```text
as_of_date = 2024-08-14
```

Jangan menggunakan `NOW()` untuk business calculation jika user memberikan tanggal perhitungan.

---

# 18. Denda

Denda:

```text
0,1% per hari
```

atau:

```text
0.001
```

Formula:

```text
Denda =
Sisa Angsuran × 0.001 × Hari Keterlambatan
```

Untuk technical test, gunakan nilai angsuran sebagai dasar denda sesuai soal.

Contoh:

```text
Angsuran = 12.907.000
Jatuh tempo = 2024-06-25
Tanggal perhitungan = 2024-08-14

Hari terlambat = 50

Denda =
12.907.000 × 0.001 × 50

= 645.350
```

---

# 19. API

## Calculate Loan

```http
POST /api/v1/loans/calculate
```

Request:

```json
{
    "otr": 240000000,
    "dp_percentage": 20,
    "tenor_months": 18
}
```

Response:

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

# 20. Create Contract

```http
POST /api/v1/contracts
```

Request:

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

Backend harus:

1. Validasi request.
2. Hitung DP.
3. Hitung pokok utang.
4. Tentukan bunga.
5. Hitung total bunga.
6. Hitung total kredit.
7. Hitung angsuran.
8. Insert contract.
9. Generate installment sesuai tenor.
10. Commit transaction.

Gunakan PostgreSQL transaction agar contract dan installments tidak dibuat setengah-setengah.

---

# 21. Get Installments

```http
GET /api/v1/contracts/{contractNo}/installments
```

Contoh:

```http
GET /api/v1/contracts/AGR00001/installments
```

Response harus menampilkan:

```json
{
    "contract_no": "AGR00001",
    "installments": [
        {
            "installment_no": 1,
            "amount": 12907000,
            "due_date": "2024-01-25"
        }
    ]
}
```

---

# 22. Create Payment

```http
POST /api/v1/payments
```

Request:

```json
{
    "installment_id": 1,
    "payment_date": "2024-01-25",
    "amount": 12907000,
    "payment_reference": "PAY-000001"
}
```

Backend harus memvalidasi bahwa installment memang ada.

---

# 23. Get Overdue

```http
GET /api/v1/contracts/{contractNo}/overdue?date=2024-08-14
```

Contoh:

```http
GET /api/v1/contracts/AGR00001/overdue?date=2024-08-14
```

Response:

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

# 24. SQL Query Overdue

Gunakan PostgreSQL untuk mengambil data overdue.

Konsep query:

```sql
SELECT
    i.id,
    c.contract_no,
    i.installment_no,
    i.amount,
    i.due_date,
    COALESCE(SUM(p.amount), 0) AS paid_amount
FROM ujicoba.installments i
JOIN ujicoba.contracts c
    ON c.id = i.contract_id
LEFT JOIN ujicoba.payments p
    ON p.installment_id = i.id
WHERE c.contract_no = $1
  AND i.due_date < $2
GROUP BY
    i.id,
    c.contract_no,
    i.installment_no,
    i.amount,
    i.due_date
HAVING COALESCE(SUM(p.amount), 0) < i.amount
ORDER BY i.installment_no;
```

Perhitungan:

```text
remaining_amount = amount - paid_amount

days_late = as_of_date - due_date

penalty = amount × 0.001 × days_late
```

---

# 25. Transaction

Create contract harus menggunakan transaction:

```text
BEGIN
    INSERT contract
    INSERT installments
COMMIT
```

Jika salah satu gagal:

```text
ROLLBACK
```

Payment juga harus menggunakan transaction jika membutuhkan lebih dari satu query.

---

# 26. Validation

Validasi minimal:

### OTR

```text
OTR > 0
```

### DP

```text
0 <= DP <= 100
```

### Tenor

```text
Tenor > 0
```

### Contract number

```text
contract_no wajib unik
```

### Payment

```text
amount > 0
```

Jangan menerima input invalid secara diam-diam.

Gunakan HTTP status yang sesuai:

```text
200 OK
201 Created
400 Bad Request
404 Not Found
409 Conflict
500 Internal Server Error
```

---

# 27. Error Response

Gunakan format sederhana:

```json
{
    "error": "contract not found"
}
```

Jangan mengembalikan stack trace kepada client.

Error detail boleh ditulis ke log server.

---

# 28. Environment

Buat `.env.example`:

```env
APP_PORT=9531

DATABASE_URL=postgres://postgres:postgres@localhost:5432/ims_finance?sslmode=disable
```

Jangan commit `.env`.

---

# 29. Docker Compose

Sediakan PostgreSQL menggunakan Docker Compose.

Contoh:

```yaml
services:
  postgres:
    image: postgres:16
    container_name: ims-finance-postgres
    environment:
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: postgres
      POSTGRES_DB: ims_finance
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data

volumes:
  postgres_data:
```

---

# 30. Migration

`migrations/001_init.sql` harus berisi:

1. CREATE SCHEMA `ujicoba`
2. CREATE TABLE `clients`
3. CREATE TABLE `contracts`
4. CREATE TABLE `installments`
5. CREATE TABLE `payments`
6. CREATE INDEX

Migration harus dapat dijalankan pada database kosong.

---

# 31. Seed Data

Sediakan seed data untuk Pak Sugus.

Client:

```text
client_code = SUGUS
name = Sugus
```

Contract:

```text
contract_no = AGR00001
client = Sugus
OTR = 240000000
DP = 20%
tenor = 18 bulan
contract_date = 2024-01-01
first_due_date = 2024-01-25
```

Setelah contract dibuat, harus tersedia 18 installment.

Pembayaran:

```text
Angsuran 1 → dibayar
Angsuran 2 → dibayar
Angsuran 3 → dibayar
Angsuran 4 → dibayar
Angsuran 5 → dibayar
```

Angsuran 6 dan seterusnya belum dibayar.

---

# 32. Expected Test Case

Gunakan data Pak Sugus untuk memastikan hasil sesuai soal.

Input:

```text
OTR = 240.000.000
DP = 20%
Tenor = 18 bulan
```

Expected:

```text
DP                  = 48.000.000
Principal           = 192.000.000
Interest Rate       = 14%
Interest Amount     = 40.320.000
Total Amount        = 232.320.000
Monthly Installment = 12.907.000
```

Tanggal:

```text
As Of = 2024-08-14
```

Pembayaran sampai:

```text
2024-05-25
```

Expected overdue:

```text
Installment 6
Due Date      = 2024-06-25
Days Late     = 50
Penalty       = 645.350

Installment 7
Due Date      = 2024-07-25
Days Late     = 20
Penalty       = 258.140
```

Total:

```text
Total Overdue = 25.814.000
Total Penalty = 903.490
```

---

# 33. Testing

Buat unit test untuk minimal:

```text
CalculateLoan
DetermineInterestRate
CalculatePenalty
```

Test case bunga:

```text
12 bulan → 12%
18 bulan → 14%
24 bulan → 14%
25 bulan → 16.5%
```

Test case loan:

```text
OTR = 240000000
DP = 20%
Tenor = 18

Expected principal = 192000000
Expected interest = 40320000
Expected total = 232320000
```

Test case penalty:

```text
12907000 × 0.001 × 50
= 645350
```

---

# 34. Coding Rules

Prioritaskan kode yang sederhana.

Gunakan:

```text
Handler
   ↓
Service
   ↓
Repository
   ↓
PostgreSQL
```

Handler tidak boleh berisi business calculation.

Contoh:

```text
Handler
    menerima HTTP request
    validasi input
    panggil service
    return JSON

Service
    business logic

Repository
    SQL/database access
```

Jangan menaruh SQL di handler.

Jangan menaruh business logic di repository.

---

# 35. Jangan Over-Engineering

Jangan menambahkan:

- authentication
- authorization
- JWT
- refresh token
- Redis
- Kafka
- event sourcing
- CQRS
- GraphQL
- Kubernetes
- microservices
- clean architecture yang terlalu kompleks
- generic repository
- dependency injection framework

Kecuali diminta secara eksplisit.

Tujuan utama adalah **BE yang berfungsi dan mudah dipahami**.

---

# 36. Urutan Implementasi

Agent harus mengerjakan secara bertahap.

## Step 1

Buat Go module.

```bash
go mod init ims-finance
```

Install dependency yang diperlukan.

---

## Step 2

Buat Docker Compose PostgreSQL.

Pastikan:

```bash
docker compose up -d
```

berhasil.

---

## Step 3

Buat migration.

Jalankan migration dan pastikan:

```text
ujicoba.clients
ujicoba.contracts
ujicoba.installments
ujicoba.payments
```

berhasil dibuat.

---

## Step 4

Buat database connection menggunakan `pgx`.

Pastikan aplikasi bisa connect ke PostgreSQL.

---

## Step 5

Buat model Go.

Buat:

```text
Client
Contract
Installment
Payment
```

---

## Step 6

Implementasikan loan calculation.

Buat function:

```go
CalculateLoan(...)
```

Pastikan unit test berhasil.

---

## Step 7

Implementasikan repository.

Mulai dari:

```text
ContractRepository
InstallmentRepository
PaymentRepository
```

---

## Step 8

Implementasikan service.

Mulai dari:

```text
LoanService
ContractService
InstallmentService
PaymentService
```

---

## Step 9

Implementasikan HTTP handler.

Endpoint minimal:

```text
POST /api/v1/loans/calculate

POST /api/v1/contracts

GET /api/v1/contracts/:contractNo/installments

POST /api/v1/payments

GET /api/v1/contracts/:contractNo/overdue
```

---

## Step 10

Masukkan seed data Pak Sugus.

Pastikan:

```text
AGR00001
```

memiliki 18 installment.

---

## Step 11

Test overdue.

Gunakan:

```text
2024-08-14
```

sebagai tanggal perhitungan.

Expected:

```text
Installment 6 → 50 hari → 645.350
Installment 7 → 20 hari → 258.140

Total penalty → 903.490
```

---

## Step 12

Buat README.

README minimal harus menjelaskan:

```text
1. Requirement
2. Tech stack
3. Cara menjalankan PostgreSQL
4. Cara menjalankan migration
5. Cara menjalankan application
6. API endpoint
7. Contoh request
8. Contoh response
9. Cara menjalankan test
```

---

# 37. Definition of Done

Project dianggap selesai jika:

- [ ] PostgreSQL bisa dijalankan
- [ ] Schema `ujicoba` tersedia
- [ ] 4 tabel tersedia
- [ ] Go application bisa dijalankan
- [ ] Database connection berhasil
- [ ] Loan calculation berhasil
- [ ] Contract bisa dibuat
- [ ] Installment otomatis dibuat
- [ ] Payment bisa dicatat
- [ ] Installment bisa ditampilkan
- [ ] Overdue bisa dihitung
- [ ] Days late bisa dihitung
- [ ] Penalty bisa dihitung
- [ ] Data Pak Sugus berhasil digunakan
- [ ] Unit test tersedia
- [ ] README tersedia
- [ ] Tidak ada password/secrets di Git

---

# 38. Prinsip Utama Untuk Agent

Jika terdapat beberapa pilihan implementasi, pilih solusi yang:

1. Paling sederhana.
2. Mudah dibaca developer lain.
3. Mudah dites.
4. Menggunakan standard Go.
5. Tidak menambahkan dependency tanpa alasan.
6. Tidak mengubah business rule.
7. Tidak mengubah struktur database tanpa alasan.
8. Jangan membuat fitur yang tidak diminta.

**Jangan melakukan refactoring besar jika tidak diperlukan.**

**Jangan mengubah requirement hanya karena menurut agent ada desain yang lebih bagus.**

Implementasikan requirement terlebih dahulu, kemudian rapikan kode seperlunya.