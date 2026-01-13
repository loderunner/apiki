#!/bin/sh
cat <<'EOF'

================================================================================
  apiki has been installed to /usr/share/apiki/

  IMPORTANT: apiki requires shell integration to work. It is NOT on your PATH.
  
  Add the following to your shell configuration:

  For bash (~/.bashrc):
    export APIKI_DIR="/usr/share/apiki"
    [ -s "$APIKI_DIR/init.bash" ] && . "$APIKI_DIR/init.bash"

  For zsh (~/.zshrc):
    export APIKI_DIR="/usr/share/apiki"
    [ -s "$APIKI_DIR/init.zsh" ] && . "$APIKI_DIR/init.zsh"

  For fish (~/.config/fish/config.fish):
    set -gx APIKI_DIR "/usr/share/apiki"
    source "$APIKI_DIR/init.fish"

  Then restart your shell or source your config file.
================================================================================

EOF
