# Golang Senior Microservices Showcase

A production-inspired event-driven microservices showcase written in Go.

This project demonstrates how modern backend systems implement reliable event publishing using the **Transactional Outbox Pattern** together with an **Idempotent Consumer (Inbox Pattern)** to guarantee safe message processing under an **at-least-once delivery** model.

---

## Architecture

```
                    HTTP API
                        │
                        ▼
               Customer Service
                        │
         MongoDB Transaction (ACID)
        ┌───────────────┴────────────────┐
        ▼                                ▼
   customers                      outbox_events
                                         │
                                         ▼
                                Outbox Worker
                                         │
                                         ▼
                                     Kafka Topic
                                         │
                                         ▼
                               Customer Consumer
                                         │
                                         ▼
                                processed_events
                                         │
                                         ▼
                              Business Handler
```

---

# Features

- Clean Architecture
- Repository Pattern
- Dependency Injection
- MongoDB Transactions
- Transactional Outbox Pattern
- Background Outbox Worker
- Kafka Producer
- Kafka Consumer
- Idempotent Consumer (Inbox Pattern)
- MongoDB Replica Set
- Swagger/OpenAPI
- Docker Compose
- Kafka UI
- Unit Tests
- Structured Logging
- Graceful Shutdown
- Configuration via Environment Variables

---

# Tech Stack

| Technology | Version |
|------------|----------|
| Go | 1.26 |
| MongoDB | 7 |
| Apache Kafka | Confluent Platform 7.7 |
| Kafka UI | Provectus |
| Docker | Latest |
| Swagger | Swaggo |
| slog | Standard Library |

---

# Project Structure

```
cmd/
    api/
    customer-consumer/

internal/
    config/
    domain/
    handlers/
    messaging/
    repositories/
    services/
    workers/

deployments/
    docker-compose.yml
```

---

# Event Flow

## 1. Client creates a customer

```
POST /customers
```

The API validates the request and starts a MongoDB transaction.

---

## 2. Transactional Outbox

Inside the same transaction:

- Customer document is inserted.
- Outbox event is inserted.

If one operation fails, both are rolled back.

---

## 3. Outbox Worker

The background worker periodically:

- Claims pending events.
- Publishes events to Kafka.
- Marks events as published.

---

## 4. Kafka

Events are published using:

```
Key = EventID
```

Example payload:

```json
{
  "eventId":"...",
  "eventType":"customer.created",
  "id":"...",
  "firstName":"John",
  "lastName":"Doe",
  "email":"john@example.com"
}
```

---

## 5. Idempotent Consumer

The consumer:

- Reads Kafka events.
- Attempts to reserve EventID.
- Executes business logic.
- Marks the event as completed.

Duplicate deliveries are ignored safely.

---

# MongoDB Collections

## customers

Stores customer information.

## outbox_events

Stores events waiting for publication.

Status lifecycle:

```
PENDING
↓

PROCESSING
↓

PUBLISHED

or

FAILED
```

---

## processed_events

Implements the Inbox Pattern.

```
{
    eventId
    eventType
    status
    startedAt
    completedAt
}
```

A unique MongoDB index guarantees the same EventID is never processed twice.

---

# Delivery Guarantees

This project intentionally follows the standard distributed systems model:

```
At-Least-Once Delivery
```

instead of

```
Exactly Once
```

Exactly-once delivery across a database and Kafka is generally impractical without distributed transactions. Instead, this project achieves reliable processing by combining:

- Transactional Outbox
- Stable EventID
- Kafka Event Key
- Durable Inbox
- Idempotent Consumer

---

# Running

Start everything:

```bash
docker compose -f deployments/docker-compose.yml up -d --build
```

API

```
http://localhost:8080
```

Swagger

```
http://localhost:8080/swagger/index.html
```

Kafka UI

```
http://localhost:8082
```

---

# Testing

Create a customer:

```bash
curl -X POST http://localhost:8080/customers \
-H "Content-Type: application/json" \
-d '{
  "firstName":"John",
  "lastName":"Doe",
  "email":"john@example.com"
}'
```

Watch the consumer:

```bash
docker compose logs -f customer-consumer
```

Expected output:

```
customer-created event handled
```

Verify MongoDB:

```
customers

outbox_events

processed_events
```

---

# Docker Services

| Service | Purpose |
|----------|----------|
| api | HTTP REST API |
| customer-consumer | Kafka Consumer |
| mongo | MongoDB Replica Set |
| mongo-init-replica | Replica Set Initialization |
| kafka | Kafka Broker |
| kafka-ui | Kafka Administration |

---

# Reliability Features

✔ MongoDB ACID Transactions

✔ Transactional Outbox

✔ EventID-based Publishing

✔ Background Event Publisher

✔ Idempotent Consumer

✔ Inbox Pattern

✔ Duplicate Detection

✔ Manual Kafka Offset Commits

✔ Graceful Shutdown

✔ Structured Logging

---

# Future Improvements

- Dead Letter Queue (DLQ)
- OpenTelemetry Distributed Tracing
- Prometheus Metrics
- Grafana Dashboard
- Integration Tests (Testcontainers)
- Horizontal Consumer Scaling
- Retry Policies
- Stale Inbox Recovery
- CI/CD Pipeline (GitHub Actions)

---

# Design Patterns

- Clean Architecture
- Repository Pattern
- Dependency Injection
- Transactional Outbox
- Inbox Pattern
- Background Worker
- Producer / Consumer
- Event-Driven Architecture
- Idempotent Consumer

---
# Kubernetes Deployment

This project was migrated from a Docker Compose-based environment to a local Kubernetes environment using Docker Desktop Kubernetes.

The objective was to deploy and test the complete event-driven workflow with:

- Go REST API
- MongoDB
- MongoDB replica set
- Persistent storage
- Kafka
- Kafka producer
- Kafka consumer group
- Idempotent consumer
- Kubernetes Deployments
- Kubernetes Services
- Kubernetes Job
- PersistentVolumeClaim

## What Was Implemented

The following components were deployed to Kubernetes:

```text
Go API
MongoDB
MongoDB PersistentVolumeClaim
MongoDB Service
MongoDB Replica Set Initialization Job
Kafka
Kafka Service
Customer Consumer

Client
  |
  | POST /customers
  v
Go API
  |
  | Store customer
  v
MongoDB
  |
  | Publish customer.created event
  v
Kafka topic: customers.events
  |
  | Consumer group reads event
  v
customer-consumer
  |
  | Process event
  v
MongoDB processed_events collection
  |
  | Commit Kafka offset
  v
Consumer lag = 0

+----------------------+
|        Client        |
+----------+-----------+
           |
           | POST /customers
           v
+----------------------+
|       Go API         |
| Kubernetes Deployment|
+----------+-----------+
           |
           | Save customer
           v
+----------------------+
|      MongoDB         |
| Replica Set + PVC    |
+----------------------+

           |
           | Publish customer.created
           v
+----------------------+
|       Kafka          |
| customers.events     |
+----------+-----------+
           |
           | customer-consumer-group
           v
+----------------------+
|  Customer Consumer   |
| Idempotent Consumer  |
+----------+-----------+
           |
           | Save processed EventID
           v
+----------------------+
| processed_events     |
| MongoDB collection   |
+----------------------+

# License

MIT
