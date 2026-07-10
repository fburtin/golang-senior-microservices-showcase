# Go Senior Microservices Showcase

A production-inspired backend application built in Go demonstrating REST APIs, MongoDB, Apache Kafka, event-driven architecture, Docker, and Swagger.

> This project was created to showcase backend engineering skills commonly required for Senior Go Developer positions.

---

# Architecture

```
                    +-----------------------+
                    |       Client          |
                    +-----------+-----------+
                                |
                          HTTP REST API
                                |
                                v
                     +--------------------+
                     |    Customer API    |
                     +--------------------+
                      |                |
                      |                |
                 MongoDB          Kafka Producer
                      |                |
                      |                v
                      |      customers.events
                      |                |
                      |                |
                      +----------------+
                                       |
                                       v
                           Customer Consumer
```

---

# Features

- RESTful API
- Customer CRUD
- MongoDB Persistence
- Apache Kafka Producer
- Apache Kafka Consumer
- Event-Driven Architecture
- Docker Compose Environment
- Swagger / OpenAPI Documentation
- Health Check Endpoint
- Environment-based Configuration

---

# Technology Stack

| Technology | Version |
|------------|---------|
| Go | 1.26 |
| MongoDB | 7 |
| Apache Kafka | 7.7 |
| Kafka UI | Latest |
| Docker | Latest |
| Swagger | swaggo |
| REST API | net/http |

---

# Project Structure

```
cmd/
│
├── api/
│   └── main.go
│
└── customer-consumer/
    └── main.go

internal/

├── config/
├── handlers/
├── messaging/
│   ├── producer.go
│   ├── consumer.go
│   └── customer_events.go
│
├── models/
├── repository/
└── services/

docker-compose.yml
```

---

# Event Flow

```
POST /customers
        │
        ▼
 Save Customer in MongoDB
        │
        ▼
 Publish CustomerCreated Event
        │
        ▼
customers.events
        │
        ▼
Customer Consumer
        │
        ▼
Process Event
```

---

# Kafka Event

Current event:

```
CustomerCreated
```

Example payload

```json
{
  "eventType": "customer.created",
  "id": "20260709182921",
  "firstName": "John",
  "lastName": "Doe",
  "email": "john@example.com",
  "createdAt": "2026-07-09T18:29:21Z"
}
```

---

# Running the Project

## Start Infrastructure

```bash
docker compose up -d
```

This starts:

- MongoDB
- Apache Kafka
- Kafka UI

---

## Start the API

```bash
go run ./cmd/api
```

---

## Start the Consumer

```bash
go run ./cmd/customer-consumer
```

---

# Swagger

```
http://localhost:8080/swagger/index.html
```

---

# Kafka UI

```
http://localhost:8082
```

---

# Example Request

```http
POST /customers
```

```json
{
    "firstName":"John",
    "lastName":"Doe",
    "email":"john@example.com"
}
```

---

# Example Consumer Output

```
Received message

topic=customers.events
partition=0
offset=3

{
   "eventType":"customer.created",
   "id":"20260709182921",
   "firstName":"John",
   "lastName":"Doe",
   "email":"john@example.com"
}
```

---

# Current Architecture

- REST API
- Repository Pattern
- Service Layer
- MongoDB
- Apache Kafka
- Kafka Producer
- Kafka Consumer
- Docker Compose
- Swagger

---

# Next Steps

Planned improvements:

- Outbox Pattern
- Dead Letter Queue (DLQ)
- Retry Policy
- Multiple Kafka Consumers
- Structured Logging
- JWT Authentication
- Integration Tests
- Kubernetes Deployment

---

# Author

**Francisco J. Burtin**

Senior Backend Developer

Technologies:

- Go
- .NET
- Azure
- Kafka
- MongoDB
- Docker

GitHub

https://github.com/fburtin

---

# License

MIT
