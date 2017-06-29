cd .. ; go build ; cd - ; cp ../hod hod
docker build -t gtfierro/hoddb .
docker push gtfierro/hoddb
