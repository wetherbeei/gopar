rm ./$1
go install gopar || exit 

/usr/bin/time -f 'Real: %es, %PCPU %Uu %Ss %er %MkB %C' ./bin/gopar build "$1"
GOMAXPROCS=2 /usr/bin/time -f 'Real: %es, %PCPU %Uu %Ss %er %MkB %C' "./$1"
GOMAXPROCS=4 /usr/bin/time -f 'Real: %es, %PCPU %Uu %Ss %er %MkB %C' "./$1"
rm ./$1
go build $1
/usr/bin/time -f 'Real: %es, %PCPU %Uu %Ss %er %MkB %C' "./$1"
