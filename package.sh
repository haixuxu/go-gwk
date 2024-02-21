#!/usr/bin/env bash

GWK_VERSION=`git describe --tags --abbrev=0`
GIT_HASH=`git rev-parse --short HEAD`
echo $GWK_VERSION'-'$GIT_HASH

cwd=`pwd`


rm -rf ./release
mkdir -p ./release

# os_all='linux windows darwin freebsd'
# arch_all='386 amd64 arm arm64'

platforms=("windows/amd64" "windows/386" "windows/arm64"  "linux/amd64" "linux/386" "linux/arm64" "darwin/amd64" "darwin/arm64" )

for platform in "${platforms[@]}"
do
    platform_split=(${platform//\// })
    os=${platform_split[0]}
    arch=${platform_split[1]}

    targetzipname="gwk_${GWK_VERSION}_${os}_${arch}"
    gwk_outdir="release/${targetzipname}"
    output_name1=gwk
    output_name2=gwkd
    mkdir -p $gwk_outdir
    if [ $os = "windows" ]; then
        output_name1+='.exe'
        output_name2+='.exe'
    fi
    echo "Build release/${targetzipname}...";\
    env CGO_ENABLED=0 GOOS=${os} GOARCH=${arch} go build -trimpath -ldflags "-X main.Version=$GWK_VERSION -X main.GitCommitHash=$GIT_HASH" -o ${gwk_outdir}/${output_name1} ./bin/gwk
    env CGO_ENABLED=0 GOOS=${os} GOARCH=${arch} go build -trimpath -ldflags "-X main.Version=$GWK_VERSION -X main.GitCommitHash=$GIT_HASH"  -o ${gwk_outdir}/${output_name2} ./bin/gwkd
    echo "Build ${gwk_outdir} done";

    cp ./LICENSE ${gwk_outdir}
    cp -rf ./etc ${gwk_outdir}

    # packages
    cd release
    if [ $os = "windows" ]; then
        zip -rq ${targetzipname}.zip ${targetzipname}
    else
        tar -zcf ${targetzipname}.tar.gz ${targetzipname}
    fi  
    rm -rf ${targetzipname}
    cd ..
done

cd -