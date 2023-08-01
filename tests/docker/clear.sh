docker-compose down --volumes --remove-orphans

sudo rm -rf gaia_validator_1

sudo rm -rf celinium_node

docker rmi docker-gaia-validator-1

docker rmi docker-relayer

# docker rmi celinium
