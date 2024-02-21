# gwk

Gwk is a tool that helps you expose your local servers or services to the
internet, even in a private network. It supports both TCP and subdomain modes.

# build

```bash
bash build.sh
```

# usage

serverHost default is `gank.75cos.com`

```bash
# example 1 , detault dispatch to 127.0.0.1:8080
gwk
```

# client more  example

```bash
# example 2
gwk  --port 8080
# example 3
gwk  --subdomain testabc001 --port 8000
# example 4
gwk  -c client.json
```

# client

```bash
go run ./bin/gwk/main.go  -c client.json
```

client.json

```json
{
  "serverHost": "gank007.com",
  "serverPort": 4443,
  "tunnels": {
    "tcp001": {
      "protocol": "tcp",
      "localPort": 5000,
      "remotePort": 7200
    },
    "tcp002": {
      "protocol": "tcp",
      "localPort": 5000,
      "remotePort": 7500
    },
    "webapp02": {
      "protocol": "web",
      "localPort": 4900,
      "subdomain": "app02"
    },
    "webappmob": {
      "protocol": "web",
      "localPort": 9000,
      "subdomain": "mob"
    }
  }
}
```

# setup a gwk server

```bash
go run ./bin/gwkd/main.go  -c server.json
```

server.json

```json
{
  "serverHost": "gwk007.com",
  "serverPort": 4443,
  "httpAddr": 80,
  "httpsAddr": 443,
  "tlsCA": "./rootCA/rootCA.crt",
  "tlsCrt": "./cert/my.crt",
  "tlsKey": "./cert/my.key.pem"
}
```
