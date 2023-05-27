study:
	@echo "Run study bot"
	go run ./cmd/study/

logger:
	@echo "Run study logger"
	go run ./cmd/logger/

up:
	@echo "Run docker compose"
	docker compose up -d

down:
	@echo "Stop docker compose"
	docker compose down

build_image: build_study_image build_logger_image
	

build_study_image:
	@echo "Build docker image for study bot"
	docker build -t piatoss3612/study-bot:$(version) -f ./build/study/Dockerfile .

build_logger_image:
	@echo "Build docker image for study logger"
	docker build -t piatoss3612/study-logger:$(version) -f ./build/logger/Dockerfile .

push_image: push_study_image push_logger_image

push_study_image:
	@echo "Push study bot docker image"
	docker push piatoss3612/study-bot:$(version)

push_logger_image:
	@echo "Push study logger docker image"
	docker push piatoss3612/study-logger:$(version)