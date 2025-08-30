SHORT ?= true
ARGS ?=

test:
	@if [ "$(SHORT)" = "true" ]; then \
		go test -short $(ARGS) ./...; \
	else \
		go test $(ARGS) ./...; \
	fi
.PHONY: test

help:
	@echo "Available targets:"
	@echo "  test    - run go test"
	@echo ""
	@echo "Variables:"
	@echo "  SHORT   - Use the -short flag when running the tests (default: true)"
	@echo "  ARGS    - Additional arguments to pass to go test"
	@echo ""
	@echo "Examples:"
	@echo "  make test"
	@echo "  make test SHORT=false"
	@echo "  make test ARGS='-v -race'"
