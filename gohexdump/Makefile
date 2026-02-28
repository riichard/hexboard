DEVICE  ?= txt
BINARY  ?= hexboard
GOBIN   := /usr/local/go/bin/go
SRCDIR  := $(shell dirname $(realpath $(firstword $(MAKEFILE_LIST))))

.PHONY: deploy run stop sync build

# Sync sources, build on device, and restart
deploy: sync build restart

# Sync source tree to device
sync:
	rsync -a --exclude='*.test' --exclude='.DS_Store' $(SRCDIR)/ $(DEVICE):~/dev/hexboard/gohexdump/

# Build on the device (uses device's native Go toolchain)
build:
	ssh $(DEVICE) 'cd ~/dev/hexboard/gohexdump && $(GOBIN) build -o ~/$(BINARY) ./cmd/hexboard'

# Stop any running instance
stop:
	ssh $(DEVICE) 'pkill $(BINARY) 2>/dev/null; true'

# Start in the foreground (interactive)
run: stop
	ssh -t $(DEVICE) '~/$(BINARY)'

# Restart in the background (survives SSH disconnect)
restart: stop
	ssh $(DEVICE) 'nohup ~/$(BINARY) > /tmp/$(BINARY).log 2>&1 &'

# Show live log output
log:
	ssh $(DEVICE) 'tail -f /tmp/$(BINARY).log'

# Send a test message
test-message:
	@echo "$(MSG)" | nc $(shell ssh $(DEVICE) hostname -I | awk '{print $$1}') 8080
