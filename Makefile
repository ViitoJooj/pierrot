ifeq ($(OS),Windows_NT)
	BIN_DIR := $(USERPROFILE)/.pierrot/bin
	BINARY  := $(BIN_DIR)/pierrot.exe
	MKDIR    = powershell -NoProfile -Command "New-Item -ItemType Directory -Force '$(BIN_DIR)' | Out-Null"
	RM       = powershell -NoProfile -Command "if (Test-Path '$(BINARY)') { Remove-Item '$(BINARY)' }"
else
	BIN_DIR := $(HOME)/.pierrot/bin
	BINARY  := $(BIN_DIR)/pierrot
	MKDIR    = mkdir -p "$(BIN_DIR)"
	RM       = rm -f "$(BINARY)"
endif

WWW_DIR := www

.PHONY: install build web test clean

# compila e instala o binário em ~/.pierrot/bin (precisa estar no PATH)
install:
	$(MKDIR)
	go build -o "$(BINARY)" ./cmd
	@echo instalado em $(BINARY)

build:
	go build ./...

# gera o site estático em www/build (usado no deploy da Vercel)
web:
	cd $(WWW_DIR) && go run ../cmd build

test:
	go test ./...

clean:
	$(RM)
