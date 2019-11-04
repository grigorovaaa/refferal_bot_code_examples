gopath = $(shell go env GOPATH)
current_dir = $(shell pwd)

# all
git-pull:
	git pull

dep-all:
	go get github.com/sirupsen/logrus
	go get github.com/Syfaro/telegram-bot-api
	go get github.com/lib/pq
	go get github.com/golang-migrate/migrate
	go get github.com/go-redis/redis

migrate-build: dep-all
	go build -tags 'postgres' -ldflags="-X main.Version='v4'" -o $(current_dir)/tools/migrate github.com/golang-migrate/migrate/cmd/migrate

# Vagrant

migrate-vagrant-up: migrate-build
	$(current_dir)/tools/migrate -source=file://$(current_dir)/build/migrations/ -database=postgres://vagrant:vagrant@localhost:5432/referral_bot?sslmode=disable up

migrate-vagrant-down: migrate-build
	$(current_dir)/tools/migrate -source=file://$(current_dir)/build/migrations/ -database=postgres://vagrant:vagrant@localhost:5432/referral_bot?sslmode=disable down

build-daemon-bot-vagrant: migrate-vagrant-up
	go build -o $(current_dir)/bin/daemon-bot $(current_dir)/cmd/daemon/bot/main.go

run-daemon-bot-vagrant: build-daemon-bot-vagrant
	$(current_dir)/bin/daemon-bot -conf=$(current_dir)/configs/daemon-bot-vagrant.json


# Prod

migrate-prod-up: migrate-build
	$(current_dir)/tools/migrate -source=file://$(current_dir)/build/migrations/ -database=postgres://referral_bot:Quoiju_eshahfe8@127.0.0.1:5432/referral_bot?sslmode=disable up

migrate-prod-down: migrate-build
	$(current_dir)/tools/migrate -source=file://$(current_dir)/build/migrations/ -database=postgres://referral_bot:Quoiju_eshahfe8@127.0.0.1:5432/referral_bot?sslmode=disable down

build-daemon-bot-prod: git-pull migrate-prod-up
	go build -o $(current_dir)/bin/daemon-bot $(current_dir)/cmd/daemon/bot/main.go

run-daemon-bot-prod: build-daemon-bot-prod
	$(current_dir)/bin/daemon-bot -conf=$(current_dir)/configs/daemon-bot-prod.json
