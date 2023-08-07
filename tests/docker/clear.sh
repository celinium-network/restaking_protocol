docker rmi -f docker-coordinator
docker rmi -f docker-consumer
docker rmi -f docker-relayer

docker rm docker-consumer-1
docker rm docker-coordinator-1
docker rm docker-relayer-1

sudo rm -rf tests/docker/consumer
sudo rm -rf tests/docker/coordinator