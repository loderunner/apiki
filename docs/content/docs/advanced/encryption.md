---
title: "Encryption"
weight: 1
---

apiki can encrypt your variable values, so you can safely share your variables file, back it up to cloud storage, or just have peace of mind that API keys and passwords are protected at rest.

## Encrypting Your Variables

To encrypt your variables file, run:

```shell
apiki encrypt
```

You'll be prompted to choose an unlock method:

```
Lock variables with [p]assword or [k]eychain?
```

### Password Mode

Password mode encrypts your variables with a password you choose. This is the most portable option:

- Works on any machine
- You can share the encrypted file with teammates (they just need the password)
- You need to enter the password each time you launch apiki

After choosing password mode, you'll be asked to enter and confirm your password.

### Keychain Mode

Keychain mode stores the encryption key in your operating system's secure keychain:

- **macOS**: Uses the macOS Keychain
- **Linux**: Uses the Secret Service API (GNOME Keyring, KWallet, etc.)

This is the most convenient option for personal use:

- No password prompts when launching apiki
- The key is tied to your user account on your machine
- Not portable—you can't share the encrypted file with others

## Using Encrypted Variables

Once your variables are encrypted, apiki works exactly the same way. When you launch `apiki`, it automatically detects the encryption and prompts you to unlock:

**Password mode:**

```
Enter password:
```

**Keychain mode:**

```
Unlocking variables with keychain...
```

After unlocking, you can browse, select, create, and edit variables as usual. Values are decrypted in memory only—the file on disk remains encrypted.

## Decrypting Your Variables

If you want to remove encryption and store values in plaintext again:

```shell
apiki decrypt
```

You'll be prompted to unlock (password or keychain), then asked to confirm:

```
Values will be stored in plaintext. Continue? [Y/n]
```

## Rotating Keys

To change your password or switch between password and keychain modes:

```shell
apiki rotate
```

This will:

1. Prompt you to unlock with your current method
2. Ask you to choose a new unlock method (password or keychain)
3. Re-encrypt all variables with the new key

Use this when:

- You want to change your password
- You want to switch from password to keychain mode (or vice versa)
- You suspect your password may have been compromised

## How It Works

For those interested in the technical details:

**Encryption**: Values are encrypted using AES-256-GCM, which provides both confidentiality and integrity protection. Each value has its own random nonce.

**Password Key Derivation**: When using password mode, your password is converted to an encryption key using Argon2id with secure parameters (64 MiB memory, 3 iterations). This makes brute-force attacks impractical.

**Keychain Storage**: In keychain mode, a random 256-bit key is generated and stored in your OS keychain. The key never touches the disk in plaintext.

**File Format**: Encrypted values are stored with a version prefix (`enc:v1:`) followed by base64-encoded ciphertext. The file header contains metadata about the encryption mode and, for password mode, the salt and verifier needed to validate passwords.
