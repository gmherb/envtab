# SOPS Integration Guide

This document explains how to use SOPS (Secrets OPerationS) encryption with envtab.

## Overview

envtab now supports SOPS encryption in two ways:
1. **File-level encryption**: Encrypt entire loadout YAML files with SOPS
2. **Value-level encryption**: Encrypt individual environment variable values with SOPS

## Prerequisites

1. Install SOPS: https://github.com/getsops/sops
2. Configure SOPS with your preferred encryption backend (AWS KMS, GCP KMS, Azure Key Vault, age, PGP, etc.)
3. Set up your `.sops.yaml` configuration file (optional, but recommended)

## Usage

### File-Level Encryption

Encrypt an entire loadout file with SOPS:

```bash
# Create a new loadout with file-level SOPS encryption
envtab add myloadout --sops-file MY_VAR=value

# Or set environment variable to always use SOPS
export ENVTAB_USE_SOPS=true
envtab add myloadout MY_VAR=value
```

**Benefits:**
- Entire file is encrypted, including metadata
- Can be edited with `sops myloadout.yaml` directly
- Works seamlessly with existing SOPS workflows

### Value-Level Encryption

Encrypt individual values with SOPS:

```bash
# Encrypt a single value
envtab add myloadout --sops-value SECRET_KEY=mysecret

# Combine with sensitive flag (falls back to SOPS if GCP KMS not configured)
envtab add myloadout -s API_KEY=apikey123
```

**Benefits:**
- Granular control over which values are encrypted
- Mix encrypted and plaintext values in the same file
- Values are decrypted automatically on export

### Automatic Decryption

When exporting loadouts, SOPS-encrypted values are automatically decrypted:

```bash
# Export will automatically decrypt SOPS values
envtab export myloadout

# Reading loadouts automatically handles SOPS-encrypted files
envtab cat myloadout
```

## Configuration

### SOPS Configuration File

Create a `.sops.yaml` in your project root or home directory:

```yaml
creation_rules:
  - path_regex: .*\.yaml$
    kms: 'arn:aws:kms:us-east-1:123456789012:key/12345678-1234-1234-1234-123456789012'
    pgp: >-
      FBC7B9E2A4F9289AC0C1D4843D16CEE4A27381B4,
      85D77543B3D624B63CEA9E06DCCB5A08F57A8DA3
```

### Environment Variables

- `ENVTAB_USE_SOPS=true`: Enable file-level SOPS encryption by default
- `ENVTAB_GCP_KMS_KEY`: GCP KMS key for GCP encryption (existing)

## Examples

### Example 1: File-Level Encryption

```bash
# Create encrypted loadout
envtab add production --sops-file \
  DB_PASSWORD=secret123 \
  API_KEY=key456

# File is encrypted, can be viewed with sops
sops ~/.envtab/production.yaml

# Export automatically decrypts
envtab export production
```

### Example 2: Value-Level Encryption

```bash
# Mix encrypted and plaintext values
envtab add staging \
  DB_HOST=localhost \
  --sops-value DB_PASSWORD=secret123 \
  DEBUG=true

# Only DB_PASSWORD is encrypted
envtab cat staging
```

### Example 3: Using with Age (SOPS backend)

```bash
# Generate age key
age-keygen -o ~/.config/sops/age/keys.txt

# Configure .sops.yaml
cat > .sops.yaml <<EOF
creation_rules:
  - path_regex: .*\.yaml$
    age: >-
      age1example1q2w3e4r5t6y7u8i9o0p1a2s3d4f5g6h7j8k9l0
EOF

# Use SOPS encryption
envtab add secrets --sops-file API_KEY=mykey
```

## Implementation Details

### File Operations

- `ReadLoadout()`: Automatically detects and decrypts SOPS-encrypted files
- `WriteLoadoutWithEncryption()`: Optionally encrypts files with SOPS
- `IsSOPSEncrypted()`: Checks if a file is SOPS-encrypted

### Value Operations

- `SOPSEncryptValue()`: Encrypts a single value
- `SOPSDecryptValue()`: Decrypts a SOPS-encrypted value
- Values prefixed with `SOPS:` are automatically decrypted on export

### Backend Support

SOPS supports multiple encryption backends:
- **AWS KMS**: `kms: 'arn:aws:kms:...'`
- **GCP KMS**: `gcp_kms: 'projects/...'`
- **Azure Key Vault**: `azure_kv: 'https://...'`
- **age**: `age: 'age1...'`
- **PGP**: `pgp: 'FINGERPRINT'`
- **HashiCorp Vault**: `vault: '...'`

Configure these in your `.sops.yaml` file.

## Migration

### From GCP KMS to SOPS

```bash
# Old way (GCP KMS per-value)
envtab add myloadout -s SECRET=value

# New way (SOPS file-level)
envtab add myloadout --sops-file SECRET=value

# Or value-level
envtab add myloadout --sops-value SECRET=value
```

## Troubleshooting

### "sops command not found"
Install SOPS: `brew install sops` or download from https://github.com/getsops/sops

### "sops encryption failed"
- Check your `.sops.yaml` configuration
- Verify your encryption backend credentials (AWS/GCP/Azure keys)
- Test with `sops -e test.yaml` directly

### "Failed to decrypt SOPS value"
- Ensure you have access to the encryption keys
- Check that SOPS can decrypt the file: `sops -d file.yaml`
- Verify your `.sops.yaml` configuration matches the encryption method used

## Security Considerations

1. **File Permissions**: SOPS-encrypted files maintain 0600 permissions
2. **Key Management**: Store encryption keys securely (use key management services)
3. **Access Control**: Use IAM/ACLs to control who can decrypt files
4. **Audit Logging**: SOPS operations can be logged for compliance

## See Also

- [SOPS Documentation](https://github.com/getsops/sops)
- [SOPS Configuration](https://github.com/getsops/sops#using-sops-yaml-conf-to-select-kms-pgp-for-new-files)

