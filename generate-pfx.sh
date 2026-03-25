#!/usr/bin/env bash
set -euo pipefail

CERT_NAME="${CERT_NAME:-mdx-codesign}"
SUBJECT="${SUBJECT:-/CN=github arimatakao}"
DAYS="${DAYS:-1825}"
PFX_PASSWORD="${PFX_PASSWORD:-}"

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

if [ -z "$PFX_PASSWORD" ]; then
  read -r -s -p "Enter password for ${CERT_NAME}.pfx: " PFX_PASSWORD
  echo
fi

KEY_FILE="${CERT_NAME}.key"
CRT_FILE="${CERT_NAME}.crt"
PFX_FILE="${CERT_NAME}.pfx"
SHA1_FINGERPRINT=""

openssl genrsa -out "$KEY_FILE" 4096

openssl req -new -x509 \
  -key "$KEY_FILE" \
  -out "$CRT_FILE" \
  -days "$DAYS" \
  -sha256 \
  -subj "$SUBJECT" \
  -addext "basicConstraints=critical,CA:FALSE" \
  -addext "keyUsage=digitalSignature" \
  -addext "extendedKeyUsage=codeSigning"

openssl pkcs12 -export \
  -out "$PFX_FILE" \
  -inkey "$KEY_FILE" \
  -in "$CRT_FILE" \
  -passout "pass:${PFX_PASSWORD}"

SHA1_FINGERPRINT="$(openssl x509 -in "$CRT_FILE" -noout -fingerprint -sha1 | cut -d= -f2 | tr -d ':')"

echo "Created files:"
echo "  ${KEY_FILE}"
echo "  ${CRT_FILE}"
echo "  ${PFX_FILE}"
echo
echo "GitHub Actions secrets:"
printf "  WINDOWS_CERT_PASSWORD_B64: "
printf '%s' "${PFX_PASSWORD}" | "${BASE64_CMD[@]}"
echo
printf "  WINDOWS_CERT_PFX_B64: "
"${BASE64_CMD[@]}" < "$PFX_FILE"
echo
echo "  WINDOWS_CERT_SHA1: ${SHA1_FINGERPRINT}"
