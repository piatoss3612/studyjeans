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

build_image: build_study_image build_recorder_image
	

build_study_image:
	@echo "Build docker image for study bot"
	docker build -t piatoss3612/study-bot:$(version) -f ./build/study/Dockerfile .

build_recorder_image:
	@echo "Build docker image for recorder"
	docker build -t piatoss3612/recorder:$(version) -f ./build/recorder/Dockerfile .

push_image: push_study_image push_recorder_image

push_study_image:
	@echo "Push study bot docker image"
	docker push piatoss3612/study-bot:$(version)

push_recorder_image:
	@echo "Push recorder docker image"
	docker push piatoss3612/recorder:$(version)