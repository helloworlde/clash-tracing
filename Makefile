build-image:
	docker build -t hellowoodes/clash-tracing .

push-image:
	docker push hellowoodes/clash-tracing

build:
	go build -o clash-tracing .

