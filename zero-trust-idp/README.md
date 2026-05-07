# Zero Trust Identity Provider (IdP)

## Category

Identity and Access Management (IAM) / Security Engineering

## Overview

This project is a Go-based authentication service that replaces traditional passwords with passkeys using WebAuthn and implements a production-style token lifecycle. It acts as the central identity authority for a zero trust architecture, issuing verifiable identity tokens that are consumed by downstream services such as a security proxy.

The system now goes beyond basic authentication and demonstrates role-aware identity, protected APIs, and real-world service integration patterns.

---

## How it Works

### Biometric Authentication

Users register and authenticate using passkeys through the WebAuthn API. This enables hardware-backed or biometric authentication such as fingerprint or face recognition without relying on passwords.

### Token-Based Identity

After successful authentication, the server issues a short-lived access token containing:

* user identity
* username
* role (admin or user)

This token is used to access protected resources across services.

### Refresh Token Lifecycle

In addition to access tokens, the system issues long-lived refresh tokens. These are stored securely in the database and used to obtain new access tokens without requiring the user to authenticate again.

### Zero Trust Enforcement

Every request to a protected endpoint must include a valid access token. No request is trusted by default, and all access is verified through cryptographic validation.

---

## Role-Based Access (New)

The system now includes role-aware endpoints to simulate real backend authorization.

### API Routes

* `/api/admin`
  Returns admin-only data

* `/api/user`
  Returns general user data

* `/api/secret-data`
  Protected route requiring a valid JWT

### Behavior

* Admin tokens can access both admin and user routes
* User tokens can only access user routes
* All protected routes require a valid JWT

This allows downstream systems (like Sentinel) to enforce authorization without needing direct database access.

---

## Authentication Flow (Production Design)

### Access Tokens

* Short-lived JWTs with a 15 minute expiration
* Contain user identity and role
* Used for authenticating API requests
* Stateless and verified on every request

### Refresh Tokens

* Long-lived tokens with a 7 day expiration
* Stored in PostgreSQL as hashed values
* Used to generate new access tokens
* Never stored or transmitted in plain text

### Session Management

* Each login creates a session record in the database
* Refresh tokens are validated against stored sessions
* Logout deletes the session, immediately invalidating the refresh token

### Security Design Decisions

* Refresh tokens are hashed before storage
* Access tokens are short-lived to limit exposure
* WebAuthn provides phishing-resistant authentication
* Server-side session storage allows explicit revocation

This mirrors how real identity providers manage sessions and token rotation.

---

## Central Pipeline Integration

This IdP is designed to plug directly into a distributed security pipeline:

1. Authentication
   The user logs in and receives an access token and refresh token

2. Authorization
   Tokens include roles that downstream services can enforce

3. Verification
   External services validate JWT signatures using the shared secret

4. Enforcement
   A proxy layer (Sentinel) can enforce access control based on token claims

5. Observability
   Identity data can be propagated for logging and monitoring

---

## Project Structure

* `/handlers`
  Authentication logic, WebAuthn flows, JWT handling, and API endpoints

* `/db`
  PostgreSQL integration, user persistence, and session storage

* `/internal`
  Supporting utilities and token logic

* `/static`
  Minimal frontend used to test authentication and protected routes

---

## Running the Project

### 1. Start PostgreSQL

```bash
docker run --name idp-postgres \
  -e POSTGRES_PASSWORD=password \
  -e POSTGRES_DB=idp \
  -p 5432:5432 \
  -d postgres
```

---

### 2. Create Required Tables

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

### 3. Configure Environment Variables

Create a `.env` file:

```env
DB_URL=postgres://postgres:password@localhost:5432/idp?sslmode=disable
JWT_SECRET=supersecret
ACCESS_TOKEN_EXPIRY=15m
REFRESH_TOKEN_EXPIRY=168h
```

---

### 4. Run the Server

```bash
go run main.go
```

---

### 5. Access the Application

Open:

http://localhost:8080

Register a passkey and log in.

---

## Testing the System

### Login

* Register a user
* Login with passkey
* Receive access and refresh tokens

### Test Role-Based Routes

```bash
curl http://localhost:8080/api/admin
curl http://localhost:8080/api/user
```

Expected:

* `/api/admin` → admin data
* `/api/user` → general data

### Test Protected Route

```bash
curl -H "Authorization: Bearer <token>" http://localhost:8080/api/secret-data
```

Expected:

* Valid token → returns secret data
* Missing/invalid token → unauthorized

### Refresh Token

```bash
curl -H "X-Refresh-Token: <token>" http://localhost:8080/auth/refresh
```

Expected:

* New access token returned

### Logout

```bash
curl -H "X-Refresh-Token: <token>" http://localhost:8080/auth/logout
```

Expected:

* Session deleted
* Future refresh attempts fail

---

## Security Notes

* Refresh tokens are never stored in plain text
* Access tokens expire quickly to reduce risk
* Sessions are stored server-side for revocation
* WebAuthn removes password-based attack vectors
* Role claims allow external enforcement without DB access

---

## Tech Stack

* Go (Golang)
* WebAuthn (Passkeys)
* PostgreSQL
* JWT (Access and Refresh Tokens)
* Vanilla JavaScript (Frontend)

---

## What This Demonstrates

This project shows how a modern identity provider works in practice:

* passwordless authentication
* secure session management
* token-based identity
* role-based access control
* integration with external enforcement layers

It is designed to feel like a real backend system rather than a demo, and it connects directly into a broader zero trust architecture.
