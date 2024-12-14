# Variables
APP_NAME = spade
IMAGE_TAG = latest
DOCKERFILE_PATH = Dockerfile
PORT = 8080

all: build

build:
	go build -o APP_NAME

docker-build:
	docker build -t $(APP_NAME):$(IMAGE_TAG) -f $(DOCKERFILE_PATH) .

docker-run:	docker-build
	docker run -d -p $(PORT):$(PORT) --name $(APP_NAME) --restart always $(APP_NAME):$(IMAGE_TAG)

docker-stop:
	docker stop $(APP_NAME)

docker-restart:
	docker-stop
	docker-run

clean:
	rm -f $(APP_NAME)