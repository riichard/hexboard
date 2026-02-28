DEVICE       ?= txt
BINARY       ?= hexboard
GOBIN        := /usr/local/go/bin/go
SRCDIR       := $(shell dirname $(realpath $(firstword $(MAKEFILE_LIST))))
SERVICE_DEST := /etc/systemd/system/hexboard.service

.PHONY: deploy run stop sync build install uninstall log test-message

# Sync sources, build on device, and restart via systemd (or plain if not installed)
deploy: sync build
	ssh $(DEVICE) 'systemctl is-active --quiet hexboard && systemctl restart hexboard || true'

# Sync source tree to device
sync:
	rsync -a --exclude='*.test' --exclude='.DS_Store' $(SRCDIR)/ $(DEVICE):~/dev/hexboard/gohexdump/

# Build on the device (uses device's native Go toolchain)
build:
	ssh $(DEVICE) 'cd ~/dev/hexboard/gohexdump && $(GOBIN) build -o ~/$(BINARY) ./cmd/hexboard'

# Stop any running instance (systemd or plain)
stop:
	ssh $(DEVICE) 'systemctl stop hexboard 2>/dev/null || pkill $(BINARY) 2>/dev/null; true'

# Start in the foreground (interactive, bypasses systemd)
run: stop
	ssh -t $(DEVICE) '~/$(BINARY)'

# Install and enable the systemd service (runs on boot, port 80)
install: sync build
	ssh $(DEVICE) 'sudo cp ~/dev/hexboard/gohexdump/hexboard.service $(SERVICE_DEST) && sudo systemctl daemon-reload && sudo systemctl enable hexboard && sudo systemctl restart hexboard'
	@echo "Service installed. Check status with: make status"

# Remove the systemd service
uninstall:
	ssh $(DEVICE) 'sudo systemctl stop hexboard; sudo systemctl disable hexboard; sudo rm -f $(SERVICE_DEST); sudo systemctl daemon-reload'

# Show service status
status:
	ssh $(DEVICE) 'systemctl status hexboard'

# Show live journal output
log:
	ssh $(DEVICE) 'journalctl -fu hexboard'

# Send a test message over TCP
test-message:
	@echo "$(MSG)" | nc $(shell ssh $(DEVICE) hostname -I | awk '{print $$1}') 8080
