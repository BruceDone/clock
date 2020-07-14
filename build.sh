echo "clear the webapp"
rm -rf webapp
mkdir -p ./webapp
cp -r /Users/brucedone/projects/personal/clock-admin/dist/* ./webapp

echo "generate the packr2 files"
packr2 clean
packr2

echo "generate the linux binary"
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go generate
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build
mv ./clock ./clock-linux-64

echo "generate the mac binary"
go generate
go build
mv ./clock ./clock-mac-64

