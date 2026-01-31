set -q APIKI_DIR; or set APIKI_DIR "$HOME/.local/share/apiki"

# Auto-restore apiki state on shell startup (opt-in via APIKI_AUTO_RESTORE)
# Only runs in the first shell, not subshells (APIKI_RESTORED marker)
if set -q APIKI_AUTO_RESTORE; and not set -q APIKI_RESTORED
  eval ("$APIKI_DIR/apiki" restore 2>/dev/null)
  set -gx APIKI_RESTORED 1
end

function apiki
  eval ("$APIKI_DIR/apiki" $argv)
end
