# gwk

Gwk is a tool that helps you expose your local servers or services to the
internet, even in a private network. It supports both TCP and subdomain modes.

# build

```bash
bash build.sh
```

# client

```bash
go run ./bin/gwk/main.go  -c client.json
```

client.json

```json
{
  "serverHost": "gank007.com", // 服务器地址
  "serverPort": 4443, // 服务器端口
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
  "serverHost": "gwk007.com", // 使用web 隧道时, 需要域名
  "serverPort": 4443, // 隧道监听端口
  "httpAddr": 80, // 启动http服务
  "httpsAddr": 443, // 启动https服务, 需要后面的证书配置
  "tlsCA": "./rootCA/rootCA.crt", // 使用自签名证书用到
  "tlsCrt": "./cert/my.crt",
  "tlsKey": "./cert/my.key.pem"
}
```
