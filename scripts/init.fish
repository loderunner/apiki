function apiki
  set -q APIKI_DIR; or set APIKI_DIR "$HOME/.local/share/apiki"
  eval ("$APIKI_DIR/apiki" $argv)
end
