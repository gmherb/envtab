# SOPS Integration Guide

This document explains how to use SOPS (Secrets OPerationS) encryption with envtab.

## Overview

envtab supports SOPS encryption in two ways:

1. **File-level encryption**: Encrypt entire loadout YAML files with SOPS
2. **Value-level encryption**: Encrypt individual environment variable values with SOPS

## Prerequisites

1. Install SOPS: https://github.com/getsops/sops
2. Configure SOPS with your preferred encryption backend (AWS KMS, GCP KMS, Azure Key Vault, age, PGP, etc.)
3. Set up your `.sops.yaml` configuration file (optional, but recommended)

## Usage

### File-Level Encryption

Encrypt an entire loadout file with SOPS.

```bash
# Create a new loadout with file-level SOPS encryption
envtab add myloadout --encrypt-file MY_VAR=value
# Or use the short form
envtab add myloadout -f MY_VAR=value

# Using the --encrypt-file (or -f) flag will always encrypt the entire file
# Even when value-level encryption (or no encryption) was used prior
envtab add myloadout --encrypt-file MY_VAR=value
```

**Benefits:**

- Entire file is encrypted, including metadata
- Can be edited with `sops myloadout.yaml` directly
- Works seamlessly with existing SOPS workflows
- Faster than value-based encryption if multiple secrets stored on single loadout.

### Value-Level Encryption

Encrypt individual values with SOPS.

```bash
# Encrypt a single value
envtab add myloadout -e SECRET_KEY=mysecret

# Or use the long form
envtab add myloadout --encrypt-value API_KEY=apikey123
```

**Benefits:**

- Granular control over which values are encrypted
- Mix encrypted and plaintext values in the same file
- Keep loadouts organized even if you have secrets

### Automatic Decryption

When exporting loadouts, SOPS-encrypted data is automatically decrypted.

```bash
# Export will automatically decrypt SOPS values
envtab export myloadout

# Reading loadouts automatically handles SOPS-encrypted files when --decrypt option is provided
envtab cat myloadout -d
```

## Configuration

### SOPS Configuration File

Create a `.sops.yaml` in your project root or home directory.

You must have creation rule(s) that match:

1. `path_regex: envtab-stdin-override`

*`envtab-stdin-override` is used for internal envtab processes that read from stdin (`sops --filename-override`). This prevents writing sensitive data to disk*

2. A creation rule for loadouts stored in your `ENVTAB_DIR` directory.

- Loadouts always end with `.yaml` and are stored directly in `ENVTAB_DIR` (no sub-directories).
- Replace `ENVTAB_DIR` in the path_regex with your actual directory path (e.g., `~/.local/share/envtab` or the value of your `ENVTAB_DIR` environment variable).
- The pattern should match loadout filenames (letters, numbers, dashes, and underscores are common).

*Dash (-), underscore (_), and numbers are included here in the example, but naming is up to you.*

```yaml
creation_rules:
  - path_regex: envtab-stdin-override
    kms: 'arn:aws:kms:us-east-1:123456789012:key/12345678-1234-1234-1234-123456789012'

  - path_regex: ^~/\.local/share/envtab/[a-zA-Z0-9_-]+\.yaml$
    kms: 'arn:aws:kms:us-east-1:123456789012:key/12345678-1234-1234-1234-123456789012'
```

If no other SOPS creation rules exist you can use a single rule to catch all (omit `path_regex`).

```yaml
creation_rules:
  - kms: 'arn:aws:kms:us-east-1:123456789012:key/12345678-1234-1234-1234-123456789012'
```

## Examples

### Example 1: File-Level Encryption

```bash
# Create encrypted loadout
envtab add production --encrypt-file DB_PASSWORD=secret123

# File is encrypted, can be viewed with sops
sops $ENVTAB_DIR/production.yaml  # or ~/.local/share/envtab/production.yaml by default

# Export automatically decrypts
envtab export production
```

### Example 2: Value-Level Encryption

