include .env

help:
	@echo "Commands:"
	@echo "  build       Build all artifacts (binary, image)"
	@echo "  clean       Remove all artifacts (binary, image)"
	@echo "  help        List of commands"
	@echo "  run         Run App locally"
	@echo "  start       Build and force start a Container"
	@echo "  stop        Force remove a Container"
clean:
	rm -f app main
	docker image prune -f
	docker container prune -f
build:
	go build ./cmd/app/main.go
	docker build --platform=linux/amd64 -t ${IMAGE_NAME} .
run:
	go run ./cmd/app/main.go
start:
	$(MAKE) build
	docker container rm --force ${CONTAINER_NAME}
	docker container run --name ${CONTAINER_NAME} ${IMAGE_NAME}
stop:
	docker container rm --force ${CONTAINER_NAME}
