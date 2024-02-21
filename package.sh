#!/usr/bin/env bash

GWK_VERSION=`git describe --tags --abbrev=0`
GIT_HASH=`git rev-parse --short HEAD`
echo $GWK_VERSION'-'$GIT_HASH

cwd=`pwd`


rm -rf ./release
mkdir -p ./release

os_all='linux windows darwin freebsd'
arch_all='386 amd64 arm arm64'

for os in $os_all; do
    for arch in $arch_all; do
        targetzipname="gwk_${GWK_VERSION}_${os}_${arch}"
        gwk_outdir="release/${targetzipname}"
        output_name1=gwk
        output_name2=gwkd
        mkdir -p $gwk_outdir
        if [ $os = "windows" ]; then
            output_name1+='.exe'
            output_name2+='.exe'
        fi
        echo "Build ${targetzipname}...";\
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
done

cd -