```bash
# Mix encrypted and plaintext values
envtab add staging DB_HOST=localhost
envtab add staging -e DB_PASSWORD=secret123

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
  - age: >-
      age1example1q2w3e4r5t6y7u8i9o0p1a2s3d4f5g6h7j8k9l0
EOF

# Use SOPS file encryption
envtab add secrets --encrypt-file API_KEY=mykey
# All consecutive entries added to file encrypted loadouts are encrypted
envtab add secrets PASSWORD=secret
```

## Implementation Details

### SOPS Metadata Preservation

**Important**: SOPS requires metadata (encryption keys used, MAC, etc.) to decrypt files. Our implementation preserves this:

1. **File-level encryption**: The entire file is encrypted with SOPS, preserving all metadata in the file itself. This is the recommended approach.

2. **Value-level encryption**: When encrypting individual values:
   - We create a temporary YAML file: `value: <secret>`
   - Encrypt it with SOPS (which adds metadata)
   - Store the **entire SOPS-encrypted YAML structure** (including metadata) as the value
   - On decryption, we write the full encrypted structure to a temp file and decrypt it
   - This ensures all SOPS metadata is preserved

### Key Rotation Handling

SOPS encryption keys can be rotated (AWS KMS key rotation, age key changes, etc.). When this happens:

1. **Automatic Detection**: The system detects key rotation errors and provides helpful messages
2. **Graceful Degradation**: Failed decryptions are skipped with warnings (not fatal errors)
3. **Re-encryption Support**: Use `SOPSReencryptFile()` to re-encrypt files with current keys

### File Operations

- `ReadLoadout()`: Automatically detects and decrypts SOPS-encrypted files
  - Handles key rotation errors gracefully
  - Provides helpful error messages
- `WriteLoadoutWithEncryption()`: Optionally encrypts files with SOPS
- `IsSOPSEncrypted()`: Checks if a file is SOPS-encrypted
- `SOPSCanDecrypt()`: Checks if a file can be decrypted with current keys
- `SOPSReencryptFile()`: Re-encrypts a file with current keys (for key rotation)

### Value Operations

- `SOPSEncryptValue()`: Encrypts a single value (preserves full SOPS metadata)
- `SOPSDecryptValue()`: Decrypts a SOPS-encrypted value (uses preserved metadata)
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

## Key Rotation

When encryption keys are rotated (e.g., AWS KMS key rotation, age key changes):

### Symptoms
- Decryption errors when reading loadouts
- Warnings like "keys may have been rotated" when exporting
- `sops -d` fails on encrypted files

### Solution

**For file-level encrypted loadouts:**
```bash
# Re-encrypt with current keys
sops -r -i $ENVTAB_DIR/myloadout.yaml  # or ~/.local/share/envtab/myloadout.yaml by default
```

**For value-level encrypted entries:**
You'll need to re-add the values with current keys:
```bash
# Remove old encrypted value entry
envtab edit myloadout --remove-entry SECRET  # or edit manually

# Re-add with current keys
envtab add myloadout -e SECRET=newvalue
```

### Prevention

- Use key management systems that support key versioning (AWS KMS, GCP KMS via SOPS)
- Keep old keys accessible during rotation period
- Use multiple encryption keys in SOPS config for redundancy

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

### "keys may have been rotated"
- This means the encryption keys used to encrypt the data are no longer available
- Re-encrypt the loadout with current keys (see Key Rotation section above)
- For file-level encryption: `sops -r -i $ENVTAB_DIR/myloadout.yaml` (or `~/.local/share/envtab/myloadout.yaml` by default)
- For value-level encryption: Re-add the values with `--encrypt-value` (or `-e`) flag

## Security Considerations

1. **File Permissions**: SOPS-encrypted files maintain 0600 permissions
2. **Key Management**: Store encryption keys securely (use key management services)
3. **Access Control**: Use IAM/ACLs to control who can decrypt files
4. **Audit Logging**: SOPS operations can be logged for compliance

## See Also

- [SOPS Documentation](https://github.com/getsops/sops)
- [SOPS Configuration](https://github.com/getsops/sops#using-sops-yaml-conf-to-select-kms-pgp-for-new-files)

