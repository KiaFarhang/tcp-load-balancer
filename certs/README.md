# TLS Certificates

## Server Cert

Generated with the following command: 

```
openssl req -newkey rsa:2048 -new -nodes -x509 -days 365 -out cert.pem -keyout key.pem
```

Written to `certs/server`

```
Country Name (2 letter code) [AU]:US
State or Province Name (full name) [Some-State]:TX
Locality Name (eg, city) []:Austin
Organization Name (eg, company) [Internet Widgits Pty Ltd]:Server LLC
Organizational Unit Name (eg, section) []:My Unit
Common Name (e.g. server FQDN or YOUR name) []:Server
Email Address []:server@example.com
```

## Client Certs

## Admin

```
openssl req -newkey rsa:2048 -new -nodes -x509 -days 365 -out admin-cert.pem -keyout admin-key.pem -addext "subjectAltName=DNS:localhost,email:admin@example.com"
```

```
Country Name (2 letter code) [AU]:US
State or Province Name (full name) [Some-State]:TX
Locality Name (eg, city) []:Austin
Organization Name (eg, company) [Internet Widgits Pty Ltd]:Client LLC
Organizational Unit Name (eg, section) []:Admin Unit
Common Name (e.g. server FQDN or YOUR name) []:Admin
Email Address []:admin@example.com
```

## User

```
openssl req -newkey rsa:2048 -new -nodes -x509 -days 365 -out user-cert.pem -keyout user-key.pem -addext "subjectAltName=DNS:localhost,email:user@example.com"
```

```
Country Name (2 letter code) [AU]:US
State or Province Name (full name) [Some-State]:TX
Locality Name (eg, city) []:Austin
Organization Name (eg, company) [Internet Widgits Pty Ltd]:Client LLC
Organizational Unit Name (eg, section) []:User Unit
Common Name (e.g. server FQDN or YOUR name) []:User
Email Address []:user@example.com
```