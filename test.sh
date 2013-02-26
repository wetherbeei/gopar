go install gopar || exit 
GOPATH=./examples/$1
./bin/gopar build $1
./$1
