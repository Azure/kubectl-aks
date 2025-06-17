INSTALLATION_DIR  ?= ~/.local/bin
RELEASE ?= latest

# Install innovation engine
.PHONY: install-ie
install-ie:
	wget -q -O ie https://github.com/Azure/InnovationEngine/releases/download/$(RELEASE)/ie >/dev/null
	chmod +x ie >/dev/null
	mkdir -p $(INSTALLATION_DIR) >/dev/null
	mv ie $(INSTALLATION_DIR) >/dev/null
	if ! echo "$$PATH" | grep -q "$(INSTALLATION_DIR)"; then \
        export PATH="$$PATH:$(INSTALLATION_DIR)"; \
    fi;>/dev/null \

# Uninstall innovation engine
.PHONY: uninstall-ie
uninstall-ie:
	rm -f $(INSTALLATION_DIR)/ie
