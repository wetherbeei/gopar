go install gopar || exit 

/usr/bin/time -f 'Real: %es, %PCPU %Uu %Ss %er %MkB %C' ./bin/gopar build "$1"
mv $1 ${1}_parallel
go build $1
