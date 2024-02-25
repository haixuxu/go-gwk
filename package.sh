#!/usr/bin/env bash

output_dir="./release"

APP_VERSION=`git describe --tags --abbrev=0`
GIT_HASH=`git rev-parse --short HEAD`
echo $APP_VERSION'-'$GIT_HASH

rm -rf $output_dir
mkdir -p $output_dir

# platforms=( "darwin/amd64" "darwin/arm64" )
platforms=("windows/amd64" "windows/386" "windows/arm64"  "linux/amd64" "linux/386" "linux/arm64" "darwin/amd64" "darwin/arm64" "freebsd/amd64" "freebsd/386" "freebsd/arm64" )
# 写入_sha256_checksums.txt文件
checksum_txt=gwk_sha256_checksums.txt

for platform in "${platforms[@]}"
do
    platform_split=(${platform//\// })
    os=${platform_split[0]}
    arch=${platform_split[1]}

    targetzip_name="gwk_${APP_VERSION}_${os}_${arch}"
    build_outdir="${output_dir}/${targetzip_name}"
    platform_bin1=gwk
    platform_bin2=gwkd
    mkdir -p $build_outdir
    if [ $os = "windows" ]; then
        platform_bin1+='.exe'
        platform_bin2+='.exe'
    fi
    echo "Build ${build_outdir}...";\
    env CGO_ENABLED=0 GOOS=${os} GOARCH=${arch} go build -trimpath -ldflags "-X main.Version=$APP_VERSION -X main.GitCommitHash=$GIT_HASH" -o ${build_outdir}/${platform_bin1} ./bin/gwk
    env CGO_ENABLED=0 GOOS=${os} GOARCH=${arch} go build -trimpath -ldflags "-X main.Version=$APP_VERSION -X main.GitCommitHash=$GIT_HASH"  -o ${build_outdir}/${platform_bin2} ./bin/gwkd
    echo "Build ${build_outdir} done";

    cp ./LICENSE ${build_outdir}
    cp -rf ./etc ${build_outdir}

    # packages
    cd $output_dir
    if [ $os = "windows" ]; then
        zip -rq ${targetzip_name}.zip ${targetzip_name}
        sha256sum "$targetzip_name.zip">> $checksum_txt
    else
        tar -zcf ${targetzip_name}.tar.gz ${targetzip_name}
        sha256sum "$targetzip_name.tar.gz">> $checksum_txt
    fi  
    rm -rf ${targetzip_name}
    cd ..
done

cd -