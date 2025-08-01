ADDR_BUILD := ":3000"
ADDR_DEV := ":6969"

MAX_SIZE := 32768

BUILD_DIR := target
APP_NAME := pastebin

.PHONY: default
default: dev

.PHONY: dev
dev:
	@go run . -addr="$(ADDR_DEV)" -max-size=$(MAX_SIZE)

.PHONY: gen
gen:
	@go tool templ generate

.PHONY: test
test:
	@go test -v ./...

.PHONY: build
build: gen
	@mkdir -p $(BUILD_DIR)
	@go build -o $(BUILD_DIR)/$(APP_NAME) .

.PHONY: run
run: build
	@$(BUILD_DIR)/$(APP_NAME) -addr=$(ADDR_BUILD) -max-size=$(MAX_SIZE)

.PHONY: clean
clean:
	@rm -rf $(BUILD_DIR)
