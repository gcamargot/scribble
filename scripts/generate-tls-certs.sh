#!/bin/bash
# Generate TLS certificates for mTLS between Node.js proxy and Go backend
# Run this script once to generate certs, then apply k8s/tls-secrets.yaml

set -e

CERT_DIR="./certs"
DAYS_VALID=365

echo "Creating certificate directory..."
mkdir -p "$CERT_DIR"

# Generate CA (Certificate Authority)
echo "Generating CA certificate..."
openssl genrsa -out "$CERT_DIR/ca.key" 4096
openssl req -new -x509 -days $DAYS_VALID -key "$CERT_DIR/ca.key" \
    -out "$CERT_DIR/ca.crt" \
    -subj "/C=US/ST=State/L=City/O=Scribble/OU=Internal/CN=Scribble Internal CA"

# Generate Go backend server certificate
echo "Generating Go backend server certificate..."
openssl genrsa -out "$CERT_DIR/go-backend.key" 2048

# Create config for SAN (Subject Alternative Names)
cat > "$CERT_DIR/go-backend.cnf" << EOF
[req]
distinguished_name = req_distinguished_name
req_extensions = v3_req
prompt = no

[req_distinguished_name]
C = US
ST = State
L = City
O = Scribble
OU = Backend
CN = go-backend

[v3_req]
keyUsage = critical, digitalSignature, keyEncipherment
extendedKeyUsage = serverAuth
subjectAltName = @alt_names

[alt_names]
DNS.1 = go-backend
DNS.2 = go-backend.default.svc.cluster.local
DNS.3 = localhost
IP.1 = 127.0.0.1
EOF

openssl req -new -key "$CERT_DIR/go-backend.key" \
    -out "$CERT_DIR/go-backend.csr" \
    -config "$CERT_DIR/go-backend.cnf"

openssl x509 -req -days $DAYS_VALID \
    -in "$CERT_DIR/go-backend.csr" \
    -CA "$CERT_DIR/ca.crt" \
    -CAkey "$CERT_DIR/ca.key" \
    -CAcreateserial \
    -out "$CERT_DIR/go-backend.crt" \
    -extensions v3_req \
    -extfile "$CERT_DIR/go-backend.cnf"

# Generate Node.js proxy client certificate
echo "Generating Node.js proxy client certificate..."
openssl genrsa -out "$CERT_DIR/nodejs-proxy.key" 2048

cat > "$CERT_DIR/nodejs-proxy.cnf" << EOF
[req]
distinguished_name = req_distinguished_name
req_extensions = v3_req
prompt = no

[req_distinguished_name]
C = US
ST = State
L = City
O = Scribble
OU = Proxy
CN = nodejs-proxy

[v3_req]
keyUsage = critical, digitalSignature, keyEncipherment
extendedKeyUsage = clientAuth
subjectAltName = @alt_names

[alt_names]
DNS.1 = nodejs-proxy
DNS.2 = nodejs-proxy.default.svc.cluster.local
DNS.3 = localhost
EOF

openssl req -new -key "$CERT_DIR/nodejs-proxy.key" \
    -out "$CERT_DIR/nodejs-proxy.csr" \
    -config "$CERT_DIR/nodejs-proxy.cnf"

openssl x509 -req -days $DAYS_VALID \
    -in "$CERT_DIR/nodejs-proxy.csr" \
    -CA "$CERT_DIR/ca.crt" \
    -CAkey "$CERT_DIR/ca.key" \
    -CAcreateserial \
    -out "$CERT_DIR/nodejs-proxy.crt" \
    -extensions v3_req \
    -extfile "$CERT_DIR/nodejs-proxy.cnf"

# Clean up CSR and config files
rm -f "$CERT_DIR"/*.csr "$CERT_DIR"/*.cnf "$CERT_DIR"/*.srl

echo ""
echo "Certificates generated successfully in $CERT_DIR/"
echo ""
echo "Files created:"
echo "  - ca.crt, ca.key (Certificate Authority)"
echo "  - go-backend.crt, go-backend.key (Go backend server)"
echo "  - nodejs-proxy.crt, nodejs-proxy.key (Node.js client)"
echo ""
echo "To create Kubernetes secrets, run:"
echo "  kubectl create secret generic scribble-tls-ca --from-file=ca.crt=$CERT_DIR/ca.crt"
echo "  kubectl create secret tls scribble-go-backend-tls --cert=$CERT_DIR/go-backend.crt --key=$CERT_DIR/go-backend.key"
echo "  kubectl create secret tls scribble-nodejs-proxy-tls --cert=$CERT_DIR/nodejs-proxy.crt --key=$CERT_DIR/nodejs-proxy.key"
