---
name: Entry Value Encryption
overview: Implement AES-256-GCM encryption for entry values using Argon2id for password-derived keys or OS keychain for stored random keys. Add CLI subcommands for encryption management outside the main TUI flow.
todos:
  - id: crypto
    content: Create crypto.go with Argon2id KDF and AES-256-GCM encrypt/decrypt
    status: pending
  - id: keychain
    content: Create keychain.go using go-keychain (Touch ID on macOS, Secret Service on Linux)
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
  - id: ci-cgo
    content: Update release workflow for CGO (matrix builds, libsecret on Linux)
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
- `verifier`: HMAC-SHA256(derived_key, salt) for fast-fail on wrong password (base64)
- Values prefixed with `enc:v1:` contain: `nonce (12 bytes) || ciphertext || tag (16 bytes)` base64-encoded

## CLI Subcommands

All subcommands are implemented as Cobra commands. To preserve shell integration (`eval $(apiki)`), subcommands must:

- Write all user-facing output to **stderr** (messages, prompts, confirmations)
- Write **nothing** to stdout (the shell would try to eval it)

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
- **macOS**: `AccessControlUserPresence` flag triggers Touch ID (or passcode fallback)
- **Linux**: D-Bus Secret Service (GNOME Keyring, KWallet) — no biometric prompt

## Dependencies

- `golang.org/x/crypto/argon2` — Argon2id KDF
- `crypto/aes` + `crypto/cipher` — AES-256-GCM (stdlib)
- `github.com/keybase/go-keychain` — macOS Keychain (Touch ID) + Linux Secret Service

## Build Requirements

`go-keychain` requires CGO for native keychain APIs:

- **macOS**: Xcode command line tools (for Security framework)
- **Linux**: `libsecret-1-dev` (Debian/Ubuntu) or `libsecret-devel` (Fedora/RHEL)

### CI/CD Changes

Current release workflow uses single Ubuntu runner with `CGO_ENABLED=0`. With CGO:

| File | Changes |

| ----------------------------------- | ---------------------------------------------------------------- |

| `.goreleaser.yaml` | Set `CGO_ENABLED=1`, split builds by OS |

| `.github/workflows/release.yml` | Matrix strategy: macOS runner for darwin, Linux runner for linux |

GoReleaser supports `--split` and `--merge` for multi-runner builds.

## Implementation

### New Files

| File | Purpose |

| ------------- | -------------------------------------------------------- |

| `crypto.go` | Encryption/decryption functions, Argon2id key derivation |

| `keychain.go` | Keychain wrapper using `go-keychain` (Touch ID on macOS, Secret Service on Linux) |

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
    fetch -. macOS .-> touchid((Touch ID))
    mode -->|password| envcheck{APIKI_PASSWORD set?}
    envcheck -->|yes| useenv[Use env var]
    envcheck -->|no| prompt[Prompt for password]
    useenv --> derive[Derive key with Argon2id]
    prompt --> derive
    fetch --> verify{Verify against verifier}
    derive --> verify
    verify -->|ok| decryptvals[Decrypt values in memory]
    verify -->|fail| error[Exit with error]
    decryptvals --> tui
```