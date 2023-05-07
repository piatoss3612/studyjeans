run:
	@echo "Run discord bot"
	go run ./cmd/study/

build_image:
	@echo "Build docker image"
	docker build -t piatoss3612/presentation-helper-bot:$(version) -f ./build/study/Dockerfile .

push_image:
	@echo "Push docker image"
	docker push piatoss3612/presentation-helper-bot:$(version)