#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

rm -rf data/grafana-aggregator

mkdir -p data/grafana-aggregator

openssl req -nodes -new -x509 -keyout data/grafana-aggregator/ca.key -out data/grafana-aggregator/ca.crt
openssl req -out data/grafana-aggregator/client.csr -new -newkey rsa:4096 -nodes -keyout data/grafana-aggregator/client.key \
  -subj "/CN=development/O=system:masters" \
  -addext "extendedKeyUsage = clientAuth"
openssl x509 -req -days 365 -in data/grafana-aggregator/client.csr -CA data/grafana-aggregator/ca.crt -CAkey data/grafana-aggregator/ca.key \
  -set_serial 01 \
  -sha256 -out data/grafana-aggregator/client.crt \
  -copy_extensions=copyall

openssl req -out data/grafana-aggregator/server.csr -new -newkey rsa:4096 -nodes -keyout data/grafana-aggregator/server.key \
  -subj "/CN=localhost/O=aggregated" \
  -addext "subjectAltName = DNS:v0alpha1.example.grafana.app.default.svc,DNS:localhost" \
  -addext "extendedKeyUsage = serverAuth, clientAuth"
openssl x509 -req -days 365 -in data/grafana-aggregator/server.csr -CA data/grafana-aggregator/ca.crt -CAkey data/grafana-aggregator/ca.key \
  -set_serial 02 \
  -sha256 -out data/grafana-aggregator/server.crt \
  -copy_extensions=copyall
