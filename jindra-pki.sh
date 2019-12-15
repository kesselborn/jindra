#!/bin/sh -ex
: ${CA_SUBJ:=jindra.io}
: ${SERVICE_NAME:=webhook-service}
: ${NAMESPACE:=jindra}
: ${PKI_DIR:=.pki}
: ${DAYS:=3650}

mkdir -p ${PKI_DIR}

CA_CERT=${PKI_DIR}/ca.crt
CA_KEY=${PKI_DIR}/ca.key

SERVER_CERT=${PKI_DIR}/server.crt
SERVER_KEY=${PKI_DIR}/server.key
SERVER_CSR=${SERVER_CERT}.csr

if [ ! -e ${CA_CERT} -o ! -e ${CA_KEY} ]; then
  openssl req -x509 -sha256 -subj "/CN=${CA_SUBJ}" -nodes -newkey rsa:4096 -keyout ${CA_KEY} -out ${CA_CERT} -days ${DAYS}
  chmod 0600 ${CA_KEY}
fi

openssl genrsa -out ${SERVER_KEY} 2048
chmod 0600 ${SERVER_KEY}

cat <<EOF >>${SERVER_CSR}.conf
[req]
req_extensions = v3_req
distinguished_name = req_distinguished_name
[req_distinguished_name]
[ v3_req ]
basicConstraints = CA:FALSE
keyUsage = nonRepudiation, digitalSignature, keyEncipherment
extendedKeyUsage = serverAuth
subjectAltName = @alt_names
[alt_names]
DNS.1 = ${SERVICE_NAME}
DNS.2 = ${SERVICE_NAME}.${NAMESPACE}
DNS.3 = ${SERVICE_NAME}.${NAMESPACE}.svc
EOF

openssl req \
  -config ${SERVER_CSR}.conf \
  -extensions v3_req \
  -key ${SERVER_KEY} \
  -new \
  -out ${SERVER_CSR} \
  -subj "/CN=${SERVICE_NAME}.${NAMESPACE}.svc"

openssl x509 \
  -CA ${CA_CERT} \
  -CAcreateserial \
  -CAkey ${CA_KEY} \
  -CAserial ${PKI_DIR}/ca.seq \
  -days ${DAYS} \
  -extensions v3_req \
  -extfile ${SERVER_CSR}.conf \
  -in ${SERVER_CSR} \
  -out ${SERVER_CERT} \
  -req \
  -sha256

cat <<EOF >config/default/admission_controllers_ca_bundle_patch.yaml
# This patch adds the jindra-pki.sh create ca bundle to the admission webhooks
apiVersion: admissionregistration.k8s.io/v1beta1
kind: MutatingWebhookConfiguration
metadata:
  name: mutating-webhook-configuration
webhooks:
- clientConfig:
    caBundle: $(cat ${CA_CERT} | base64)
  name: defaulter.jindra.io
---
apiVersion: admissionregistration.k8s.io/v1beta1
kind: ValidatingWebhookConfiguration
metadata:
  name: validating-webhook-configuration
webhooks:
- clientConfig:
    caBundle: $(cat ${CA_CERT} | base64)
  name: validator.jindra.io
EOF

kubectl -n ${NAMESPACE} create --dry-run -oyaml secret tls webhook-server-cert --cert .pki/server.crt --key .pki/server.key
