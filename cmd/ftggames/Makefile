GO = go
DESTDIR = ../../../
DESTNAME = server
TARGET = $(DESTDIR)$(DESTNAME)
LDFLAGS = -s -w -extldflags -static
GOFLAGS = -trimpath -buildmode exe

build:
	$(GO) env -w CGO_ENABLED=0
	$(GO) build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o $(TARGET).tmp

build-release:
	$(GO) env -w CGO_ENABLED=0
	$(GO) build -tags release $(GOFLAGS) -ldflags "$(LDFLAGS)" -o $(TARGET)

run: build
	echo "cd $(DESTDIR); ./$(DESTNAME).tmp" | bash
