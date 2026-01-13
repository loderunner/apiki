---
title: "Installation"
weight: 11
---

apiki can be installed using the official installation script, various package managers, or manually from GitHub releases.

## Installation Script

The easiest way to install apiki is using the official installation script:

```shell
curl -fsSL https://github.com/loderunner/apiki/releases/latest/download/install.sh | sh
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

### Custom Installation Directory

You can install apiki to a custom directory by setting the `APIKI_DIR` environment variable before running the install script:

```shell
export APIKI_DIR="$HOME/custom/path"
curl -fsSL https://github.com/loderunner/apiki/releases/latest/download/install.sh | bash
```

## Package Managers

### Homebrew (macOS/Linux)

```bash
brew install loderunner/tap/apiki
```

After installation, add the following to your shell configuration:

{{< tabs items="Bash,Zsh,Fish" >}}

{{< tab >}}
Add to `~/.bashrc`:

```bash
export APIKI_DIR="$(brew --prefix)/share/apiki"
[ -s "$APIKI_DIR/init.bash" ] && . "$APIKI_DIR/init.bash"
```
{{< /tab >}}

{{< tab >}}
Add to `~/.zshrc`:

```bash
export APIKI_DIR="$(brew --prefix)/share/apiki"
[ -s "$APIKI_DIR/init.zsh" ] && . "$APIKI_DIR/init.zsh"
```
{{< /tab >}}

{{< tab >}}
Add to `~/.config/fish/config.fish`:

```fish
set -gx APIKI_DIR (brew --prefix)/share/apiki
source "$APIKI_DIR/init.fish"
```
{{< /tab >}}

{{< /tabs >}}

### mise (formerly rtx)

```bash
mise use -g apiki
```

After installation, add the following to your shell configuration:

{{< tabs items="Bash,Zsh,Fish" >}}

{{< tab >}}
Add to `~/.bashrc`:

```bash
export APIKI_DIR="$HOME/.local/share/mise/installs/apiki/$(mise current apiki)"
[ -s "$APIKI_DIR/init.bash" ] && . "$APIKI_DIR/init.bash"
```
{{< /tab >}}

{{< tab >}}
Add to `~/.zshrc`:

```bash
export APIKI_DIR="$HOME/.local/share/mise/installs/apiki/$(mise current apiki)"
[ -s "$APIKI_DIR/init.zsh" ] && . "$APIKI_DIR/init.zsh"
```
{{< /tab >}}

{{< tab >}}
Add to `~/.config/fish/config.fish`:

```fish
set -gx APIKI_DIR "$HOME/.local/share/mise/installs/apiki/(mise current apiki)"
source "$APIKI_DIR/init.fish"
```
{{< /tab >}}

{{< /tabs >}}

## Manual Installation

### Debian/Ubuntu (.deb)

Download the `.deb` package from the [GitHub releases page](https://github.com/loderunner/apiki/releases):

```bash
# Download the latest .deb for your architecture (amd64 or arm64)
curl -LO https://github.com/loderunner/apiki/releases/latest/download/apiki_<version>_linux_amd64.deb

# Optional: Verify signature
curl -LO https://github.com/loderunner/apiki/releases/latest/download/apiki_<version>_linux_amd64.deb.sig
curl -LO https://loderunner.github.io/apiki/apiki-signing-key.asc
gpg --import apiki-signing-key.asc
gpg --verify apiki_<version>_linux_amd64.deb.sig apiki_<version>_linux_amd64.deb

# Install with dpkg
sudo dpkg -i apiki_<version>_linux_amd64.deb
```

After installation, add shell integration (see below).

### Fedora/RHEL/CentOS (.rpm)

Download the `.rpm` package from the [GitHub releases page](https://github.com/loderunner/apiki/releases):

```bash
# Download the latest .rpm for your architecture (amd64 or arm64)
curl -LO https://github.com/loderunner/apiki/releases/latest/download/apiki_<version>_linux_amd64.rpm

# Optional: Verify signature
curl -LO https://github.com/loderunner/apiki/releases/latest/download/apiki_<version>_linux_amd64.rpm.sig
curl -LO https://loderunner.github.io/apiki/apiki-signing-key.asc
gpg --import apiki-signing-key.asc
gpg --verify apiki_<version>_linux_amd64.rpm.sig apiki_<version>_linux_amd64.rpm

# Install with rpm or dnf
sudo rpm -i apiki_<version>_linux_amd64.rpm
# or
sudo dnf install ./apiki_<version>_linux_amd64.rpm
```

After installation, add shell integration (see below).

### Shell Integration for .deb/.rpm

After installing via `.deb` or `.rpm`, add the following to your shell configuration:

{{< tabs items="Bash,Zsh,Fish" >}}

{{< tab >}}
Add to `~/.bashrc`:

```bash
export APIKI_DIR="/usr/share/apiki"
[ -s "$APIKI_DIR/init.bash" ] && . "$APIKI_DIR/init.bash"
```
{{< /tab >}}

{{< tab >}}
Add to `~/.zshrc`:

```bash
export APIKI_DIR="/usr/share/apiki"
[ -s "$APIKI_DIR/init.zsh" ] && . "$APIKI_DIR/init.zsh"
```
{{< /tab >}}

{{< tab >}}
Add to `~/.config/fish/config.fish`:

```fish
set -gx APIKI_DIR "/usr/share/apiki"
source "$APIKI_DIR/init.fish"
```
{{< /tab >}}

{{< /tabs >}}

### Download Binary Archive

1. Visit the [GitHub releases page](https://github.com/loderunner/apiki/releases)
2. Download the archive for your platform:
   - `apiki_<version>_darwin_amd64.tar.gz` (macOS Intel)
   - `apiki_<version>_darwin_arm64.tar.gz` (macOS Apple Silicon)
   - `apiki_<version>_linux_amd64.tar.gz` (Linux x86_64)
   - `apiki_<version>_linux_arm64.tar.gz` (Linux ARM64)

3. Verify the checksum (optional but recommended):
   ```shell
   # Download the checksums file and signature
   curl -LO https://github.com/loderunner/apiki/releases/latest/download/apiki_<version>_checksums.txt
   curl -LO https://github.com/loderunner/apiki/releases/latest/download/apiki_<version>_checksums.txt.sig
   
   # Verify the signature (requires GPG and the signing key)
   gpg --verify apiki_<version>_checksums.txt.sig apiki_<version>_checksums.txt
   
   # Verify the checksum
   sha256sum -c apiki_<version>_checksums.txt --ignore-missing
   ```

4. Extract the archive:
   ```shell
   tar -xzf apiki_<version>_<os>_<arch>.tar.gz
   ```

5. Move the binary and init scripts to your desired location (e.g., `~/.local/share/apiki/`):
   ```shell
   mkdir -p ~/.local/share/apiki
   mv apiki init.bash init.zsh init.fish ~/.local/share/apiki/
   chmod +x ~/.local/share/apiki/apiki
   ```

### Shell Integration for Manual Install

Add the following to your shell configuration file:

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

## Important Notes

**apiki is NOT installed on your PATH** - it must be used via a shell function. This is by design:

- The `apiki` command is a shell function that wraps the binary with `eval`
- Running the binary directly just prints `export` commands to stdout, which is useless
- By installing to a non-PATH location, users who haven't set up shell integration get a clear "command not found" error instead of confusing output

## Verifying

After installation and shell integration, run `apiki version` to confirm it's working.
