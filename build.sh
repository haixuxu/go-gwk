#!/usr/bin/env bash

program_names=("gwk" "gwkd")

version=v0.0.1

platforms=("windows/amd64" "windows/386" "windows/arm64"  "linux/amd64" "linux/386" "linux/arm64" "darwin/amd64" "darwin/arm64" )

rm -rf release

for platform in "${platforms[@]}"
do
    platform_split=(${platform//\// })
    GOOS=${platform_split[0]}
    GOARCH=${platform_split[1]}

    output_dir='release/'$version'/'$GOOS'-'$GOARCH'/'

    # 创建输出目录
    mkdir -p $output_dir

    for program_name in "${program_names[@]}"
    do
        output_name=$output_dir$program_name

        if [ $GOOS = "windows" ]; then
            output_name+='.exe'
        fi

        build_maingo='./bin/'$program_name'/main.go'

        cmd="env GOOS=$GOOS GOARCH=$GOARCH go build -o $output_name $build_maingo"
        echo $cmd
        `$cmd`
#        env GOOS=$GOOS GOARCH=$GOARCH go build -o $output_name $build_maingo
        if [ $? -ne 0 ]; then
            echo 'An error has occurred! Aborting the script execution...'
            exit 1
        fi
    done
done
