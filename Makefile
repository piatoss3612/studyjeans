study:
	@echo "Run discord bot"
	go run ./cmd/study/

recorder:
	@echo "Run recorder"
	go run ./cmd/recorder/

up:
	@echo "Run docker compose"
	docker compose up -d

down:
	@echo "Stop docker compose"
	docker compose down

build_image:
	@echo "Build docker image"
	docker build -t piatoss3612/presentation-helper-bot:$(version) -f ./build/study/Dockerfile .

push_image:
	@echo "Push docker image"
	docker push piatoss3612/presentation-helper-bot:$(version)