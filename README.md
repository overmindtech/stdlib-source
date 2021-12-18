# stdlib

The source includes a standard library of item types that should always be present. Many of these item types exist in the `global` context and serve to link together items in other contexts (e.g. ip addresses)

<!-- @import "[TOC]" {cmd="toc" depthFrom=2 depthTo=6 orderedList=false} -->

<!-- code_chunk_output -->

- [Sources](#sources)
  - [Networksocket](#networksocket)
    - [`Get`](#get)
    - [`Find`](#find)
  - [Certificate](#certificate)
    - [`Get`](#get-1)
    - [`Find`](#find-1)
    - [`Search`](#search)
  - [DNS](#dns)
    - [`Get`](#get-2)
    - [`Find`](#find-2)
  - [HTTP](#http)
    - [`Get`](#get-3)
    - [`Find`](#find-3)
  - [IP](#ip)
    - [`Get`](#get-4)
    - [`Find`](#find-4)
- [Config](#config)
  - [`srcman` config](#srcman-config)
  - [Health Check](#health-check)
- [Development](#development)
  - [Running Locally](#running-locally)
  - [Testing](#testing)
  - [Packaging](#packaging)

<!-- /code_chunk_output -->

## Sources

### Networksocket

This returns information about network sockets. A network socket is a combination of a host (IP or DNS name) and a port. Note that this doesn't enrich the data at all and is mostly for the purpose of linking.

Example item:

```json
{
    "type": "networksocket",
    "uniqueAttribute": "socket",
    "attributes": {
        "attrStruct": {
            "host": "www.google.com",
            "port": "443",
            "socket": "www.google.com:443"
        }
    },
    "context": "global",
    "linkedItemRequests": [
        {
            "type": "dns",
            "query": "www.google.com",
            "context": "global"
        }
    ]
}
```

**Note:** I'm aware that a DNS name might not map to a single IP and therefore isn't technically a network socket. For example a DNS name might return multiple A records and therefore could refer to many different sockets. From a logical perspective though it's usually safe to think about it as being a single socket, and the value that it provides as a link probably outweights any confusion caused. If you disagree or can think of a better approach, please raise an issue.

#### `Get`

Returns socket information for a given ip:port combo

**Query format:** An IP and port in the format `ip:port`.

#### `Find`

This method is not implemented. Use `Get` or `Search` instead

#### `Search`

Returns socket information for a given host:port combo, resolving any DNS queries as required to get a list of `networksockets`

**Query format:** A host and port in the format `host:port`. Host can be an IP or a DNS name, DNS names will be resolved and may result in many sockets being returned (like in the case of a DNS query resolving to many A records)


### Certificate

The source gathers information about certificates by parsing them. Note that it doesn't make connections to the servers hosting teh certificates, or read certificates from disk. It relies on other sources that come across certificates (e.g. as part of a HTTPS connection), passing them on to this source via a linked item e.g.

```go
var certs []string

// Loop over peer certificates form the connection and encode to PEM
for _, cert := range tlsState.PeerCertificates {
  block := pem.Block{
    Type:  "CERTIFICATE",
    Bytes: cert.Raw,
  }

  certs = append(certs, string(pem.EncodeToMemory(&block)))
}

// Create a linked item request with the PEM bundle
request := sdp.ItemRequest{
  Type:   "certificate",
  Method: sdp.RequestMethod_SEARCH,
  Query:  strings.Join(certs, "\n"),
})
```

Example item:

```json
{
    "type": "certificate",
    "uniqueAttribute": "subject",
    "attributes": {
        "attrStruct": {
            "CRLDistributionPoints": [
                "http://crl3.digicert.com/DigiCertHighAssuranceEVRootCA.crl",
                "http://crl4.digicert.com/DigiCertHighAssuranceEVRootCA.crl"
            ],
            "authorityKeyIdentifier": "B1:3E:C3:69:03:F8:BF:47:01:D4:98:26:1A:08:02:EF:63:64:2B:C3",
            "basicConstraints": {
                "CA": true,
                "pathLen": 0
            },
            "extendedKeyUsage": [
                "TLS Web Server Authentication",
                "TLS Web Client Authentication",
                "Code Signing",
                "E-mail Protection",
                "Time Stamping"
            ],
            "issuer": "CN=DigiCert High Assurance EV Root CA,OU=www.digicert.com,O=DigiCert Inc,C=US",
            "issuingCertificateURL": "http://www.digicert.com/CACerts/DigiCertHighAssuranceEVRootCA.crt",
            "keyUsage": [
                "Digital Signature",
                "Certificate Sign",
                "CRL Sign"
            ],
            "notAfter": "2021-11-10 00:00:00 +0000 UTC",
            "notBefore": "2007-11-09 12:00:00 +0000 UTC",
            "ocspServer": "http://ocsp.digicert.com",
            "policyIdentifiers": [
                "2.16.840.1.114412.2.1"
            ],
            "publicKeyAlgorithm": "RSA",
            "serialNumber": "03:37:B9:28:34:7C:60:A6:AE:C5:AD:B1:21:7F:38:60",
            "signature": "4C:7A:17:87:28:5D:17:BC:B2:32:73:BF:CD:2E:F5:58:31:1D:F0:B1:71:54:9C:D6:9B:67:93:DB:2F:03:3E:16:6F:1E:03:C9:53:84:A3:56:60:1E:78:94:1B:A2:A8:6F:A3:A4:8B:52:91:D7:DD:5C:95:BB:EF:B5:16:49:E9:A5:42:4F:34:F2:47:FF:AE:81:7F:13:54:B7:20:C4:70:15:CB:81:0A:81:CB:74:57:DC:9C:DF:24:A4:29:0C:18:F0:1C:E4:AE:07:33:EC:F1:49:3E:55:CF:6E:4F:0D:54:7B:D3:C9:E8:15:48:D4:C5:BB:DC:35:1C:77:45:07:48:45:85:BD:D7:7E:53:B8:C0:16:D9:95:CD:8B:8D:7D:C9:60:4F:D1:A2:9B:E3:D0:30:D6:B4:73:36:E6:D2:F9:03:B2:E3:A4:F5:E5:B8:3E:04:49:00:BA:2E:A6:4A:72:83:72:9D:F7:0B:8C:A9:89:E7:B3:D7:64:1F:D6:E3:60:CB:03:C4:DC:88:E9:9D:25:01:00:71:CB:03:B4:29:60:25:8F:F9:46:D1:7B:71:AE:CD:53:12:5B:84:8E:C2:0F:C7:ED:93:19:D9:C9:FA:8F:58:34:76:32:2F:AE:E1:50:14:61:D4:A8:58:A3:C8:30:13:23:EF:C6:25:8C:36:8F:1C:80",
            "signatureAlgorithm": "SHA1-RSA",
            "subject": "CN=DigiCert High Assurance EV CA-1,OU=www.digicert.com,O=DigiCert Inc,C=US",
            "subjectKeyIdentifier": "4C:58:CB:25:F0:41:4F:52:F4:28:C8:81:43:9B:A6:A8:A0:E6:92:E5",
            "version": 3
        }
    },
    "context": "global",
    "linkedItemRequests": [
        {
            "type": "certificate",
            "query": "CN=DigiCert High Assurance EV Root CA,OU=www.digicert.com,O=DigiCert Inc,C=US"
        }
    ]
}
```

#### `Get`

This method is not implemented. Use `Search` instead

#### `Find`

This method is not implemented. Use `Search` instead

#### `Search`

The `Search` method will parse the PEM encoded certificate or bundle and return all certificates found.

**Query format:** A full certificate, or certificate bundle in PEM encoded format e.g.

```
-----BEGIN CERTIFICATE-----
MIIDxTCCAq2gAwIBAgIQAqxcJmoLQJuPC3nyrkYldzANBgkqhkiG9w0BAQUFADBs
MQswCQYDVQQGEwJVUzEVMBMGA1UEChMMRGlnaUNlcnQgSW5jMRkwFwYDVQQLExB3
d3cuZGlnaWNlcnQuY29tMSswKQYDVQQDEyJEaWdpQ2VydCBIaWdoIEFzc3VyYW5j
ZSBFViBSb290IENBMB4XDTA2MTExMDAwMDAwMFoXDTMxMTExMDAwMDAwMFowbDEL
MAkGA1UEBhMCVVMxFTATBgNVBAoTDERpZ2lDZXJ0IEluYzEZMBcGA1UECxMQd3d3
LmRpZ2ljZXJ0LmNvbTErMCkGA1UEAxMiRGlnaUNlcnQgSGlnaCBBc3N1cmFuY2Ug
RVYgUm9vdCBDQTCCASIwDQYJKoZIhvcNAQEBBQADggEPADCCAQoCggEBAMbM5XPm
+9S75S0tMqbf5YE/yc0lSbZxKsPVlDRnogocsF9ppkCxxLeyj9CYpKlBWTrT3JTW
PNt0OKRKzE0lgvdKpVMSOO7zSW1xkX5jtqumX8OkhPhPYlG++MXs2ziS4wblCJEM
xChBVfvLWokVfnHoNb9Ncgk9vjo4UFt3MRuNs8ckRZqnrG0AFFoEt7oT61EKmEFB
Ik5lYYeBQVCmeVyJ3hlKV9Uu5l0cUyx+mM0aBhakaHPQNAQTXKFx01p8VdteZOE3
hzBWBOURtCmAEvF5OYiiAhF8J2a3iLd48soKqDirCmTCv2ZdlYTBoSUeh10aUAsg
EsxBu24LUTi4S8sCAwEAAaNjMGEwDgYDVR0PAQH/BAQDAgGGMA8GA1UdEwEB/wQF
MAMBAf8wHQYDVR0OBBYEFLE+w2kD+L9HAdSYJhoIAu9jZCvDMB8GA1UdIwQYMBaA
FLE+w2kD+L9HAdSYJhoIAu9jZCvDMA0GCSqGSIb3DQEBBQUAA4IBAQAcGgaX3Nec
nzyIZgYIVyHbIUf4KmeqvxgydkAQV8GK83rZEWWONfqe/EW1ntlMMUu4kehDLI6z
eM7b41N5cdblIZQB2lWHmiRk9opmzN6cN82oNLFpmyPInngiK3BD41VHMWEZ71jF
hS9OMPagMRYjyOfiZRYzy78aG6A9+MpeizGLYAiJLQwGXFK3xPkKmNEVX58Svnw2
Yzi9RKR/5CYrCsSXaQ3pjOLAEFe4yHYSkVXySGnYvCoCWw9E1CAx2/S6cCZdkGCe
vEsXCS+0yx5DaMkHJ8HSXPfqIbloEpw8nL+e/IBcm2PN7EeqJSdnoDfzAIJ9VNep
+OkuE6N36B9K
-----END CERTIFICATE-----
```

### DNS

The source makes DNS queries, returning the outcome.

Example item:

```json
{
    "type": "dns",
    "uniqueAttribute": "name",
    "attributes": {
        "attrStruct": {
            "ips": [
                "2606:4700:4700::1001",
                "2606:4700:4700::1111",
                "1.0.0.1",
                "1.1.1.1"
            ],
            "name": "one.one.one.one"
        }
    },
    "context": "global",
    "linkedItemRequests": [
        {
            "type": "ip",
            "query": "2606:4700:4700::1001",
            "context": "global"
        },
        {
            "type": "ip",
            "query": "2606:4700:4700::1111",
            "context": "global"
        },
        {
            "type": "ip",
            "query": "1.0.0.1",
            "context": "global"
        },
        {
            "type": "ip",
            "query": "1.1.1.1",
            "context": "global"
        }
    ]
}
```

#### `Get`

Looks up the given host in DNS, returing the details if found.

**Query format:** A hostname e.g. `www.google.com`

#### `Find`

This method is not implemented.

### HTTP

The source makes a HTTP `HEAD` request and returns information abotu what the server would have returned if a `GET` request was made. This is useful for ensuring that endpoints actually exist, what their status is, and what certificates they are using. Note that as long as the server returns an error that is valid HTML (like a 500) this source will still return an item.

Example item for a 200 response:

```json
{
    "type": "http",
    "uniqueAttribute": "name",
    "attributes": {
        "attrStruct": {
            "headers": {
                "Alt-Svc": "h3=\":443\"; ma=2592000,h3-29=\":443\"; ma=2592000,h3-Q050=\":443\"; ma=2592000,h3-Q046=\":443\"; ma=2592000,h3-Q043=\":443\"; ma=2592000,quic=\":443\"; ma=2592000; v=\"46,43\"",
                "Cache-Control": "private",
                "Content-Type": "text/html; charset=ISO-8859-1",
                "Date": "Wed, 01 Dec 2021 10:45:16 GMT",
                "Expires": "Wed, 01 Dec 2021 10:45:16 GMT",
                "P3p": "CP=\"This is not a P3P policy! See g.co/p3phelp for more info.\"",
                "Server": "gws",
                "Set-Cookie": "CONSENT=PENDING+703; expires=Fri, 01-Dec-2023 10:45:16 GMT; path=/; domain=.google.com; Secure",
                "X-Frame-Options": "SAMEORIGIN",
                "X-Xss-Protection": "0"
            },
            "name": "https://www.google.com",
            "proto": "HTTP/1.1",
            "status": 200,
            "statusString": "200 OK",
            "tls": {
                "certificate": "www.google.com (SHA-1: C3:EE:B4:93:EB:C8:75:F9:66:66:54:FE:B1:F8:D0:42:A7:E3:A5:FE)",
                "serverName": "www.google.com",
                "version": "TLSv1.3"
            },
            "transferEncoding": []
        }
    },
    "context": "global",
    "linkedItemRequests": [
        {
            "type": "certificate",
            "method": 2,
            "query": "-----BEGIN CERTIFICATE-----\nMIIEhzCCA2+gAwIBAgIQFDlrjDZN65cKAAAAARlVRTANBgkqhkiG9w0BAQsFADBG\nMQswCQYDVQQGEwJVUzEiMCAGA1UEChMZR29vZ2xlIFRydXN0IFNlcnZpY2VzIExM\nQzETMBEGA1UEAxMKR1RTIENBIDFDMzAeFw0yMTExMDEwMzE5MzZaFw0yMjAxMjQw\nMzE5MzVaMBkxFzAVBgNVBAMTDnd3dy5nb29nbGUuY29tMFkwEwYHKoZIzj0CAQYI\nKoZIzj0DAQcDQgAEw0ArW187m3SoPF2ITzroDOpZGFDwQX30z1DaO+xXHxVnsGSr\nVdFuhi/lMUHfmreHB3PGaN/yWHFEyWRBVkT8lqOCAmcwggJjMA4GA1UdDwEB/wQE\nAwIHgDATBgNVHSUEDDAKBggrBgEFBQcDATAMBgNVHRMBAf8EAjAAMB0GA1UdDgQW\nBBSKwanauR7Xpe4rA2uF33rs/wV3OTAfBgNVHSMEGDAWgBSKdH+vhc3ulc09nNDi\nRhTzcTUdJzBqBggrBgEFBQcBAQReMFwwJwYIKwYBBQUHMAGGG2h0dHA6Ly9vY3Nw\nLnBraS5nb29nL2d0czFjMzAxBggrBgEFBQcwAoYlaHR0cDovL3BraS5nb29nL3Jl\ncG8vY2VydHMvZ3RzMWMzLmRlcjAZBgNVHREEEjAQgg53d3cuZ29vZ2xlLmNvbTAh\nBgNVHSAEGjAYMAgGBmeBDAECATAMBgorBgEEAdZ5AgUDMDwGA1UdHwQ1MDMwMaAv\noC2GK2h0dHA6Ly9jcmxzLnBraS5nb29nL2d0czFjMy9tb1ZEZklTaWEyay5jcmww\nggEEBgorBgEEAdZ5AgQCBIH1BIHyAPAAdgApeb7wnjk5IfBWc59jpXflvld9nGAK\n+PlNXSZcJV3HhAAAAXzZuV/aAAAEAwBHMEUCIDDDUZWrPK/zBkTtkwppuI8U1vNM\niwD+SYPgYfynhILDAiEAzMf/O+Iqax11PR2CxlNheePWOGD2fVTHsfZorvyiKXUA\ndgBByMqx3yJGShDGoToJQodeTjGLGwPr60vHaPCQYpYG9gAAAXzZuWAIAAAEAwBH\nMEUCIQDjwmMNrWMMfBI6lZPlNhvJvdEd0Ua0IYP2+1caUr+TCgIgW6YT4gfB5FGj\nTHTE/FNb+9Bp1yM+8IQoCUSXwYFAJOIwDQYJKoZIhvcNAQELBQADggEBAFbf3ZKB\nTDnm1byzVaADhOWM7SFSQYnGQA+WlSpO1o4M/F5g7lcrOzanMddgscIQ7oK96Qfh\nmFp3+1LH1HT7Os2Ugx7t0ZF5LU2ijf0k6o3hl3/nCfulEtoDthIpmOnmpJ0BjaJu\nbhxp515ZWum/PaGODp5Yhv+w/84l6R6Mm0IuVqfYc5wj9w3OT0v9aJfrc7h+st50\ncrOqNFbc4fZMXO5iNHe9jexYMs07LKVu9S31oiFb4lKIv9MFU3TEMTy8l2iSQd+c\nIEZj47cmPSPSMzDGacITAywn5ZZIxhHujwFixvYevbYwlrJOXC59B+nAk7wbg+1y\nJtbIOv3GD0OM1Nc=\n-----END CERTIFICATE-----\n\n-----BEGIN CERTIFICATE-----\nMIIFljCCA36gAwIBAgINAgO8U1lrNMcY9QFQZjANBgkqhkiG9w0BAQsFADBHMQsw\nCQYDVQQGEwJVUzEiMCAGA1UEChMZR29vZ2xlIFRydXN0IFNlcnZpY2VzIExMQzEU\nMBIGA1UEAxMLR1RTIFJvb3QgUjEwHhcNMjAwODEzMDAwMDQyWhcNMjcwOTMwMDAw\nMDQyWjBGMQswCQYDVQQGEwJVUzEiMCAGA1UEChMZR29vZ2xlIFRydXN0IFNlcnZp\nY2VzIExMQzETMBEGA1UEAxMKR1RTIENBIDFDMzCCASIwDQYJKoZIhvcNAQEBBQAD\nggEPADCCAQoCggEBAPWI3+dijB43+DdCkH9sh9D7ZYIl/ejLa6T/belaI+KZ9hzp\nkgOZE3wJCor6QtZeViSqejOEH9Hpabu5dOxXTGZok3c3VVP+ORBNtzS7XyV3NzsX\nlOo85Z3VvMO0Q+sup0fvsEQRY9i0QYXdQTBIkxu/t/bgRQIh4JZCF8/ZK2VWNAcm\nBA2o/X3KLu/qSHw3TT8An4Pf73WELnlXXPxXbhqW//yMmqaZviXZf5YsBvcRKgKA\ngOtjGDxQSYflispfGStZloEAoPtR28p3CwvJlk/vcEnHXG0g/Zm0tOLKLnf9LdwL\ntmsTDIwZKxeWmLnwi/agJ7u2441Rj72ux5uxiZ0CAwEAAaOCAYAwggF8MA4GA1Ud\nDwEB/wQEAwIBhjAdBgNVHSUEFjAUBggrBgEFBQcDAQYIKwYBBQUHAwIwEgYDVR0T\nAQH/BAgwBgEB/wIBADAdBgNVHQ4EFgQUinR/r4XN7pXNPZzQ4kYU83E1HScwHwYD\nVR0jBBgwFoAU5K8rJnEaK0gnhS9SZizv8IkTcT4waAYIKwYBBQUHAQEEXDBaMCYG\nCCsGAQUFBzABhhpodHRwOi8vb2NzcC5wa2kuZ29vZy9ndHNyMTAwBggrBgEFBQcw\nAoYkaHR0cDovL3BraS5nb29nL3JlcG8vY2VydHMvZ3RzcjEuZGVyMDQGA1UdHwQt\nMCswKaAnoCWGI2h0dHA6Ly9jcmwucGtpLmdvb2cvZ3RzcjEvZ3RzcjEuY3JsMFcG\nA1UdIARQME4wOAYKKwYBBAHWeQIFAzAqMCgGCCsGAQUFBwIBFhxodHRwczovL3Br\naS5nb29nL3JlcG9zaXRvcnkvMAgGBmeBDAECATAIBgZngQwBAgIwDQYJKoZIhvcN\nAQELBQADggIBAIl9rCBcDDy+mqhXlRu0rvqrpXJxtDaV/d9AEQNMwkYUuxQkq/BQ\ncSLbrcRuf8/xam/IgxvYzolfh2yHuKkMo5uhYpSTld9brmYZCwKWnvy15xBpPnrL\nRklfRuFBsdeYTWU0AIAaP0+fbH9JAIFTQaSSIYKCGvGjRFsqUBITTcFTNvNCCK9U\n+o53UxtkOCcXCb1YyRt8OS1b887U7ZfbFAO/CVMkH8IMBHmYJvJh8VNS/UKMG2Yr\nPxWhu//2m+OBmgEGcYk1KCTd4b3rGS3hSMs9WYNRtHTGnXzGsYZbr8w0xNPM1IER\nlQCh9BIiAfq0g3GvjLeMcySsN1PCAJA/Ef5c7TaUEDu9Ka7ixzpiO2xj2YC/WXGs\nYye5TBeg2vZzFb8q3o/zpWwygTMD0IZRcZk0upONXbVRWPeyk+gB9lm+cZv9TSjO\nz23HFtz30dZGm6fKa+l3D/2gthsjgx0QGtkJAITgRNOidSOzNIb2ILCkXhAd4FJG\nAJ2xDx8hcFH1mt0G/FX0Kw4zd8NLQsLxdxP8c4CU6x+7Nz/OAipmsHMdMqUybDKw\njuDEI/9bfU1lcKwrmz3O2+BtjjKAvpafkmO8l7tdufThcV4q5O8DIrGKZTqPwJNl\n1IXNDw9bg1kWRxYtnCQ6yICmJhSFm/Y3m6xv+cXDBlHz4n/FsRC6UfTd\n-----END CERTIFICATE-----\n\n-----BEGIN CERTIFICATE-----\nMIIFYjCCBEqgAwIBAgIQd70NbNs2+RrqIQ/E8FjTDTANBgkqhkiG9w0BAQsFADBX\nMQswCQYDVQQGEwJCRTEZMBcGA1UEChMQR2xvYmFsU2lnbiBudi1zYTEQMA4GA1UE\nCxMHUm9vdCBDQTEbMBkGA1UEAxMSR2xvYmFsU2lnbiBSb290IENBMB4XDTIwMDYx\nOTAwMDA0MloXDTI4MDEyODAwMDA0MlowRzELMAkGA1UEBhMCVVMxIjAgBgNVBAoT\nGUdvb2dsZSBUcnVzdCBTZXJ2aWNlcyBMTEMxFDASBgNVBAMTC0dUUyBSb290IFIx\nMIICIjANBgkqhkiG9w0BAQEFAAOCAg8AMIICCgKCAgEAthECix7joXebO9y/lD63\nladAPKH9gvl9MgaCcfb2jH/76Nu8ai6Xl6OMS/kr9rH5zoQdsfnFl97vufKj6bwS\niV6nqlKr+CMny6SxnGPb15l+8Ape62im9MZaRw1NEDPjTrETo8gYbEvs/AmQ351k\nKSUjB6G00j0uYODP0gmHu81I8E3CwnqIiru6z1kZ1q+PsAewnjHxgsHA3y6mbWwZ\nDrXYfiYaRQM9sHmklCitD38m5agI/pboPGiUU+6DOogrFZYJsuB6jC511pzrp1Zk\nj5ZPaK49l8KEj8C8QMALXL32h7M1bKwYUH+E4EzNktMg6TO8UpmvMrUpsyUqtEj5\ncuHKZPfmghCN6J3Cioj6OGaK/GP5Afl4/Xtcd/p2h/rs37EOeZVXtL0m79YB0esW\nCruOC7XFxYpVq9Os6pFLKcwZpDIlTirxZUTQAs6qzkm06p98g7BAe+dDq6dso499\niYH6TKX/1Y7DzkvgtdizjkXPdsDtQCv9Uw+wp9U7DbGKogPeMa3Md+pvez7W35Ei\nEua++tgy/BBjFFFy3l3WFpO9KWgz7zpm7AeKJt8T11dleCfeXkkUAKIAf5qoIbap\nsZWwpbkNFhHax2xIPEDgfg1azVY80ZcFuctL7TlLnMQ/0lUTbiSw1nH69MG6zO0b\n9f6BQdgAmD06yK56mDcYBZUCAwEAAaOCATgwggE0MA4GA1UdDwEB/wQEAwIBhjAP\nBgNVHRMBAf8EBTADAQH/MB0GA1UdDgQWBBTkrysmcRorSCeFL1JmLO/wiRNxPjAf\nBgNVHSMEGDAWgBRge2YaRQ2XyolQL30EzTSo//z9SzBgBggrBgEFBQcBAQRUMFIw\nJQYIKwYBBQUHMAGGGWh0dHA6Ly9vY3NwLnBraS5nb29nL2dzcjEwKQYIKwYBBQUH\nMAKGHWh0dHA6Ly9wa2kuZ29vZy9nc3IxL2dzcjEuY3J0MDIGA1UdHwQrMCkwJ6Al\noCOGIWh0dHA6Ly9jcmwucGtpLmdvb2cvZ3NyMS9nc3IxLmNybDA7BgNVHSAENDAy\nMAgGBmeBDAECATAIBgZngQwBAgIwDQYLKwYBBAHWeQIFAwIwDQYLKwYBBAHWeQIF\nAwMwDQYJKoZIhvcNAQELBQADggEBADSkHrEoo9C0dhemMXoh6dFSPsjbdBZBiLg9\nNR3t5P+T4Vxfq7vqfM/b5A3Ri1fyJm9bvhdGaJQ3b2t6yMAYN/olUazsaL+yyEn9\nWprKASOshIArAoyZl+tJaox118fessmXn1hIVw41oeQa1v1vg4Fv74zPl6/AhSrw\n9U5pCZEt4Wi4wStz6dTZ/CLANx8LZh1J7QJVj2fhMtfTJr9w4z30Z209fOU0iOMy\n+qduBmpvvYuR7hZL6Dupszfnw0Skfths18dG9ZKb59UhvmaSGZRVbNQpsg3BZlvi\nd0lIKO2d1xozclOzgjXPYovJJIultzkMu34qQb9Sz/yilrbCgj8=\n-----END CERTIFICATE-----\n"
        }
    ]
}
```

Example item for an error:

```json
{
    "type": "http",
    "uniqueAttribute": "name",
    "attributes": {
        "attrStruct": {
            "headers": {
                "Access-Control-Allow-Origin": "*",
                "Access-Control-Expose-Headers": "Link, Content-Range, Location, WWW-Authenticate, Proxy-Authenticate, Retry-After, Request-Context",
                "Alt-Svc": "h3=\":443\"; ma=86400, h3-29=\":443\"; ma=86400, h3-28=\":443\"; ma=86400, h3-27=\":443\"; ma=86400",
                "Cache-Control": "private",
                "Cf-Cache-Status": "DYNAMIC",
                "Cf-Ray": "6b6bc10eed1276ba-LHR",
                "Connection": "keep-alive",
                "Content-Length": "25",
                "Content-Type": "text/plain; charset=utf-8",
                "Date": "Wed, 01 Dec 2021 10:50:22 GMT",
                "Expect-Ct": "max-age=604800, report-uri=\"https://report-uri.cloudflare.com/cdn-cgi/beacon/expect-ct\"",
                "Nel": "{\"success_fraction\":0,\"report_to\":\"cf-nel\",\"max_age\":604800}",
                "Report-To": "{\"endpoints\":[{\"url\":\"https:\\/\\/a.nel.cloudflare.com\\/report\\/v3?s=xrN4i%2BwsS320rmj42Awzqcum%2B9DhZ1WgrXds9d2MxQZcAjnXCxscR3oZak0z0FhvEOlGz7kHX0Blmkll3xuyKy3WFWvE9maopwdZpvTxOsMdd%2F5SFuHW4%2FSphfBNchg3%2FiHO4Eqpi8BzUw%3D%3D\"}],\"group\":\"cf-nel\",\"max_age\":604800}",
                "Request-Context": "appId=cid-v1:7585021b-2db7-4da6-abff-2cf23005f0a9",
                "Server": "cloudflare",
                "Set-Cookie": "ARRAffinity=b3cb4dc0a78e3947c788498e968cc99ce12b1ba1bcd59699dde8236e5ceb6c1c;Path=/;HttpOnly;Domain=httpstat.us",
                "X-Aspnet-Version": "4.0.30319",
                "X-Aspnetmvc-Version": "5.1",
                "X-Powered-By": "ASP.NET"
            },
            "name": "https://httpstat.us/500",
            "proto": "HTTP/1.1",
            "status": 500,
            "statusString": "500 Internal Server Error",
            "tls": {
                "certificate": "sni.cloudflaressl.com (SHA-1: C8:96:C9:3D:51:CA:C8:20:07:19:D4:EE:AF:26:CF:FB:5F:D0:12:5A)",
                "serverName": "httpstat.us",
                "version": "TLSv1.3"
            },
            "transferEncoding": []
        }
    },
    "context": "global",
    "linkedItemRequests": [
        {
            "type": "certificate",
            "method": 2,
            "query": "-----BEGIN CERTIFICATE-----\nMIIFNDCCBNqgAwIBAgIQDmlmSoK/s/cogl6lbCRSazAKBggqhkjOPQQDAjBKMQsw\nCQYDVQQGEwJVUzEZMBcGA1UEChMQQ2xvdWRmbGFyZSwgSW5jLjEgMB4GA1UEAxMX\nQ2xvdWRmbGFyZSBJbmMgRUNDIENBLTMwHhcNMjEwNjI3MDAwMDAwWhcNMjIwNjI2\nMjM1OTU5WjB1MQswCQYDVQQGEwJVUzETMBEGA1UECBMKQ2FsaWZvcm5pYTEWMBQG\nA1UEBxMNU2FuIEZyYW5jaXNjbzEZMBcGA1UEChMQQ2xvdWRmbGFyZSwgSW5jLjEe\nMBwGA1UEAxMVc25pLmNsb3VkZmxhcmVzc2wuY29tMFkwEwYHKoZIzj0CAQYIKoZI\nzj0DAQcDQgAEUops6RVkUuUcYhX3BeAJhGwZh00oFbCwoKO+x6Y9wOZ8ApYJdqKA\nHXyLW7MLbPLS2fjdj8keEpnCc00GLS8Zq6OCA3UwggNxMB8GA1UdIwQYMBaAFKXO\nN+rrsHUOlGeItEX62SQQh5YfMB0GA1UdDgQWBBR2IlEI2jW5rEZwpysu3pH1Rd0e\nCzA8BgNVHREENTAzgg0qLmh0dHBzdGF0LnVzghVzbmkuY2xvdWRmbGFyZXNzbC5j\nb22CC2h0dHBzdGF0LnVzMA4GA1UdDwEB/wQEAwIHgDAdBgNVHSUEFjAUBggrBgEF\nBQcDAQYIKwYBBQUHAwIwewYDVR0fBHQwcjA3oDWgM4YxaHR0cDovL2NybDMuZGln\naWNlcnQuY29tL0Nsb3VkZmxhcmVJbmNFQ0NDQS0zLmNybDA3oDWgM4YxaHR0cDov\nL2NybDQuZGlnaWNlcnQuY29tL0Nsb3VkZmxhcmVJbmNFQ0NDQS0zLmNybDA+BgNV\nHSAENzA1MDMGBmeBDAECAjApMCcGCCsGAQUFBwIBFhtodHRwOi8vd3d3LmRpZ2lj\nZXJ0LmNvbS9DUFMwdgYIKwYBBQUHAQEEajBoMCQGCCsGAQUFBzABhhhodHRwOi8v\nb2NzcC5kaWdpY2VydC5jb20wQAYIKwYBBQUHMAKGNGh0dHA6Ly9jYWNlcnRzLmRp\nZ2ljZXJ0LmNvbS9DbG91ZGZsYXJlSW5jRUNDQ0EtMy5jcnQwDAYDVR0TAQH/BAIw\nADCCAX0GCisGAQQB1nkCBAIEggFtBIIBaQFnAHUAKXm+8J45OSHwVnOfY6V35b5X\nfZxgCvj5TV0mXCVdx4QAAAF6Tc4PQAAABAMARjBEAiAD/ek6JJV6HIJuzcJpx+ta\nmjcQFgUnbqWpf7FdtYU90gIgXQeRW/xptGGfPpM9VKTcGVkEdtfQeBHwceT0oGxN\npm4AdgAiRUUHWVUkVpY/oS/x922G4CMmY63AS39dxoNcbuIPAgAAAXpNzg7xAAAE\nAwBHMEUCIQDYUDr9qYuGMLHhosa2CpzmiH89DKoVNyYhhGv9LTJ6DwIgQBb1O1rC\n7ccAf0GZ2g7oGWXcgW8/CaVsWrOYgZFBbmsAdgBRo7D1/QF5nFZtuDd4jwykeswb\nJ8v3nohCmg3+1IsF5QAAAXpNzg8NAAAEAwBHMEUCIQCYqpZCKWyKr3DEYFik9lp1\n28VbvyAhV9BL0NlrUJZuGAIgRhOgP6coKy/ibeJwXw6EqrCxLX3eHBeZw+7fXFuC\nGIQwCgYIKoZIzj0EAwIDSAAwRQIhAKeFaqw8irbE3256KyNqGDdi9j75tdvnl2di\nPo2y6T5yAiArJMm0Szw1u8ijmARYN5t0SCCs22NqoicwyfVt8mw3Ig==\n-----END CERTIFICATE-----\n\n-----BEGIN CERTIFICATE-----\nMIIDzTCCArWgAwIBAgIQCjeHZF5ftIwiTv0b7RQMPDANBgkqhkiG9w0BAQsFADBa\nMQswCQYDVQQGEwJJRTESMBAGA1UEChMJQmFsdGltb3JlMRMwEQYDVQQLEwpDeWJl\nclRydXN0MSIwIAYDVQQDExlCYWx0aW1vcmUgQ3liZXJUcnVzdCBSb290MB4XDTIw\nMDEyNzEyNDgwOFoXDTI0MTIzMTIzNTk1OVowSjELMAkGA1UEBhMCVVMxGTAXBgNV\nBAoTEENsb3VkZmxhcmUsIEluYy4xIDAeBgNVBAMTF0Nsb3VkZmxhcmUgSW5jIEVD\nQyBDQS0zMFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAEua1NZpkUC0bsH4HRKlAe\nnQMVLzQSfS2WuIg4m4Vfj7+7Te9hRsTJc9QkT+DuHM5ss1FxL2ruTAUJd9NyYqSb\n16OCAWgwggFkMB0GA1UdDgQWBBSlzjfq67B1DpRniLRF+tkkEIeWHzAfBgNVHSME\nGDAWgBTlnVkwgkdYzKz6CFQ2hns6tQRN8DAOBgNVHQ8BAf8EBAMCAYYwHQYDVR0l\nBBYwFAYIKwYBBQUHAwEGCCsGAQUFBwMCMBIGA1UdEwEB/wQIMAYBAf8CAQAwNAYI\nKwYBBQUHAQEEKDAmMCQGCCsGAQUFBzABhhhodHRwOi8vb2NzcC5kaWdpY2VydC5j\nb20wOgYDVR0fBDMwMTAvoC2gK4YpaHR0cDovL2NybDMuZGlnaWNlcnQuY29tL09t\nbmlyb290MjAyNS5jcmwwbQYDVR0gBGYwZDA3BglghkgBhv1sAQEwKjAoBggrBgEF\nBQcCARYcaHR0cHM6Ly93d3cuZGlnaWNlcnQuY29tL0NQUzALBglghkgBhv1sAQIw\nCAYGZ4EMAQIBMAgGBmeBDAECAjAIBgZngQwBAgMwDQYJKoZIhvcNAQELBQADggEB\nAAUkHd0bsCrrmNaF4zlNXmtXnYJX/OvoMaJXkGUFvhZEOFp3ArnPEELG4ZKk40Un\n+ABHLGioVplTVI+tnkDB0A+21w0LOEhsUCxJkAZbZB2LzEgwLt4I4ptJIsCSDBFe\nlpKU1fwg3FZs5ZKTv3ocwDfjhUkV+ivhdDkYD7fa86JXWGBPzI6UAPxGezQxPk1H\ngoE6y/SJXQ7vTQ1unBuCJN0yJV0ReFEQPaA1IwQvZW+cwdFD19Ae8zFnWSfda9J1\nCZMRJCQUzym+5iPDuI9yP+kHyCREU3qzuWFloUwOxkgAyXVjBYdwRVKD05WdRerw\n6DEdfgkfCv4+3ao8XnTSrLE=\n-----END CERTIFICATE-----\n"
        }
    ]
}
```

#### `Get`

Runs a `HEAD` request agains the given URL

**Query format:** A URL e.g. `https://www.google.com`

#### `Find`

This method is not implemented.

### IP

The source returns details about an IP address. Note that it doesn't make any external requests and mostly exists to provide a node in the graph for other nodes to be linked through.

Example item:

```json
{
    "type": "ip",
    "uniqueAttribute": "ip",
    "attributes": {
        "attrStruct": {
            "interfaceLocalMulticast": false,
            "ip": "192.168.1.27",
            "linkLocalMulticast": false,
            "linkLocalUnicast": false,
            "loopback": false,
            "multicast": false,
            "private": true,
            "unspecified": false
        }
    },
    "context": "global"
}
```


#### `Get`

Returns IP information for a given IPv4 or IPv6 address.

**Query format:** An IP in a format that can be parsed by net.ParseIP() such as `192.0.2.1` or `2001:db8::68`

#### `Find`

This method is not implemented.

## Config

All configuration options can be provided via the command line or as environment variables:

| Environment Variable | CLI Flag | Automatic | Description |
|----------------------|----------|-----------|-------------|
| `CONFIG`| `--config` | ✅ | Config file location. Can be used instead of the CLI or environment variables if needed |
| `LOG`| `--log` | ✅ | Set the log level. Valid values: panic, fatal, error, warn, info, debug, trace |
| `NATS_SERVERS`| `--nats-servers` | ✅ | A list of NATS servers to connect to |
| `NATS_NAME_PREFIX`| `--nats-name-prefix` | ✅ | A name label prefix. Sources should append a dot and their hostname .{hostname} to this, then set this is the NATS connection name which will be sent to the server on CONNECT to identify the client |
| `NATS_CA_FILE`| `--nats-ca-file` | ✅ | Path to the CA file that NATS should use when connecting over TLS |
| `NATS_JWT_FILE`| `--nats-jwt-file` | ✅ | Path to the file containing the user JWT |
| `NATS_NKEY_FILE`| `--nats-nkey-file` | ✅ | Path to the file containing the NKey seed |
| `MAX-PARALLEL`| `--max-parallel` | ✅ | Max number of requests to run in parallel |
| `YOUR_CUSTOM_FLAG`| `--your-custom-flag` |   | Configuration that you add should be documented here |

### `srcman` config

When running in srcman, all of the above parameters marked with a checkbox are provided automatically, any additional parameters must be provided under the `config` key. These key-value pairs will become files in the `/etc/srcman/config` directory within the container.

```yaml
apiVersion: srcman.example.com/v0
kind: Source
metadata:
  name: stdlib-source
spec:
  image: ghcr.io/overmindtech/stdlib-source:latest
  replicas: 2
  manager: manager-source
```

**NOTE:** Remove the above boilerplate once you know what configuration will be required.

### Health Check

The source hosts a health check on `:8080/healthz` which will return an error if NATS is not connected. An example Kubernetes readiness probe is:

```yaml
readinessProbe:
  httpGet:
    path: /healthz
    port: 8080
```

## Development

### Running Locally

The source CLI can be interacted with locally by running:

```shell
go run main.go --help
```

### Testing

Tests in this package can be run using:

```shell
go test ./...
```

### Packaging

Docker images can be created manually using `docker build`, but GitHub actions also exist that are able to create, tag and push images. Images will be build for the `main` branch, and also for any commits tagged with a version such as `v1.2.0`
