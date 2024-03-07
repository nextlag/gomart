build:
	docker build -t gomart ..
run:
	docker run -p 8080:8080 gomart
test:
	docker-compose -f ./Docker-Compose.yaml up --build --abort-on-container-exit
	docker-compose -f ./Docker-Compose.yaml down --volumes
postgres:
	docker run --name gophermart -p 5432:5432 -e POSTGRES_USER=postgres -e POSTGRES_PASSWORD=Xer_0101 -e POSTGRES_DB=gophermart -d postgres:12-alpine
