# Distributed-MutualExclusion[![GoDoc](https://img.shields.io/github/go-mod/go-version/Dreamacro/clash?style=flat-square)]()


Academic project to build an application that supports the execution of 3 mutual-exclusion algorithms: Lamport, Ricart-Agrawala and Token-Based Centralized. The application also provides a registration service to the mutual exclusion group.


Instruction for application execution:
## Clone repository
```
git clone https://github.com/mary-pie/Distributed-MutualExclusion/
cd Distributed-MutualExclusion
```
## Build images
1. Token-based centralized
	```
	docker build -t node-token -f docker/token.Dockerfile .
	```
2. Lamport
	```
	docker build -t node-lamport -f docker/lamport.Dockerfile .
	```
3. Ricart-Agrawala
	```
	docker build -t node-ra -f docker/ricart-agrawala.Dockerfile .
	```
## Run app with Compose
1. Token-based centralized
	```
	docker-compose -p token-centr -f docker/docker-compose.token.yml up
	```
2. Lamport
	```
	docker-compose -p lamport -f docker/docker-compose.lamport.yml up
	```
3. Ricart-Agrawala
	```
	docker-compose -p ricart-agrawala -f docker/docker-compose.ricartagrawala.yml up
	```
