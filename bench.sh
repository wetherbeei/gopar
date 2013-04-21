/usr/bin/time -f 'Real: %es, %PCPU %Uu %Ss %er %MkB %C' ./$1 $2
for procs in 1 2 4 8
do
GOMAXPROCS=$procs /usr/bin/time -f 'Real: %es, %PCPU %Uu %Ss %er %MkB %C' ./${1}_parallel $2
done