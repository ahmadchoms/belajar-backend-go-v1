# 📦 Inventory API (Go + Clean Architecture)

> **"From Zero to Hero" Project.**
> Simulasi Backend High-Performance yang dibangun dari nol untuk mendemonstrasikan implementasi Golang di lingkungan Production.

Project ini menerapkan standar industri backend modern: **Clean Architecture**, **Containerization**, **Distributed Caching**, dan **Graceful Shutdown**.

---

## 🛠 Technology Stack

| Komponen | Teknologi | Keterangan |
| :--- | :--- | :--- |
| **Language** | Golang 1.23 | Strict typing, Concurrency (Goroutines). |
| **Database** | PostgreSQL 16 | Relational DB untuk persistensi data (ACID). |
| **Caching** | Redis (Alpine) | In-memory DB untuk mempercepat *read query*. |
| **Architecture** | Clean Arch | Pemisahan layer: *Handler* ↔ *Repository* ↔ *Models*. |
| **Infra** | Docker & Compose | Multi-stage build (Image size < 30MB). |
| **Driver** | `lib/pq` & `go-redis` | Driver native tanpa ORM (Raw SQL). |

---

## 🚀 Cara Menjalankan (How to Run)

Pastikan kamu sudah menginstall **Docker** dan **Docker Desktop**.

### 1. Clone & Masuk Folder
```bash
git clone https://github.com/ahmadchoms/belajar-backend-go-v1
cd [folder clone kamu]
````

### 2\. Jalankan Aplikasi

Saya sudah menyediakan `Makefile` untuk mempermudah eksekusi.

```bash
# Cara Cepat (Shortcut)
make run

# Atau Cara Manual
docker-compose up --build
```

*Tunggu hingga muncul log: `✅ Tabel 'products' siap!` dan `Server Phase 8 running at :8080`.*

### 3\. Matikan Aplikasi

```bash
# Cara Cepat
make stop

# Atau Cara Manual
docker-compose down
```

-----

## 📡 API Documentation (Cara Test)

Gunakan terminal terpisah untuk mencoba endpoint berikut.

### 1\. Create Product (POST)

Menambahkan data ke PostgreSQL.

```bash
curl -i -X POST http://localhost:8080/products \
  -H "Content-Type: application/json" \
  -d '{"name": "Macbook Pro M3", "price": 25000000, "stock": 50}'
```

**Response:** `201 Created`

### 2\. Get Product by ID (GET)

Mendemonstrasikan strategi Caching (Redis).

```bash
curl -i http://localhost:8080/products/1
```

  * **Request Pertama:** Lambat (Ambil dari DB, simpan ke Redis). Cek log: `🐢 Cache Miss`.
  * **Request Kedua:** Cepat (Langsung dari Redis). Cek log: `🚀 Kena Cache Redis!`.

### 3\. Graceful Shutdown Test

Saat server sedang memproses request, tekan `Ctrl+C`. Server tidak akan mati mendadak, melainkan menunggu request selesai (timeout 10s).

-----

## 📚 Learning Cheat Sheet (Journey Log)

Dokumentasi perintah CLI yang digunakan selama pengembangan project ini (Phase 0 - Phase 8).

### 🐧 Phase 0: Linux & Go Setup

| Command | Detail Command Asli | Fungsi |
| :--- | :--- | :--- |
| **Install WSL** | `wsl --install` | Menginstall Ubuntu subsystem di Windows. |
| **Download Go** | `wget https://go.dev/dl/go1.23.4.linux-amd64.tar.gz` | Download file instalasi Go versi 1.23.4 (Linux). |
| **Ekstrak Go** | `sudo tar -C /usr/local -xzf go1.23.4.linux-amd64.tar.gz` | Membuka file zip Go ke folder sistem `/usr/local`. |
| **Cek Versi** | `go version` | Memastikan instalasi sukses (Wajib muncul `go1.23.4`). |

### 🐳 Phase 1-4: Docker Basics

| Command | Detail Command Asli | Fungsi |
| :--- | :--- | :--- |
| **Permission** | `sudo usermod -aG docker $USER` | Memberi izin user agar bisa `docker run` tanpa sudo. |
| **Run Postgres** | `docker run --name pg-dev -e POSTGRES_PASSWORD=rahasia -p 5432:5432 -d postgres:alpine` | Menyalakan database manual (sebelum ada Docker Compose). |
| **Build Image** | `docker build -t inventory-api:v1 .` | Memasak `Dockerfile` menjadi image siap pakai. |
| **Cek Image** | `docker images` | Melihat daftar image (Pastikan size \< 30MB untuk Go). |
| **Hapus Container**| `docker rm -f pg-dev` | Menghapus container secara paksa (`-f`). |

### 🛠 Phase 2-3: Golang CLI

| Command | Detail Command Asli | Fungsi |
| :--- | :--- | :--- |
| **Init Project** | `go mod init inventory-api` | Membuat KTP project (`go.mod`). |
| **Install Lib** | `go get github.com/lib/pq` | Download driver PostgreSQL. |
| **Install Redis**| `go get github.com/redis/go-redis/v9` | Download library Redis v9. |
| **Bersih-bersih**| `go mod tidy` | Hapus library tak terpakai & download yang kurang. |

### 🐙 Phase 5-8: Docker Compose & Network

| Command | Detail Command Asli | Fungsi |
| :--- | :--- | :--- |
| **Start All** | `docker-compose up --build` | Build ulang kode Go, lalu nyalakan DB + Redis + App bersamaan. |
| **Background** | `docker-compose up -d` | Menyalakan server di latar belakang (terminal tidak terkunci). |
| **Cek Log** | `docker-compose logs -f` | Mengintip log server yang berjalan di background. |
| **Matikan Total**| `docker-compose down` | Mematikan container sekaligus menghapus network-nya. |

-----

## 📂 Project Structure

```text
.
├── cmd/                # (Optional) Entry point
├── handler/            # HTTP Layer (Menerima Request, Validasi JSON)
├── repository/         # DB Layer (Query SQL, Redis Logic)
├── models/             # Struct Data (Product struct)
├── main.go             # Wiring & Dependency Injection
├── Dockerfile          # Resep Multi-stage Build
├── docker-compose.yml  # Orkestrasi Infrastruktur
├── Makefile            # Shortcut command
└── go.mod              # Dependency Management
```