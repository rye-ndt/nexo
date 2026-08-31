#!/usr/bin/env bash
set -eo pipefail

root=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)
data_dir=$(sed -n 's/^[[:space:]]*data_dir:[[:space:]]*\([^[:space:]#]*\).*/\1/p' "$root/config.yaml" | head -1)
data_dir=${data_dir:-nexo}
bundle_id=com.wails.nexo

targets=""
add() { [ -e "$1" ] && targets="$targets$1"$'\n'; return 0; }

case "$(uname -s)" in
  Darwin)
    add "/Applications/Nexo.app"
    add "$HOME/Applications/Nexo.app"
    add "$HOME/Library/Application Support/$data_dir"
    add "$HOME/Library/WebKit/$bundle_id"
    add "$HOME/Library/Caches/$bundle_id"
    add "$HOME/Library/HTTPStorages/$bundle_id"
    add "$HOME/Library/Preferences/$bundle_id.plist"
    add "$HOME/Library/Saved Application State/$bundle_id.savedState"
    ;;
  Linux)
    add "${XDG_CONFIG_HOME:-$HOME/.config}/$data_dir"
    add "/usr/bin/nexo"
    add "/usr/share/applications/nexo.desktop"
    add "/usr/share/pixmaps/nexo.png"
    ;;
  *)
    echo "uninstall: unsupported platform $(uname -s). On Windows, remove Nexo from Apps & features." >&2
    exit 1
    ;;
esac

add "$root/build/bin"

deb=""
if command -v dpkg-query >/dev/null 2>&1 &&
   dpkg-query -W -f='${Status}' nexo 2>/dev/null | grep -q "install ok installed"; then
  deb=nexo
fi

if [ -z "$targets" ] && [ -z "$deb" ]; then
  echo "Nexo is not installed: nothing to remove."
  exit 0
fi

echo "This permanently removes Nexo and everything it stores, including the"
echo "workflow database, the installed agent harnesses and every credential:"
echo
[ -n "$deb" ] && echo "  apt package  nexo"
printf %s "$targets" | sed 's/^/  /'
echo

if [ "$FORCE" != 1 ]; then
  printf "Type 'yes' to continue: "
  read -r reply 2>/dev/null < /dev/tty || { echo; echo "Aborted: no terminal to confirm on. Re-run with FORCE=1."; exit 1; }
  [ "$reply" = yes ] || { echo "Aborted."; exit 1; }
fi

pkill -f "Nexo.app/Contents/MacOS/nexo" 2>/dev/null || true
pkill -x nexo 2>/dev/null || true

[ -n "$deb" ] && sudo dpkg --purge nexo

while IFS= read -r target; do
  [ -e "$target" ] || continue
  rm -rf "$target" 2>/dev/null && continue
  echo "  needs elevation: $target"
  sudo rm -rf "$target"
done <<< "$targets"

echo "Nexo removed."
