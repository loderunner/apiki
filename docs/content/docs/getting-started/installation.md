---
title: "Installation"
weight: 11
---

# Installation

apiki can be installed using the official installation script or by downloading pre-built binaries from GitHub releases.

## Quick Install

The easiest way to install apiki is using the installation script:

```shell
curl -fsSL https://raw.githubusercontent.com/loderunner/apiki/main/scripts/install.sh | sh
```

This script will:
1. Detect your operating system and architecture
2. Download the appropriate binary from GitHub releases
3. Install it to `~/.local/share/apiki` (or `$XDG_DATA_HOME/apiki` if set)
4. Update your shell configuration files (`.bashrc`, `.zshrc`, `.config/fish/config.fish`, etc.)

After installation, close and reopen your terminal, or run:

```shell
source ~/.bashrc  # or ~/.zshrc, etc.
```

## Manual Installation

### Download Binary

1. Visit the [GitHub releases page](https://github.com/loderunner/apiki/releases)
2. Download the archive for your platform:
   - `apiki_<version>_darwin_amd64.tar.gz` (macOS Intel)
   - `apiki_<version>_darwin_arm64.tar.gz` (macOS Apple Silicon)
   - `apiki_<version>_linux_amd64.tar.gz` (Linux x86_64)
   - `apiki_<version>_linux_arm64.tar.gz` (Linux ARM64)

3. Extract the archive:
   ```shell
   tar -xzf apiki_<version>_<os>_<arch>.tar.gz
   ```

4. Move the binary to a directory in your PATH:
   ```shell
   mv apiki /usr/local/bin/  # or ~/.local/bin/
   chmod +x /usr/local/bin/apiki
   ```

### Shell Integration

apiki needs to be integrated into your shell to work properly. Add the following to your shell configuration file:

{{< tabs items="Bash,Zsh,Fish" >}}

{{< tab >}}
Add to `~/.bashrc`:

```shell
export APIKI_DIR="$HOME/.local/share/apiki"
[ -s "$APIKI_DIR/init.bash" ] && . "$APIKI_DIR/init.bash"
```
{{< /tab >}}

{{< tab >}}
Add to `~/.zshrc`:

```shell
export APIKI_DIR="$HOME/.local/share/apiki"
[ -s "$APIKI_DIR/init.zsh" ] && . "$APIKI_DIR/init.zsh"
```
{{< /tab >}}

{{< tab >}}
Add to `~/.config/fish/config.fish`:

```fish
set -gx APIKI_DIR "$HOME/.local/share/apiki"
source "$APIKI_DIR/init.fish"
```
{{< /tab >}}

{{< /tabs >}}

## Custom Installation Directory

You can install apiki to a custom directory by setting the `APIKI_DIR` environment variable before running the install script:

```shell
export APIKI_DIR="$HOME/custom/path"
curl -fsSL https://raw.githubusercontent.com/loderunner/apiki/main/scripts/install.sh | bash
```

Or download and extract manually, then set `APIKI_DIR` in your shell configuration.

## Verify Installation

After installation, verify that apiki is working:

```shell
apiki version
```

You should see the version number printed. If you see a "command not found" error, make sure:
1. The binary is in a directory in your PATH, or
2. Your shell configuration has been sourced (restart your terminal)

## Next Steps

Once installed, head to the [Quick Start](/docs/getting-started/quickstart/) guide to learn how to use apiki.
