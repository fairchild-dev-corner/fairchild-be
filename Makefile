# Default environment
ENV := dev

# Intercept CLI-style flags like `--dev` or `--stage`
ifeq ($(firstword $(MAKECMDGOALS)),--dev)
  ENV := dev
  override MAKECMDGOALS := $(wordlist 2,$(words $(MAKECMDGOALS)),$(MAKECMDGOALS))
endif

ifeq ($(firstword $(MAKECMDGOALS)),--stage)
  ENV := stage
  override MAKECMDGOALS := $(wordlist 2,$(words $(MAKECMDGOALS)),$(MAKECMDGOALS))
endif

#Targets
.PHONY: run-dev run-staging build-dev build-stage run build migrate-up migrate-down migrate-up-dev migrate-down-dev migrate-up-stage migrate-down-stage migrate-legacy-members migrate-legacy-credentials

run-dev:
	$(MAKE) run ENV=dev
	
run-staging:
	$(MAKE) run ENV=stage

build-dev:
	$(MAKE) build ENV=dev

build-stage:
	$(MAKE) build ENV=stage

run: build
	@bin/fairchild_be

build:
	@go build -ldflags "-X 'main.BuildEnv=${ENV}'" -o bin/fairchild_be cmd/main.go

migrate-up-dev:
	$(MAKE) migrate-up ENV=dev

migrate-down-dev:
	$(MAKE) migrate-down ENV=dev

migrate-up-stage:
	$(MAKE) migrate-up ENV=stage

migrate-down-stage:
	$(MAKE) migrate-down ENV=stage

migrate-up:
	@go run -ldflags "-X 'main.BuildEnv=${ENV}'" cmd/migrate/main.go up

migrate-down:
	@go run -ldflags "-X 'main.BuildEnv=${ENV}'" cmd/migrate/main.go down

migrate-legacy-members:
	@go run -ldflags "-X 'main.BuildEnv=${ENV}'" cmd/migrate_legacy_members/main.go

migrate-legacy-credentials:
	@go run -ldflags "-X 'main.BuildEnv=${ENV}'" cmd/migrate_legacy_credentials/main.go

## Usage:
# make run-dev
# make run-staging

# make build-dev
# make build-stage

# make migrate-up-dev
# make migrate-down-dev
# make migrate-up-stage
# make migrate-down-stage

# make migrate-legacy-members --dev      (member_record -> users, run first)
# make migrate-legacy-credentials --dev  (tb_account -> user_account_cred, run second)