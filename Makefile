# diretório onde o `go install` coloca binários (GOBIN ou GOPATH/bin)
GOBIN := $(shell go env GOBIN)
ifeq ($(GOBIN),)
	GOBIN := $(shell go env GOPATH)/bin
endif

ifeq ($(OS),Windows_NT)
	BINARY := $(GOBIN)/pierrot.exe
else
	BINARY := $(GOBIN)/pierrot
endif

WWW_DIR := www
PKG     := ./cmd/pierrot

.PHONY: install build web test clean deps path

# baixa as dependências do módulo
deps:
	go mod download

# compila e instala o binário em $(GOBIN) (precisa estar no PATH)
install: deps
	go install $(PKG)
	@echo "instalado em $(BINARY)"
	@echo "garanta que '$(GOBIN)' está no PATH (rode 'make path')"

build:
	go build ./...

# gera o site estático em www/build (usado no deploy da Vercel)
web:
	cd $(WWW_DIR) && go run ../cmd/pierrot build

test:
	go test ./...

clean:
ifeq ($(OS),Windows_NT)
	powershell -NoProfile -Command "if (Test-Path '$(BINARY)') { Remove-Item '$(BINARY)' }"
else
	rm -f "$(BINARY)"
endif

# adiciona $(GOBIN) às variáveis de ambiente (PATH)
path:
ifeq ($(OS),Windows_NT)
	powershell -NoProfile -Command "$$p=[Environment]::GetEnvironmentVariable('Path','User'); if ($$p -notlike '*$(GOBIN)*') { [Environment]::SetEnvironmentVariable('Path', $$p + ';$(GOBIN)', 'User'); Write-Host 'PATH atualizado. Reabra o terminal.' } else { Write-Host 'já está no PATH.' }"
else
	@echo "adicione esta linha ao seu ~/.zshrc ou ~/.bashrc:"
	@echo "  export PATH=\"$(GOBIN):$$PATH\""
endif
