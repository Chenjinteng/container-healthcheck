$path="main"
$version="V1.0.0-release"
$goversion=$(go version)
$buildtime=$(Get-Date -Format 'yyyy-MM-dd hh:mm:ss')
$author="jintengchen@outlook.com"
$gitcommit=$(git rev-parse HEAD)

$toolname="container-healthcheck"

$flags="-X '$path.Author=$author' -X $path.GitCommit=$gitcommit -X $path.Version=$version -X '$path.GoVersion=$goversion' -X '$path.BuildTime=$buildtime'"

# $yflags="-X 'bkhttpd/handlers.Author=$author' -X 'bkhttpd/handlers.GitCommit=$gitcommit' -X 'bkhttpd/handlers.Version=$version' -X 'bkhttpd/handlers.GoVersion=$goversion' -X 'bkhttpd/handlers.BuildTime=$buildtime'"

$env:CGO_ENABLED="0"
$env:GOOS="linux"
$env:GOARCH="amd64"


$version |Out-File -FilePath VERSION

if(Test-Path release/$toolname-$version) {
    Remove-Item release/$toolname-$version/* -Recurse -Force
}
else {
    New-Item -Path release/$toolname-$version/ -ItemType Directory
    New-Item -Path release/$toolname-$version/usr/local/bin/ -ItemType Directory 
    New-Item -Path release/$toolname-$version/etc/sysconfig/ -ItemType Directory 
    New-Item -Path release/$toolname-$version/usr/lib/systemd/system -ItemType Directory 
    New-Item -Path release/$toolname-$version/var/log/$toolname -ItemType Directory 
    New-Item -Path release/$toolname-$version/etc/rsyslog.d/ -ItemType Directory 
}



Copy-Item ./$toolname.service release/$toolname-$version/usr/lib/systemd/system/ -Recurse -Force
Copy-Item ./sysconfig release/$toolname-$version/etc/sysconfig/$toolname -Recurse -Force
Copy-Item ./rsyslog release/$toolname-$version/etc/rsyslog.d/$toolname -Recurse -Force

go build -ldflags "$flags $yflags" -o release/$toolname-$version/usr/local/bin/$toolname main.go