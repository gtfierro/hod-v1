#!/bin/bash
set -ex
sudo docker pull gtfierro/hod
sudo docker kill hod
sudo docker rm hod
sudo docker run -d --name hod -e "HOD_TLSHOST=mysite.org" -e "HOD_SERVERPORT=443" -p443:443 gtfierro/hod
sudo docker logs -f hod
