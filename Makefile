
build:
	docker build -t ashno-onepay .
swagger:
	#go install github.com/swaggo/swag/cmd/swag@v1.8.4
	swag init -d internal/server -g swagger.go -o docs/swagger --parseDependency --parseInternal