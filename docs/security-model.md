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

## OIDC Token Validation

### Example: Validating JWT

```go
import "github.com/coreos/go-oidc/v3/oidc"

func validateToken(ctx context.Context, rawToken string) (*Claims, error) {
    // Create OIDC provider (fetches public keys)
    provider, err := oidc.NewProvider(ctx, "https://auth.cloudplane.io")
    if err != nil {
        return nil, err
    }

    // Create verifier
    verifier := provider.Verifier(&oidc.Config{
        ClientID: "credential-broker",
    })

    // Verify token signature and claims
    idToken, err := verifier.Verify(ctx, rawToken)
    if err != nil {
        return nil, err  // Invalid token
    }

    // Extract claims
    var claims Claims
    idToken.Claims(&claims)
    return &claims, nil
}
```

---

## AWS STS Integration

### Example: AssumeRoleWithWebIdentity

```go
import (
    "github.com/aws/aws-sdk-go-v2/config"
    "github.com/aws/aws-sdk-go-v2/service/sts"
)

func assumeRole(ctx context.Context, roleARN, oidcToken string) (*Credentials, error) {
    cfg, _ := config.LoadDefaultConfig(ctx)
    stsClient := sts.NewFromConfig(cfg)

    result, err := stsClient.AssumeRoleWithWebIdentity(ctx, &sts.AssumeRoleWithWebIdentityInput{
        RoleArn:          aws.String(roleARN),
        WebIdentityToken: aws.String(oidcToken),
        RoleSessionName:  aws.String("cloudplane-session"),
        DurationSeconds:  aws.Int32(900),  // 15 minutes
    })
    if err != nil {
        return nil, err
    }

    return &Credentials{
        AccessKeyID:     *result.Credentials.AccessKeyId,
        SecretAccessKey: *result.Credentials.SecretAccessKey,
        SessionToken:    *result.Credentials.SessionToken,
        ExpiresAt:       *result.Credentials.Expiration,
    }, nil
}
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

```go
const MaxTTL = 900  // 15 minutes

func issueCredentials(ttl int32) int32 {
    if ttl > MaxTTL || ttl <= 0 {
        return MaxTTL
    }
    return ttl
}
```

### 4. Service Isolation

| Service | Internet-facing | Credential access | Protocol |
|---------|-----------------|-------------------|----------|
| Control Plane API | Yes | None | REST |
| Credential Broker | No (internal) | STS only | gRPC |
| Orchestrator | No (internal) | Via broker | gRPC client |
| Observability | No (internal) | Read-only | REST |

---

## IAM Trust Policy

Users set this up in their AWS account:

```json
{
  "Version": "2012-10-17",
  "Statement": [{
    "Effect": "Allow",
    "Principal": {
      "Federated": "arn:aws:iam::CLOUDPLANE_ACCOUNT:oidc-provider/auth.cloudplane.io"
    },
    "Action": "sts:AssumeRoleWithWebIdentity",
    "Condition": {
      "StringEquals": {
        "auth.cloudplane.io:aud": "credential-broker"
      }
    }
  }]
}
```

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

```go
// To revoke cloudplane access:
// 1. Delete IAM trust policy in user's account
// 2. All subsequent AssumeRoleWithWebIdentity calls fail immediately
// 3. Existing credentials expire within 15 minutes max

// Example: Checking if access is revoked
_, err := stsClient.AssumeRoleWithWebIdentity(ctx, input)
if err != nil {
    // Access denied - trust policy removed
    return ErrAccessRevoked
}
```
