# TLS Certificates

## Local Certificate Authority

Private key generation:

`openssl genrsa -des3 -out myCA.key 2048`

Passphrase: Kia has this; not storing in repo

Root certificate:

`openssl req -x509 -new -nodes -key CA.key -sha256 -days 365 -out CA.pem`

```
Country Name (2 letter code) [AU]:US
State or Province Name (full name) [Some-State]:US
Locality Name (eg, city) []:Austin
Organization Name (eg, company) [Internet Widgits Pty Ltd]:TCP Load Balancer
Organizational Unit Name (eg, section) []:1A
Common Name (e.g. server FQDN or YOUR name) []:TCP CA
Email Address []:kfarhang0@gmail.com
```

## Server Cert

Private key generation:

`openssl genrsa -out localhost.key 2048`

CSR generation:

`openssl req -new -key localhost.key -out localhost.csr`

```
Country Name (2 letter code) [AU]:US
State or Province Name (full name) [Some-State]:TX
Locality Name (eg, city) []:Austin
Organization Name (eg, company) [Internet Widgits Pty Ltd]:TCP Load Balancer
Organizational Unit Name (eg, section) []:Server
Common Name (e.g. server FQDN or YOUR name) []:localhost
Email Address []:server@example.com

Please enter the following 'extra' attributes
to be sent with your certificate request
A challenge password []:foodfight
An optional company name []:

```

Generation of signed cert:

`openssl x509 -req -in localhost.csr -CA ../ca/CA.pem -CAkey ../ca/CA.key -CAcreateserial -out localhost.crt -days 365 -sha256 -extfile <(printf "subjectAltName=DNS:localhost")`

Note: openssl's `x509` doesn't have an `-addext` like `req` does, so we just use a subshell to get around creating a whole file for the SAN

## Client Certs

### Admin

Private key generation:

`openssl genrsa -out client.key 2048`

CSR generation:

`openssl req -new -key client.key -out admin.csr`

```
Country Name (2 letter code) [AU]:US
State or Province Name (full name) [Some-State]:TX
Locality Name (eg, city) []:Austin
Organization Name (eg, company) [Internet Widgits Pty Ltd]:Client
Organizational Unit Name (eg, section) []:Admins
Common Name (e.g. server FQDN or YOUR name) []:localhost
Email Address []:admin@example.com

Please enter the following 'extra' attributes
to be sent with your certificate request
A challenge password []:client
An optional company name []:
```

Generation of signed cert:

`openssl x509 -req -in admin.csr -CA ../ca/CA.pem -CAkey ../ca/CA.key -CAcreateserial -out admin.crt -days 365 -sha256 -extfile <(printf "subjectAltName=DNS:localhost,email:admin@example.com")`

Note: openssl's `x509` doesn't have an `-addext` like `req` does, so we just use a subshell to get around creating a whole file for the SAN

### User

### Self signed

Generation of self-signed cert (AKA not using the CA the server trusts):

```
openssl req -newkey rsa:2048 -new -nodes -x509 -days 365 -out self-signed.pem -keyout self-signed-key.pem -addext "subjectAltName=DNS:localhost,email:self-signed@example.com"
```

```
Country Name (2 letter code) [AU]:US
State or Province Name (full name) [Some-State]:TX
Locality Name (eg, city) []:Austin
Organization Name (eg, company) [Internet Widgits Pty Ltd]:Self Signed Inc
Organizational Unit Name (eg, section) []:OU 
Common Name (e.g. server FQDN or YOUR name) []:localhost
Email Address []:self-signed@example.com
```
