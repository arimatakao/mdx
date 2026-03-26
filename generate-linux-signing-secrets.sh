#!/usr/bin/env bash
set -euo pipefail

SIGNING_NAME="${SIGNING_NAME:-arimatakao}"
SIGNING_EMAIL="${SIGNING_EMAIL:-}"
GPG_PASSPHRASE="${GPG_PASSPHRASE:-}"
APK_PASSPHRASE="${APK_PASSPHRASE:-}"
GPG_VALIDITY="${GPG_VALIDITY:-5y}"
APK_KEY_NAME="${APK_KEY_NAME:-mdx-signing}"
OUTPUT_DIR="${OUTPUT_DIR:-linux-signing}"

if ! command -v gpg >/dev/null 2>&1; then
  echo "gpg not found in PATH" >&2
  exit 1
fi

if ! command -v openssl >/dev/null 2>&1; then
  echo "openssl not found in PATH" >&2
  exit 1
fi

if command -v base64 >/dev/null 2>&1; then
  BASE64_CMD=(base64 -w 0)
elif command -v openssl >/dev/null 2>&1; then
  BASE64_CMD=(openssl base64 -A)
else
  echo "base64 encoder not found" >&2
  exit 1
fi

if [ -z "$SIGNING_EMAIL" ]; then
  read -r -p "Enter signer email for GPG/DEB metadata: " SIGNING_EMAIL
fi

if [ -z "$GPG_PASSPHRASE" ]; then
  read -r -s -p "Enter passphrase for Linux GPG signing key: " GPG_PASSPHRASE
  echo
fi

if [ -z "$APK_PASSPHRASE" ]; then
  read -r -s -p "Enter passphrase for Alpine APK signing key: " APK_PASSPHRASE
  echo
fi

SIGNER_UID="${SIGNING_NAME} <${SIGNING_EMAIL}>"
GPG_HOME="${OUTPUT_DIR}/gnupg"
GPG_KEY_FILE="${OUTPUT_DIR}/linux-signing.gpg.asc"
APK_KEY_FILE="${OUTPUT_DIR}/${APK_KEY_NAME}.rsa"
APK_PUB_FILE="${OUTPUT_DIR}/${APK_KEY_NAME}.rsa.pub"

rm -rf "$OUTPUT_DIR"
mkdir -p "$OUTPUT_DIR"
chmod 700 "$OUTPUT_DIR"
mkdir -p "$GPG_HOME"
chmod 700 "$GPG_HOME"

cat > "${OUTPUT_DIR}/gpg-batch.conf" <<EOF
Key-Type: RSA
Key-Length: 4096
Key-Usage: sign
Name-Real: ${SIGNING_NAME}
Name-Email: ${SIGNING_EMAIL}
Expire-Date: ${GPG_VALIDITY}
Passphrase: ${GPG_PASSPHRASE}
%commit
EOF

gpg --batch --homedir "$GPG_HOME" --pinentry-mode loopback --generate-key "${OUTPUT_DIR}/gpg-batch.conf"

GPG_KEY_ID="$(
  gpg --batch --homedir "$GPG_HOME" --with-colons --list-secret-keys "$SIGNER_UID" \
    | awk -F: '$1 == "sec" { print $5; exit }'
)"

if [ -z "$GPG_KEY_ID" ]; then
  echo "Failed to resolve generated GPG key id" >&2
  exit 1
fi

gpg --batch --homedir "$GPG_HOME" --pinentry-mode loopback --passphrase "$GPG_PASSPHRASE" \
  --armor --export-secret-keys "$GPG_KEY_ID" > "$GPG_KEY_FILE"

openssl genrsa -aes256 -passout "pass:${APK_PASSPHRASE}" -out "$APK_KEY_FILE" 4096
openssl rsa -in "$APK_KEY_FILE" -passin "pass:${APK_PASSPHRASE}" -pubout -out "$APK_PUB_FILE" >/dev/null 2>&1

echo "Created files:"
echo "  ${GPG_KEY_FILE}"
echo "  ${APK_KEY_FILE}"
echo "  ${APK_PUB_FILE}"
echo
echo "GitHub Actions secrets:"
printf "  LINUX_GPG_PRIVATE_KEY_B64: "
"${BASE64_CMD[@]}" < "$GPG_KEY_FILE"
echo
printf "  LINUX_GPG_PASSPHRASE_B64: "
printf '%s' "${GPG_PASSPHRASE}" | "${BASE64_CMD[@]}"
echo
echo "  LINUX_GPG_KEY_ID: ${GPG_KEY_ID}"
echo "  LINUX_DEB_SIGNER: ${SIGNER_UID}"
printf "  LINUX_APK_PRIVATE_KEY_B64: "
"${BASE64_CMD[@]}" < "$APK_KEY_FILE"
echo
printf "  LINUX_APK_PASSPHRASE_B64: "
printf '%s' "${APK_PASSPHRASE}" | "${BASE64_CMD[@]}"
echo
echo "  LINUX_APK_KEY_NAME: ${APK_KEY_NAME}"
