# Security Model

This document describes the security model for cloudplane.

## Overview

cloudplane operates on a **delegated trust** model using OIDC and short-lived credentials. The platform never stores cloud access keys or secrets.

---

## Authentication Flow

```
User → Control Plane API (JWT over REST)
          ↓
Orchestrator → Credential Broker (gRPC + service JWT)
          ↓
Credential Broker → AWS STS (AssumeRoleWithWebIdentity)
          ↓
Orchestrator → User's AWS Account (temp credentials)
```

---

## Communication Security

| Communication | Protocol | Security |
|---------------|----------|----------|
| User → API | REST/HTTPS | JWT in Authorization header |
| Orchestrator → Credential Broker | gRPC/TLS | Mutual TLS + service JWT |
| Internal services | gRPC/TLS | mTLS between pods |

---

## Key Principles

### 1. Delegated Trust via OIDC

- Users configure IAM trust policies in their AWS accounts
- cloudplane exchanges OIDC tokens for temporary STS credentials
- Credentials expire in 15-60 minutes
- Revoke access instantly by deleting the trust policy

### 2. No Credential Storage

| What we store | What we NEVER store |
|---------------|---------------------|
| `project_id` | Access keys |
| `role_arn` | Secret keys |
| Audit logs | Session tokens |
| | OIDC tokens |

### 3. Short-Lived Credentials

- Max TTL: 900 seconds (15 minutes)
- Credentials are vended on-demand, used immediately, then discarded
- Never written to disk or database

### 4. Least Privilege

- Users define minimal IAM permissions in their accounts
- No `AdministratorAccess` required
- cloudplane requests only what's needed for the operation

### 5. Service Isolation

| Service | Internet-facing | Credential access | Protocol |
|---------|-----------------|-------------------|----------|
| Control Plane API | Yes | None | REST |
| Credential Broker | No (internal) | STS only | gRPC |
| Orchestrator | No (internal) | Via broker | gRPC client |
| Observability | No (internal) | Read-only | REST |

---

## Audit Trail

All operations are logged:

1. **cloudplane logs**: Every credential request, job execution, API call
2. **AWS CloudTrail**: All AWS API calls with cloudplane's session name
3. **User visibility**: Users see all operations in their CloudTrail

---

## Threat Mitigation

| Threat | Mitigation |
|--------|------------|
| Stolen credentials | Short-lived (15 min), revocable |
| Credential leakage | Never stored, never logged |
| Privilege escalation | User-defined IAM policies |
| Insider threat | Audit logs, no persistent access |
| Service compromise | Blast radius limited to single service |
| Man-in-the-middle | TLS/mTLS for all communication |

---

## Immediate Revocation

To revoke cloudplane access:

1. Delete IAM trust policy in user's account
2. All subsequent `AssumeRoleWithWebIdentity` calls fail immediately
3. Existing credentials expire within 15 minutes max
