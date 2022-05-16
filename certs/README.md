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

