# gwk

Gwk is a tool that helps you expose your local servers or services to the
internet, even in a private network. It supports both TCP and subdomain modes.

## build

```bash
bash package.sh
```

## usage

serverHost default is `gank.75cos.com`

```bash
# example 1 , detault dispatch to 127.0.0.1:8080
gwk
```

## client more  example

```bash
# example 2
gwk  --port 8080
# example 3
gwk  --subdomain testabc001 --port 8000
# example 4
gwk  -c client.json
```

## develop 


1. generate root CA

```bash
bash ./scripts/gen_rootca.sh
```

2. generate domain cert

```bash
bash ./scripts/gen_certbyca.sh
```

3. move `certs` to `etc` directory


## client

```bash
go run ./bin/gwk/main.go  -c etc/client.json
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

## setup a gwk server

```bash
go run ./bin/gwkd/main.go  -c etc/server.json
```

server.json

```json
{
  "serverHost": "gank007.com",
  "serverPort": 4443,
  "httpAddr": 8080,
  "httpsAddr": 8043,
  "tlsCA":"./scripts/certs/rootCA.crt",
  "tlsCrt":"./scripts/certs/gank007.com/my.crt",
  "tlsKey":"./scripts/certs/gank007.com/my.key.pem"
}

```

