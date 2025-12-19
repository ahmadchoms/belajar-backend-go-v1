# 📦 Inventory API

High-performance, production-ready RESTful Inventory Management API built with **Go (Golang)** using **Clean Architecture**, secured by **JWT Authentication & RBAC**, and fully containerized with **Docker**.

![Go Version](https://img.shields.io/badge/Go-1.23-blue)
![Docker](https://img.shields.io/badge/Docker-Enabled-2496ED)
![License](https://img.shields.io/badge/License-MIT-green)

---

## 🚀 Key Features

### 🧱 Architecture
- Clean Architecture (Handler → Use Case → Repository)
- Clear separation of concerns
- Testable and maintainable codebase

### 🔐 Authentication & Authorization
- Secure password hashing using **bcrypt**
- JWT-based authentication
- **RBAC (Role-Based Access Control)**:
  - User: Read-only access
  - Admin: Full CRUD access

### 🗄️ Data Layer
- **PostgreSQL** with Raw SQL (performance-oriented)
- **Redis** for caching (read optimization & invalidation strategy)

### 📊 API Capabilities
- Pagination & filtering
- Dynamic search queries
- Input validation using `go-playground/validator`

### 🔍 Observability
- Structured JSON logging using `log/slog`
- Request tracing with unique Request ID per request

### 📚 Documentation
- Auto-generated **Swagger / OpenAPI** documentation

### 🧪 Testing
- Unit testing with `stretchr/testify`
- Mock-based repository testing

---

## 🛠️ Tech Stack

| Category        | Technology |
|-----------------|------------|
| Language        | Go 1.23 |
| HTTP Framework  | Native `net/http` |
| Database        | PostgreSQL 15 |
| Cache           | Redis 7 |
| Auth            | JWT |
| Logging         | `log/slog` |
| Container       | Docker & Docker Compose |
| Documentation   | Swagger (Swag) |

---

## 📂 Project Structure

```bash
├── docs/               # Swagger generated documentation
├── handler/            # HTTP handlers (controllers)
├── middleware/         # Logger, Auth, RBAC middlewares
├── mocks/              # Mocks for unit testing
├── models/             # Domain & data models
├── repository/         # Database access (Raw SQL)
├── utils/              # JWT, hashing, response helpers
├── .env                # Environment configuration
├── docker-compose.yml  # Multi-container orchestration
├── Dockerfile          # Application container definition
├── main.go             # Application bootstrap
└── README.md
````

---

## 🏁 Getting Started

### Prerequisites

* Docker & Docker Compose
* (Optional) Go 1.23+ for local development

---

### 1️⃣ Clone Repository

```bash
git clone https://github.com/ahmadchoms/belajar-backend-go-v1
cd inventory-api
```

---

### 2️⃣ Environment Configuration

Create a `.env` file in the root directory:

```env
# Application
APP_PORT=8080

# Database
DB_HOST=db
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=yourpassword
DB_NAME=inventory_db

# Redis
REDIS_HOST=redis
REDIS_PORT=6379

# Security
JWT_SECRET=super_secret_key_change_this_in_production
```

---

### 3️⃣ Run with Docker Compose (Recommended)

```bash
docker-compose up --build
```

The API will be available at:

```
http://localhost:8080
```

---

## 📖 API Documentation (Swagger)

Once the server is running, access Swagger UI at:

```
http://localhost:8080/swagger/index.html
```

### Usage Flow

1. Register a new user via `/register`
2. Login via `/login` to obtain JWT
3. Click **Authorize** in Swagger
4. Use format: `Bearer <your_token>`
5. Access protected endpoints

---

## 🧪 Running Tests

Execute all unit tests:

```bash
go test ./... -v
```

---

## 🔒 Security Notes

* JWT secret **must not** be hardcoded in production
* RBAC enforced at middleware level
* Passwords are never stored in plaintext
* Structured logs avoid leaking sensitive data

---

## 👨‍💻 Author

**ahmadchoms**

* GitHub: [https://github.com/ahmadchoms]
* LinkedIn: [https://www.linkedin.com/in/ahmad-chomsin-aba1b332b/]