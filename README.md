# CRUD-go-lang

A production-ready RESTful API built with **Golang**, following **Clean Architecture** principles.

## Tech Stack

| Layer | Technology |
|-------|-----------|
| Web Framework | [Gin](https://github.com/gin-gonic/gin) v1.9 |
| ORM | [GORM](https://gorm.io) v1.25 |
| Database | PostgreSQL (SQLite for tests) |
| Authentication | JWT ([golang-jwt/jwt](https://github.com/golang-jwt/jwt) v5) |
| Config | [Viper](https://github.com/spf13/viper) |
| Password Hashing | bcrypt |
| Testing | testing + [testify](https://github.com/stretchr/testify) |

---

## Project Structure

```
CRUD-go-lang/
├── cmd/
│   └── main.go                  # Entry point, DI wiring, route registration
├── config/
│   └── config.go                # Viper-based config loader
├── internal/
│   ├── models/
│   │   ├── user.go              # User model + DTOs
│   │   ├── post.go              # Post model + DTOs
│   │   └── role.go              # Role constants
│   ├── repositories/
│   │   ├── user_repository.go   # DB operations for User
│   │   └── post_repository.go   # DB operations for Post
│   ├── services/
│   │   ├── auth_service.go      # Register / Login logic
│   │   ├── user_service.go      # User business logic
│   │   └── post_service.go      # Post business logic
│   ├── handlers/
│   │   ├── auth_handler.go      # POST /auth/register, POST /auth/login
│   │   ├── user_handler.go      # CRUD /users
│   │   └── post_handler.go      # CRUD /posts
│   ├── middleware/
│   │   ├── auth_middleware.go   # JWT auth + AdminOnly guard
│   │   └── logger_middleware.go # Request logging
│   └── utils/
│       ├── jwt.go               # Token generation/validation
│       ├── password.go          # bcrypt helpers
│       └── response.go          # Standard JSON response helpers
├── migrations/
│   └── migrate.go               # GORM AutoMigrate
├── scripts/
│   └── seed.go                  # Seed data script
├── tests/
│   ├── auth_test.go
│   ├── user_test.go
│   └── post_test.go
├── .env.example
├── Dockerfile
├── docker-compose.yml
└── go.mod
```

---

## Database Schema

### User

| Column     | Type      | Notes                                        |
|------------|-----------|----------------------------------------------|
| id         | uint      | Primary key, auto-increment                  |
| name       | string    | Required                                     |
| email      | string    | Unique, required                             |
| password   | string    | Hashed with bcrypt (never returned in JSON)  |
| role       | string    | `"admin"` or `"user"` (default: `"user"`)    |
| created_at | timestamp |                                              |
| updated_at | timestamp |                                              |
| deleted_at | timestamp | Soft delete                                  |

### Post

| Column     | Type      | Notes                     |
|------------|-----------|---------------------------|
| id         | uint      | Primary key, auto-increment |
| title      | string    | Required, max 255 chars   |
| content    | text      | Required                  |
| user_id    | uint      | FK → users.id             |
| created_at | timestamp |                           |
| updated_at | timestamp |                           |
| deleted_at | timestamp | Soft delete               |

**Relation:** One User has many Posts (1:N).

---

## API Endpoints

### Auth (Public)

| Method | Endpoint         | Description              |
|--------|------------------|--------------------------|
| POST   | `/auth/register` | Register a new user      |
| POST   | `/auth/login`    | Login and get JWT token  |

### Users (Authenticated)

| Method | Endpoint     | Description                    | Permission           |
|--------|--------------|--------------------------------|----------------------|
| GET    | `/users`     | List all users (paginated)     | Admin only           |
| GET    | `/users/:id` | Get user by ID                 | Admin or own profile |
| PUT    | `/users/:id` | Update user                    | Admin or own profile |
| DELETE | `/users/:id` | Delete user                    | Admin or own account |

### Posts (Authenticated)

| Method | Endpoint     | Description                    | Permission           |
|--------|--------------|--------------------------------|----------------------|
| GET    | `/posts`     | List all posts (paginated)     | Any authenticated    |
| GET    | `/posts/:id` | Get post by ID                 | Any authenticated    |
| POST   | `/posts`     | Create a post                  | Any authenticated    |
| PUT    | `/posts/:id` | Update a post                  | Admin or post owner  |
| DELETE | `/posts/:id` | Delete a post                  | Admin or post owner  |

---

## Quick Start

### Option 1: Docker Compose (Recommended)

```bash
# 1. Clone the repository
git clone https://github.com/baodhtv01/CRUD-go-lang.git
cd CRUD-go-lang

# 2. Start the application (app + PostgreSQL)
docker-compose up --build

# 3. The API is available at http://localhost:8080
```

### Option 2: Local Development

**Prerequisites:** Go 1.21+, PostgreSQL running locally.

```bash
# 1. Clone the repository
git clone https://github.com/baodhtv01/CRUD-go-lang.git
cd CRUD-go-lang

# 2. Copy and edit the environment file
cp .env.example .env
# Edit .env with your PostgreSQL credentials

# 3. Download dependencies
go mod download

# 4. Run the server
go run cmd/main.go
```

---

## Configuration

Copy `.env.example` to `.env` and adjust the values:

```env
SERVER_PORT=8080
SERVER_MODE=debug          # debug | release

DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=crudgo
DB_SSLMODE=disable

JWT_SECRET=your-super-secret-jwt-key-change-in-production
JWT_EXPIRES_IN=24          # token expiry in hours
```

---

## Database Migration

The app uses **GORM AutoMigrate** — tables are created/updated automatically on server startup.

To run migration manually:

```bash
go run migrations/migrate.go
```

---

## Seed Data

Populate the database with sample users and posts:

```bash
go run scripts/seed.go
```

This creates:
- 1 admin user: `admin@example.com` / `admin123`
- 2 regular users: `alice@example.com` / `password123`, `bob@example.com` / `password123`
- Sample posts for each user

---

## Running Tests

Tests use an SQLite in-memory database (no PostgreSQL needed):

```bash
go test ./tests/... -v
```

---

## Example Requests (curl)

### Register

```bash
curl -X POST http://localhost:8080/auth/register \
  -H "Content-Type: application/json" \
  -d '{"name":"John Doe","email":"john@example.com","password":"secret123"}'
```

**Response:**

```json
{
  "success": true,
  "message": "user registered successfully",
  "data": {
    "id": 1,
    "name": "John Doe",
    "email": "john@example.com",
    "role": "user",
    "created_at": "2024-01-01T00:00:00Z"
  }
}
```

---

### Login

```bash
curl -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"john@example.com","password":"secret123"}'
```

**Response:**

```json
{
  "success": true,
  "message": "login successful",
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIs...",
    "user": {
      "id": 1,
      "name": "John Doe",
      "email": "john@example.com",
      "role": "user"
    }
  }
}
```

Save the token for subsequent requests:

```bash
TOKEN="eyJhbGciOiJIUzI1NiIs..."
```

---

### Create a Post

```bash
curl -X POST http://localhost:8080/posts \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"title":"My First Post","content":"Hello World!"}'
```

**Response:**

```json
{
  "success": true,
  "message": "post created successfully",
  "data": {
    "id": 1,
    "title": "My First Post",
    "content": "Hello World!",
    "user_id": 1,
    "created_at": "2024-01-01T00:00:00Z"
  }
}
```

---

### List Posts (with Pagination)

```bash
curl "http://localhost:8080/posts?page=1&limit=10" \
  -H "Authorization: Bearer $TOKEN"
```

**Response:**

```json
{
  "success": true,
  "message": "posts retrieved successfully",
  "data": [...],
  "page": 1,
  "limit": 10,
  "total": 42,
  "total_pages": 5
}
```

---

### Get User by ID

```bash
curl http://localhost:8080/users/1 \
  -H "Authorization: Bearer $TOKEN"
```

---

### Update User

```bash
curl -X PUT http://localhost:8080/users/1 \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"name":"Jane Doe"}'
```

---

### Delete a Post

```bash
curl -X DELETE http://localhost:8080/posts/1 \
  -H "Authorization: Bearer $TOKEN"
```

---

## Error Responses

All errors follow a consistent format:

```json
{
  "success": false,
  "message": "descriptive error message",
  "error": "detailed error info (in debug mode)"
}
```

| HTTP Status | Meaning |
|-------------|---------|
| 400 | Bad Request — invalid input / validation failure |
| 401 | Unauthorized — missing or invalid JWT token |
| 403 | Forbidden — insufficient permissions |
| 404 | Not Found — resource does not exist |
| 409 | Conflict — e.g., email already in use |
| 500 | Internal Server Error |

---

## Authorization Rules

| Role  | Users endpoint               | Posts endpoint             |
|-------|------------------------------|----------------------------|
| admin | Full CRUD on all users       | Full CRUD on all posts     |
| user  | Read/update/delete own account | Create, read all; update/delete own posts only |

---

## Docker

### Build and run with Docker Compose

```bash
docker-compose up --build
```

### Services started

| Service  | Port | Notes |
|----------|------|-------|
| app      | 8080 | Go API server |
| postgres | 5432 | PostgreSQL 15 |

### Stop and clean up

```bash
docker-compose down -v   # also removes volumes
```

---

## License

MIT
