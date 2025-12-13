# Inventory API (Golang + Clean Arch)

Simulasi Backend Inventory System yang High-Performance dan Scalable.
Dibangun dengan fokus pada standar industri dan arsitektur yang bersih.

## 🛠 Tech Stack
- **Language:** Golang 1.23
- **Database:** PostgreSQL (Primary), Redis (Caching)
- **Architecture:** Clean Architecture (Handler, Repository, Model)
- **Infrastructure:** Docker & Docker Compose
- **Features:** Graceful Shutdown, Dependency Injection

## 🚀 How to Run
Prerequisites: Docker & Docker Desktop installed.

1. **Clone Repository**
   ```bash
   git clone [https://github.com/username/inventory-api.git](https://github.com/username/inventory-api.git)
   cd inventory-api

2. **Run Application**
make run
# Atau: docker-compose up --build

3. **Tests Endpoints**
- Create Product:
```curl -X POST http://localhost:8080/products -d '{"name":"Macbook", "price":20000000, "stock":5}'
- Get Product (Cache Strategy):
```curl http://localhost:8080/products/1