BINARY_NAME="container-healthcheck"
Version=$(cat VERSION)
Author="Chenjinteng"
GoVersion=$(go version)
BuildTime=$(date +'%Y-%m-%d %H:%M:%S')
GitCommit=$(git rev-parse --short HEAD)

FLAGS="-X 'main.Author=$Author' \
	   -X 'main.Version=$Version' \
	   -X 'main.GitCommit=$GitCommit' \
       -X 'main.GoVersion=$GoVersion' \
	   -X 'main.BuildTime=$BuildTime'" \


pre_make:
    mkdir -p release/${BINARY_NAME}-${Version}/usr/local/bin/
    mkdir -p release/${BINARY_NAME}-${Version}/etc/sysconfig/
    mkdir -p release/${BINARY_NAME}-${Version}/usr/lib/systemd/system/
    mkdir -p release/${BINARY_NAME}-${Version}/var/log/${BINARY_NAME}/
    mkdir -p release/${BINARY_NAME}-${Version}/etc/rsyslog.d/

build:  
    pre_make
    CGO_ENABLED=0 GOOS="linux" GOARCH="amd64" go build -o $(BINARY_NAME)



clean:
    rm -rf release/${BINARY_NAME}-${Version}/


dep:
    export GOPROXY=https://goproxy.io,direct
    go mod download