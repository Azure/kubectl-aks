IE_VERSION ?= 40e125e61516ab401084da65eed36095b2b3850a
INSTALLATION_DIR  ?= ~/.local/bin

# Install innovation engine
.PHONY: install-ie
install-ie:
	DIR=$$(mktemp -d) && \
	cd $$DIR && \
	git clone https://github.com/Azure/InnovationEngine.git && \
	cd InnovationEngine && \
	git checkout $(IE_VERSION) && \
	make build-ie && \
	mkdir -p $(INSTALLATION_DIR) && \
	cp bin/ie $(INSTALLATION_DIR) && \
	rm -rf $$DIR

# Uninstall innovation engine
.PHONY: uninstall-ie
uninstall-ie:
	rm -f $(INSTALLATION_DIR)/ie
