help:
	@echo "Commands:"
	@echo "  build       Build artifacts"
	@echo "  clean       Remove all artifacts"
	@echo "  help        List of commands"
	@echo "  run         Run App locally"
build:
	go build ./cmd/app/main.go
clean:
	rm -f app main
run:
	go run ./cmd/app/main.go
