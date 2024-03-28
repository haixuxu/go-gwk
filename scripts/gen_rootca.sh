#!/bin/bash

# 定义变量
country="CN"
organization="xuxihai"
common_name="gwkbyxuxihai"

cwd=`pwd`

outdir="${cwd}/certs"
rootcakey=$outdir/rootCA.key.pem
rootcacrt=$outdir/rootCA.crt

mkdir -p $outdir

# 生成根证书的私钥
openssl genpkey -algorithm RSA -out $rootcakey

# 生成自签名的根证书
openssl req -new -x509 \
    -key $rootcakey \
    -out $rootcacrt \
    -subj "/C=$country/O=$organization/CN=$common_name"

rm -f "${outdir}/rootCA.srl"
# 输出成功信息
echo "根证书生成成功!"
echo "根证书的私钥: $rootcakey"
echo "自签名根证书: $rootcacrt"

