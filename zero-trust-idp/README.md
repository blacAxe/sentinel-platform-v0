# **Zero Trust Identity Provider (IdP)**

## **Category**

Identity and Access Management (IAM) / Security Engineering

## **Overview**

This project is a Go-based authentication service that replaces traditional passwords with passkeys using WebAuthn and implements a production-style token lifecycle.

It acts as the central identity authority for a zero trust architecture, issuing verifiable identity tokens that are consumed by downstream services such as Sentinel Proxy.

The system now goes beyond basic authentication and demonstrates:

* passwordless authentication
* role-aware authorization
* protected APIs
* session management
* refresh token rotation
* distributed service integration

It is designed to feel closer to a real backend identity service rather than a small authentication demo.

---

## **How it Works**

### **Biometric Authentication**

Users register and authenticate using passkeys through the WebAuthn API.

This enables hardware-backed or biometric authentication such as:

* fingerprint recognition
* Face ID
* Windows Hello
* security keys

without relying on passwords.

---

### **Token-Based Identity**

After successful authentication, the server issues a short-lived JWT access token containing:

* user identity
* username
* role (`admin` or `user`)

This token is then used to access protected APIs and downstream services.

---

### **Refresh Token Lifecycle**

In addition to access tokens, the system issues long-lived refresh tokens.

These tokens are:

* securely stored in PostgreSQL
* hashed before persistence
* tied to server-side sessions
* used to generate new access tokens

This allows users to remain authenticated without repeatedly logging in.

---

### **Zero Trust Enforcement**

Every request to a protected endpoint must include a valid access token.

No request is trusted automatically, and all access decisions are verified through cryptographic validation.

This mirrors real zero trust identity enforcement models used in modern systems.

---

## **Role-Based Access**

The system includes role-aware endpoints to simulate real backend authorization behavior.

### **API Routes**

* `/api/admin`
  Admin-only endpoint

* `/api/user`
  General authenticated user endpoint

* `/api/secret-data`
  Protected endpoint requiring a valid JWT

---

### **Authorization Behavior**

* Admin tokens can access both admin and user routes
* User tokens can only access user routes
* All protected routes require valid JWT verification

This allows downstream systems such as Sentinel Proxy to enforce authorization without requiring direct database access.

---

## **Authentication Flow**

### **Access Tokens**

* Short-lived JWTs
* 15 minute expiration
* Stateless validation
* Contain identity and role claims

Used for authenticating API requests.

---

### **Refresh Tokens**

* Long-lived tokens
* 7 day expiration
* Stored as hashed values in PostgreSQL
* Used to generate new access tokens

Refresh tokens are never stored or transmitted in plain text after issuance.

---

### **Session Management**

* Every login creates a database-backed session
* Refresh tokens are validated against stored sessions
* Logout invalidates the active session immediately

This provides explicit revocation behavior similar to production identity systems.

---

### **Security Design Decisions**

* Refresh tokens are hashed before storage
* Access tokens are intentionally short-lived
* WebAuthn provides phishing-resistant authentication
* Session persistence enables revocation and lifecycle management
* JWT claims allow external authorization enforcement

The architecture mirrors how modern identity providers manage authentication and token rotation internally.

---

## **Distributed System Integration**

This IdP is designed to integrate directly into a distributed security architecture.

### **Identity Pipeline**

1. User authenticates with WebAuthn
2. Server issues access token and refresh token
3. Downstream services validate JWT signatures
4. Sentinel Proxy enforces authorization rules
5. Security events can be propagated into observability pipelines
6. Identity metadata becomes available for centralized logging and analytics

This allows authentication, authorization, and observability systems to remain loosely coupled while sharing trusted identity information.

---

## **Project Structure**

* `/handlers`
  Authentication flows, JWT handling, WebAuthn logic, protected APIs

* `/db`
  PostgreSQL integration, user persistence, session management

* `/internal`
  Supporting utilities and token helpers

* `/static`
  Minimal frontend used for authentication and API testing

---

## **Running the Project**

### **1. Start PostgreSQL**

```bash
docker run --name idp-postgres \
  -e POSTGRES_PASSWORD=password \
  -e POSTGRES_DB=idp \
  -p 5432:5432 \
  -d postgres
```

---

### **2. Create Required Tables**

```sql
CREATE TABLE users (
    id TEXT PRIMARY KEY,
    username TEXT UNIQUE NOT NULL,
    credentials JSONB,
    created_at TIMESTAMP
);

CREATE TABLE sessions (
    id TEXT PRIMARY KEY,
    user_id TEXT,
    refresh_token TEXT,
    expires_at TIMESTAMP,
    created_at TIMESTAMP
);
```

---

### **3. Configure Environment Variables**

Create a `.env` file:

```env
DB_URL=postgres://postgres:password@localhost:5432/idp?sslmode=disable
JWT_SECRET=supersecret
ACCESS_TOKEN_EXPIRY=15m
REFRESH_TOKEN_EXPIRY=168h
```

---

### **4. Run the Server**

```bash
go run main.go
```

---

### **5. Access the Application**

Open:

```txt
http://localhost:8080
```

Register a passkey and log in.

---

## **Testing the System**

### **Login Flow**

* Register a user
* Login using passkey authentication
* Receive access and refresh tokens

---

### **Test Role-Based Routes**

```bash
curl http://localhost:8080/api/admin
curl http://localhost:8080/api/user
```

Expected behavior:

* `/api/admin` returns admin-only data
* `/api/user` returns general user data

---

### **Test Protected Route**

```bash
curl -H "Authorization: Bearer <token>" \
http://localhost:8080/api/secret-data
```

Expected behavior:

* Valid token → protected data returned
* Invalid token → unauthorized response

---

### **Refresh Token Flow**

```bash
curl -H "X-Refresh-Token: <token>" \
http://localhost:8080/auth/refresh
```

Expected behavior:

* New access token returned

---

### **Logout Flow**

```bash
curl -H "X-Refresh-Token: <token>" \
http://localhost:8080/auth/logout
```

Expected behavior:

* Session removed
* Future refresh attempts fail

---

## **Security Notes**

* Refresh tokens are never stored in plain text
* Access tokens expire quickly to reduce exposure
* Sessions are persisted server-side for revocation
* WebAuthn removes password-based attack vectors
* JWT claims allow external authorization enforcement

---

## **Tech Stack**

* Go (Golang)
* WebAuthn (Passkeys)
* PostgreSQL
* JWT (Access and Refresh Tokens)
* Vanilla JavaScript
* Docker

---

## **What This Demonstrates**

This project demonstrates how a modern identity provider works internally:

* passwordless authentication
* secure session lifecycle management
* token-based identity
* role-based access control
* cryptographic verification
* integration with external security systems

The goal was to build a realistic authentication and authorization system that could integrate directly into a broader zero trust architecture.