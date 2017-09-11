# go-client-cert-poc
This repo demonstrates a system by which SSL client certificates can be used dynamically for authentication between client-server apps. The app contains a server-mode and client-mode. It allows a client without a client cert to register with a name. Upon registration, the server generates and sends a client cert which it has associated with the name. Then, in subsequent requests, the client can present that cert and the server can identify it.

In one terminal:
```
paul-sanford-mbp:go-client-cert-poc paul.sanford$ go build
paul-sanford-mbp:go-client-cert-poc paul.sanford$ ./go-client-cert-poc -s
2017/09/11 01:12:31 Created CA key
2017/09/11 01:12:31 Created Server key
2017/09/11 01:12:31 Starting server
```

In another:
```
paul-sanford-mbp:go-client-cert-poc paul.sanford$ ./go-client-cert-poc -c
2017/09/11 01:12:43 Doing things with Paul
2017/09/11 01:12:43
2017/09/11 01:12:43 Please register; no user found for cert with serial
2017/09/11 01:12:43
2017/09/11 01:12:43 Doing things with Zac
2017/09/11 01:12:43
2017/09/11 01:12:43 Doing things with Paul
2017/09/11 01:12:43
2017/09/11 01:12:43 Doing things with Zac
2017/09/11 01:12:43
```
