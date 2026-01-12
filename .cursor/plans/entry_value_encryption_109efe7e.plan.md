---
name: Entry Value Encryption
overview: Implement AES-256-GCM encryption for entry values using Argon2id for password-derived keys or OS keychain for stored random keys. Add CLI subcommands for encryption management outside the main TUI flow.
todos:
  - id: crypto
    content: Create crypto.go with Argon2id KDF and AES-256-GCM encrypt/decrypt
    status: pending
  - id: keychain
    content: Create keychain.go wrapper for cross-platform keychain access
    status: pending
  - id: file-format
    content: Update entry.go to handle new JSON structure with encryption metadata
    status: pending
  - id: commands
    content: Create commands.go with encrypt/decrypt/rotate-key subcommand handlers
    status: pending
  - id: main-routing
    content: Update main.go with subcommand routing and startup unlock flow
    status: pending
  - id: password-prompt
    content: Implement password prompt (stdin) with APIKI_PASSWORD env var support
    status: pending
---

# Entry Value Encryption

## Architecture

```mermaid
flowchart LR
    subgraph cli [CLI Entry Points]
        main[apiki]
        encrypt[apiki encrypt]
        decrypt[apiki decrypt]
        rotate[apiki rotate-key]
    end

    subgraph crypto [Crypto Layer]
        argon[Argon2id KDF]
        aes[AES-256-GCM]
        keychain[OS Keychain]
    end

    subgraph storage [Storage]
        file[variables.json]
    end

    main --> file
    encrypt --> argon
    encrypt --> keychain
    argon --> aes
    keychain --> aes
    aes --> file
    rotate --> crypto
    decrypt --> file
```

## File Format

The `variables.json` file evolves to include an optional encryption header:

```json
{
  "encryption": {
    "mode": "password",
    "salt": "base64...",
    "verifier": "base64..."
  },
  "entries": [
    {
      "name": "DATABASE_URL",
      "value": "enc:v1:base64ciphertext...",
      "label": "Production DB"
    }
  ]
}
```

- `mode`: `"password"` or `"keychain"`
- `salt`: Random 16-byte salt for Argon2id (base64)
- `verifier`: SHA-256 hash of salt, encrypted with the derived key (for fast-fail on wrong password)
- Values prefixed with `enc:v1:` contain: `nonce (12 bytes) || ciphertext || tag (16 bytes)` base64-encoded

## CLI Subcommands

| Command | Description |

| ------------------ | ---------------------------------------------------------------------------------------------------------------- |

| `apiki` | Normal TUI. If encrypted, prompts for password at startup (unless `APIKI_PASSWORD` env var set or keychain mode) |

| `apiki encrypt` | Interactive setup: choose password vs keychain, encrypt all values |

| `apiki decrypt` | Remove encryption, restore plaintext values |

| `apiki rotate-key` | Re-encrypt with new key (can switch between password/keychain) |

## Key Derivation

**Password mode (Argon2id):**

- Memory: 64 MiB
- Iterations: 3
- Parallelism: 4
- Output: 32 bytes (256-bit key)

**Keychain mode:**

- Generate 32-byte random key via `crypto/rand`
- Store in OS keychain under service name `"apiki"`, account `"encryption-key"`

## Dependencies

- `golang.org/x/crypto/argon2` — Argon2id KDF
- `crypto/aes` + `crypto/cipher` — AES-256-GCM (stdlib)
- `github.com/zalando/go-keyring` — Cross-platform keychain access (macOS Keychain, Windows Credential Manager, Linux Secret Service)

## Implementation

### New Files

| File | Purpose |

| ------------- | -------------------------------------------------------- |

| `crypto.go` | Encryption/decryption functions, Argon2id key derivation |

| `keychain.go` | Keychain read/write wrapper |

| `commands.go` | Subcommand handlers (`encrypt`, `decrypt`, `rotate-key`) |

### Modified Files

| File | Changes |

| ---------------------- | -------------------------------------------------------------------------------------- |

| [`entry.go`](entry.go) | Update `LoadEntries`/`SaveEntries` to handle new JSON structure with encryption header |

| [`main.go`](main.go) | Add subcommand routing, password prompt before TUI startup |

## Unlock Flow (Normal Startup)

```mermaid
flowchart TD
    start[apiki starts] --> check{encryption enabled?}
    check -->|no| tui[Launch TUI]
    check -->|yes| mode{mode?}
    mode -->|keychain| fetch[Fetch key from keychain]
    mode -->|password| envcheck{APIKI_PASSWORD set?}
    envcheck -->|yes| useenv[Use env var]
    envcheck -->|no| prompt[Prompt for password]
    fetch --> derive[Derive key]
    useenv --> derive
    prompt --> derive
    derive --> verify{Verify against verifier}
    verify -->|ok| decryptvals[Decrypt values in memory]
    verify -->|fail| error[Exit with error]
    decryptvals --> tui
```
