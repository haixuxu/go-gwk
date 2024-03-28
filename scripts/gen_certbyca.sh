#!/bin/bash

# 定义变量
country="CN"
organization="gwk007"
common_name="gank007.com"
dns_names=("*.gank007.com" "gank007.com")

cwd=`pwd`

outdir="${cwd}/certs/gank007.com"

rootcakey="./certs/rootCA.key.pem"
rootcacrt="./certs/rootCA.crt"

mkdir -p $outdir

subcert_key="${outdir}/my.key.pem"
subcert_csr="${outdir}/my.csr"
subcert_crt="${outdir}/my.crt"

# 生成子证书的私钥
openssl genpkey -algorithm RSA -out $subcert_key

# 生成证书签名请求
openssl req -new \
    -key $subcert_key \
    -out $subcert_csr \
    -subj "/C=$country/O=$organization/CN=$common_name"

# 创建扩展配置文件
echo -e "authorityKeyIdentifier=keyid,issuer\nbasicConstraints=CA:FALSE\nkeyUsage = digitalSignature, nonRepudiation, keyEncipherment, dataEncipherment\nsubjectAltName = @alt_names\n\n[alt_names]" > sub.ext

# 添加DNS条目到扩展配置文件
index=1
for dns in "${dns_names[@]}"; do
  echo "DNS.$index = $dns" >> sub.ext
  index=$((index + 1))
done

# 使用根证书签发子证书
openssl x509 -req \
    -in $subcert_csr \
    -CA $rootcacrt \
    -CAkey $rootcakey \
    -CAcreateserial -out $subcert_crt \
    -extfile sub.ext


rm -f sub.ext
# 输出成功信息
echo "子证书生成成功！"

echo "子证书的私钥: $subcert_key"
echo "子证书的CSR: $subcert_csr"
echo "子证书的证书: $subcert_crt"

