#!/usr/bin/env bash
# Build and install the Devzat GIF plugin.
#
# Run as root from this directory:
#   sudo ./install.sh
#
# Pass config inline to avoid editing the env file afterwards:
#   sudo DEVZAT_TOKEN=dvz@... GIF_PROVIDER=giphy ./install.sh
set -euo pipefail

SRC_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BIN=/usr/local/bin/devzat-gif
SVC=/etc/systemd/system/devzat-gif.service
ENV_FILE=/etc/devzat-gif.env

echo "==> Building devzat-gif"
( cd "$SRC_DIR" && go build -o "$BIN" . )
chmod 755 "$BIN"

echo "==> Installing systemd unit"
install -m 0644 "$SRC_DIR/devzat-gif.service" "$SVC"

if [ ! -f "$ENV_FILE" ]; then
    echo "==> Creating $ENV_FILE"
    install -m 0600 /dev/null "$ENV_FILE"
    cat >"$ENV_FILE" <<'EOF'
# Devzat GIF plugin configuration
# DEVZAT_TOKEN: same token the other plugins use
DEVZAT_TOKEN=
# GIF_PROVIDER: giphy (default) or tenor
GIF_PROVIDER=giphy
# GIF_API_KEY: optional; Giphy falls back to its public beta key
GIF_API_KEY=
# GIF_RATING: g, pg, pg-13, r (default pg-13)
GIF_RATING=pg-13
EOF
else
    echo "==> Keeping existing $ENV_FILE"
fi

update_env() {
    local key="$1" val="${2:-}"
    [ -n "$val" ] || return 0
    if grep -q "^${key}=" "$ENV_FILE"; then
        sed -i "s|^${key}=.*|${key}=${val}|" "$ENV_FILE"
    else
        printf '%s=%s\n' "$key" "$val" >>"$ENV_FILE"
    fi
}
update_env DEVZAT_TOKEN "${DEVZAT_TOKEN:-}"
update_env GIF_PROVIDER "${GIF_PROVIDER:-}"
update_env GIF_API_KEY "${GIF_API_KEY:-}"
update_env GIF_RATING "${GIF_RATING:-}"
chmod 600 "$ENV_FILE"

if ! grep -q '^DEVZAT_TOKEN=.\+' "$ENV_FILE"; then
    echo
    echo "WARNING: DEVZAT_TOKEN is empty — the plugin will fail to start."
fi

echo "==> Starting service"
systemctl daemon-reload
systemctl enable devzat-gif
systemctl restart devzat-gif

echo
systemctl --no-pager --full status devzat-gif || true
