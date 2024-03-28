build-binary:
	go build -o bin/ddot .

run-binary: build-binary
	./bin/ddot

build-image:
	docker buildx build --platform linux/arm/v8 -t ecojuntak/ddot:latest .

run-container:
	docker run -d -p 5533:5533 -p 5533:5533/udp -v ./.env:/.env --name ddot ecojuntak/ddot:latest ddot

stop-container:
	docker container stop ddot
	docker container rm ddot
