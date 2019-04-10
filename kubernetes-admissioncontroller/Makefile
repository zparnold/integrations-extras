build:
	docker build -t app .
	docker tag app:latest quay.io/ygrene/dd-admission-controller:latest
	docker push quay.io/ygrene/dd-admission-controller:latest