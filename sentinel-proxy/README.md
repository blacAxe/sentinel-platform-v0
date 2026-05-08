# **Sentinel Go Security Proxy and Self Healer WAF**

![CI](https://github.com/blacAxe/sentinel-proxy/actions/workflows/ci.yml/badge.svg)

## **Category**

Security Engineering / Distributed Systems

Sentinel is a lightweight security gateway written in Go that sits in front of a backend service and inspects every request before it reaches the application.

It started as a simple reverse proxy + WAF and evolved into a small identity-aware security platform with logging, metrics, real-time monitoring, and distributed event ingestion.

The system now integrates with the LumenLog observability pipeline, allowing security events to move through Kafka / Redpanda into ClickHouse for centralized analytics.

---

## **Architecture**

Sentinel is built as a modular security system with clear separation of concerns:

* **Proxy Layer** — handles incoming traffic and forwards requests upstream
* **Middleware Chain** — request IDs, authentication, rate limiting, and WAF inspection
* **Rule Engine** — detects SQLi, XSS, path traversal, and suspicious access patterns
* **Identity Provider** — JWT-based authentication and WebAuthn support
* **Event Pipeline** — structured security events generated from every request
* **Metrics Layer** — aggregates attacks, request volume, IP activity, and timelines
* **Dashboard (SSE)** — streams live logs and metrics to connected clients
* **Rust Security Agent** — receives and ships security events into Kafka
* **LumenLog Ingestor** — consumes Kafka events and stores them in ClickHouse
* **Target Applications** — protected backend services

---

## **Core Features**

* Reverse proxy built using Go’s `net/http` and `httputil`
* Rule-based WAF detecting:
  * SQL Injection
  * XSS payloads
  * suspicious paths
  * admin route abuse
* Per-IP rate limiting
* Identity-aware access control using JWT validation
* WebAuthn authentication flow
* Real-time dashboard with:
  * live request logs
  * attack visibility
  * metrics streaming
  * request analytics
* Structured security event generation
* Distributed security event pipeline using Kafka / Redpanda
* Rust-based event forwarding service
* ClickHouse persistence for observability and analytics
* Dockerized multi-service deployment
* Failure handling and upstream timeout protection

---

## **Distributed Security Pipeline**

Sentinel now includes a distributed telemetry pipeline built with Rust, Kafka, and ClickHouse.

### **Pipeline Flow**

1. Sentinel Proxy blocks or inspects suspicious traffic
2. A structured `SecurityEvent` is generated
3. Event is forwarded to the Rust security agent
4. Rust agent serializes the event using Protocol Buffers
5. Event is published into Kafka / Redpanda
6. LumenLog ingests the event
7. Logs are persisted into ClickHouse for querying and analytics

This allows Sentinel to move beyond a standalone WAF into a distributed security telemetry system.

---

## **Event & Processing Flow**

Every request follows a centralized processing pipeline:

1. Request enters Sentinel Proxy
2. Optional JWT validation for protected endpoints
3. Middleware chain applies:
   * request ID generation
   * rate limiting
   * WAF inspection
4. A structured `SecurityEvent` is generated
5. Event is:
   * streamed to dashboard clients
   * aggregated into metrics
   * optionally shipped to the Rust security pipeline
6. Request is either:
   * forwarded upstream
   * or blocked immediately

This keeps logging, metrics, and enforcement fully consistent across the system.

---

## **Example Security Event**

```json
{
  "event_type": "security",
  "ip": "::1",
  "path": "/admin?id=1 union select",
  "attack_detected": true,
  "attack_type": "SQLi UNION",
  "action": "blocked",
  "timestamp": 1714880000
}
```

---

## **Testing**

The project includes middleware, rule engine, and integration testing using Go’s testing framework.

Run tests:

```bash
go test ./...
```

CI automatically runs:

* dependency verification
* project build validation
* automated tests

on every push and pull request.

---

## **How to Run**

### **Run Locally**

```bash
go mod tidy
go run cmd/sentinel/main.go
```

Sentinel starts on:

```txt
http://localhost:8081
```

Dashboard:

```txt
http://localhost:8081/dashboard/
```

---

## **Docker Deployment**

Build and start the full stack:

```bash
docker compose up --build
```

This launches:

* Sentinel Proxy
* PostgreSQL
* Identity Provider
* Rust Security Agent
* Redpanda / Kafka
* ClickHouse
* LumenLog Ingestor

---

## **Tech Stack**

* Go
* Rust
* net/http
* Reverse Proxy
* JWT Authentication
* WebAuthn
* SQLite
* PostgreSQL
* Server-Sent Events (SSE)
* Docker
* GitHub Actions CI
* Kafka / Redpanda
* ClickHouse
* Protocol Buffers

---

## **Project Goals**

This project focuses on building security systems from first principles:

* understanding reverse proxy internals
* designing lightweight WAF logic
* implementing identity-aware access control
* building distributed event pipelines
* structuring logs and metrics for observability
* handling failures and upstream instability
* keeping services modular and loosely coupled

The goal was not only to block malicious traffic, but also to understand how modern security infrastructure is designed underneath the abstractions.