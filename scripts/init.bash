# Auto-restore apiki state on shell startup (opt-in via APIKI_AUTO_RESTORE)
# Only runs in the first shell, not subshells (APIKI_RESTORED marker)
if [ -n "$APIKI_AUTO_RESTORE" ] && [ -z "$APIKI_RESTORED" ]; then
  eval "$("${APIKI_DIR:-$HOME/.local/share/apiki}/apiki" restore 2>/dev/null)"
  export APIKI_RESTORED=1
fi

apiki() {
  eval "$("${APIKI_DIR:-$HOME/.local/share/apiki}/apiki" "$@")"
}
