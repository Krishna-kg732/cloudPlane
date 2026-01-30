# Security Model

OIDC-based delegated trust for zero-credential access to user AWS accounts.

---

## Overview

cloudplane **never stores AWS credentials**. Instead:

1. User creates IAM role with OIDC trust policy
2. cloudplane exchanges OIDC token for short-lived STS credentials
3. Credentials expire in 15 minutes
4. User can revoke access instantly by deleting trust policy

---

## OIDC Trust Flow

```
┌─────────────┐     ┌───────────────┐     ┌─────────────┐
│ cloudplane  │     │   AWS STS     │     │ User's IAM  │
│ Credential  │────▶│               │────▶│   Role      │
│ Broker      │     │ AssumeRole    │     │             │
└─────────────┘     │ WithWebIdentity     └─────────────┘
       │            └───────────────┘
       │                    │
       ▼                    ▼
   OIDC Token         Temp Credentials
   (from IdP)           (15 min TTL)
```

---

## IAM Role Setup

Users deploy this CloudFormation:

```yaml
AWSTemplateFormatVersion: '2010-09-09'
Description: cloudplane OIDC trust

Parameters:
  OIDCProviderArn:
    Type: String
    Default: arn:aws:iam::CLOUDPLANE_ACCOUNT:oidc-provider/auth.cloudplane.io

Resources:
  CloudplaneRole:
    Type: AWS::IAM::Role
    Properties:
      RoleName: CloudplaneRole
      AssumeRolePolicyDocument:
        Version: '2012-10-17'
        Statement:
          - Effect: Allow
            Principal:
              Federated: !Ref OIDCProviderArn
            Action: sts:AssumeRoleWithWebIdentity
            Condition:
              StringEquals:
                auth.cloudplane.io:aud: credential-broker
                auth.cloudplane.io:sub: !Sub "project:${AWS::AccountId}"
      ManagedPolicyArns:
        - !Ref CloudplanePolicy

  CloudplanePolicy:
    Type: AWS::IAM::ManagedPolicy
    Properties:
      PolicyDocument:
        Version: '2012-10-17'
        Statement:
          # EKS
          - Effect: Allow
            Action:
              - eks:*
            Resource: "*"
          # EC2 (for node groups)
          - Effect: Allow
            Action:
              - ec2:*
            Resource: "*"
          # FSx
          - Effect: Allow
            Action:
              - fsx:*
            Resource: "*"
          # S3 (state + data)
          - Effect: Allow
            Action:
              - s3:*
            Resource:
              - arn:aws:s3:::cloudplane-*
              - arn:aws:s3:::cloudplane-*/*
          # IAM (for service accounts)
          - Effect: Allow
            Action:
              - iam:CreateRole
              - iam:AttachRolePolicy
              - iam:CreateOpenIDConnectProvider
            Resource: "*"

Outputs:
  RoleArn:
    Value: !GetAtt CloudplaneRole.Arn
```

---

## Credential Lifecycle

| Step | Duration | What Happens |
|------|----------|--------------|
| 1 | 0s | Orchestrator requests credentials |
| 2 | ~100ms | Broker calls STS AssumeRoleWithWebIdentity |
| 3 | ~100ms | STS returns temp credentials |
| 4 | 15 min | Credentials valid, Terraform/kubectl execute |
| 5 | 15 min+ | Credentials expire, discarded |

**Never stored**: Credentials exist only in memory during execution.

---

## Revocation

User can revoke access instantly:

```bash
# Delete the trust policy
aws iam delete-role-policy --role-name CloudplaneRole --policy-name trust

# Or delete the entire role
aws cloudformation delete-stack --stack-name cloudplane-trust
```

**Effect**: Next credential request fails immediately.

---

## Audit Trail

All operations logged in user's CloudTrail:

```json
{
  "eventName": "AssumeRoleWithWebIdentity",
  "userIdentity": {
    "type": "WebIdentityUser",
    "principalId": "arn:aws:iam::USER_ACCOUNT:role/CloudplaneRole",
    "webIdFederationData": {
      "federatedProvider": "arn:aws:iam::CLOUDPLANE:oidc-provider/auth.cloudplane.io"
    }
  }
}
```

Plus cloudplane logs every credential request:
- Timestamp
- Project ID
- Role ARN (no credentials logged)
- Success/failure

---

## Security Guarantees

| Guarantee | How |
|-----------|-----|
| No credential storage | Vended on-demand, discarded after use |
| Short-lived | 15-min max TTL |
| User-controlled | IAM permissions defined by user |
| Instant revocation | Delete trust policy |
| Full audit | CloudTrail + cloudplane logs |
| Least privilege | User scopes permissions to what's needed |
