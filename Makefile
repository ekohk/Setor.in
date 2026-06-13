.PHONY: docker-build
docker-build:
	docker build --pull --tag setorin/backend:local -f backend/Dockerfile backend
