# **Sentinel Go Security Proxy and Self Healer WAF**

![CI](https://github.com/blacAxe/sentinel-proxy/actions/workflows/ci.yml/badge.svg)

## **Category**

Security Engineering

Sentinel is a lightweight security gateway written in Go that sits in front of a backend service and inspects every request before it reaches the application.

It started as a simple reverse proxy + WAF and has evolved into a small **identity-aware security layer** with logging, metrics, and real-time visibility into traffic and attacks.

---

## **Architecture**

Sentinel is built as a modular Go service with clear separation of concerns:

* **Proxy Layer** — handles all incoming traffic and forwards it to the target service
* **Middleware Chain** — request ID, rate limiting, and WAF inspection
* **Rule Engine** — pattern-based detection for SQLi, XSS, and path abuse
* **Event System** — every request generates a structured security event
* **Metrics Layer** — aggregates traffic, attacks, IP activity, and timelines
* **Dashboard (SSE)** — streams live logs and metrics to the UI
* **Target App** — any backend (local or deployed)

---

## **Core Features**

* Reverse proxy built using Go’s `net/http` and `httputil`
* Rule-based WAF detecting SQL injection, XSS, and suspicious paths
* Per-IP rate limiting to reduce abuse and scanning
* Identity-aware protection for sensitive routes using JWT validation
* Real-time dashboard with:

  * live request logs (SSE)
  * traffic metrics (allowed vs blocked)
  * attack breakdown and top IP tracking
* Structured JSON event pipeline powering both logs and metrics
* Timeout + failure handling for upstream services
* Local persistence using SQLite

---

## **Event & Processing Flow**

Every request goes through a consistent pipeline:

1. Request enters proxy
2. Optional **JWT validation** for protected endpoints
3. Middleware chain applies:

   * request ID
   * rate limiting
   * WAF inspection
4. A structured `SecurityEvent` is created
5. Event is:

   * streamed to the dashboard (logs)
   * recorded into metrics (analytics)
6. Request is either:

   * forwarded to upstream
   * or blocked with a response

This keeps logging, metrics, and security decisions **fully consistent and centralized**.

---

## **Example Events**

```json
{
  "event_type": "security",
  "ip": "::1",
  "path": "/?id=1 union select",
  "attack_detected": true,
  "attack_type": "SQLi UNION",
  "action": "blocked",
  "timestamp": 1714880000
}
```

---

## **How to Run**

```bash
go mod tidy
go run cmd/sentinel/main.go
```

Then open:

http://localhost:8081/dashboard/

---

## **Tech Stack**

* Go (net/http, reverse proxy)
* JWT (authentication / identity validation)
* SQLite (local storage)
* JSON (event pipeline + dashboard streaming)

---

## **Notes**

This project focuses on building security systems from first principles:

* understanding how traffic flows through a proxy
* designing a simple WAF engine
* structuring events for observability
* handling failures and timeouts
* keeping components loosely coupled

It’s intentionally built without heavy frameworks to stay close to how real systems work under the hood.
