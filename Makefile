build-image:
	docker build -t hellowoodes/clash-reporter .

push-image:
	docker push hellowoodes/clash-reporter

build:
	go build -o clash-reporter .

