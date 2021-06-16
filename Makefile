SHELL       =   /bin/sh
TAG         ?=  dev

.SUFFIXES:
.PHONY: help build-dev push-dev build push

help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

build-dev: ## Build DEV version
	@echo Build dev
	docker build --build-arg version=dev -t ghcr.io/negasus/reproxy-ip2location-plugin:dev -t negasus/reproxy-ip2location-plugin:dev .

push-dev: ## Build balerter/balerter and balerter/test images to docker registry DEV version
	@echo Push Balerter dev
	docker push negasus/reproxy-ip2location-plugin:dev
	docker push ghcr.io/negasus/reproxy-ip2location-plugin:dev

build: ## Build docker images
	@echo Build $(TAG)
	docker build --build-arg version=$(TAG) -t ghcr.io/negasus/reproxy-ip2location-plugin:$(TAG) -t ghcr.io/negasus/reproxy-ip2location-plugin:latest -t negasus/reproxy-ip2location-plugin:$(TAG) -t negasus/reproxy-ip2location-plugin:latest .

push: ## Build balerter/balerter and balerter/test images to docker registry
	@echo Push $(TAG)
	docker push negasus/reproxy-ip2location-plugin:$(TAG)
	docker push negasus/reproxy-ip2location-plugin:latest
	docker push ghcr.io/negasus/reproxy-ip2location-plugin:$(TAG)
	docker push ghcr.io/negasus/reproxy-ip2location-plugin:latest
