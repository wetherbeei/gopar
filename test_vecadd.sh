go install gopar || exit 
GOPATH=./examples/vecadd
./bin/gopar build vecadd
./vecadd
