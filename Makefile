APP := wigglegram-maker
BIN := bin

ifeq ($(OS),Windows_NT)
EXE := .exe
LDFLAGS := -ldflags="-H=windowsgui"
MKDIR_BIN := if not exist $(BIN) mkdir $(BIN)
RM_BIN := if exist $(BIN) rmdir /s /q $(BIN)
else
EXE :=
LDFLAGS :=
MKDIR_BIN := mkdir -p $(BIN)
RM_BIN := rm -rf $(BIN)
endif

.PHONY: build run test fmt package-windows clean

build:
	$(MKDIR_BIN)
	go build $(LDFLAGS) -o $(BIN)/$(APP)$(EXE) .

run:
	go run .

test:
	go test ./...

fmt:
	gofmt -w main.go resources.go models/*.go processor/*.go store/*.go ui/*.go

package-windows: build
	fyne package -os windows -executable $(BIN)/$(APP)$(EXE) -icon assets/icon.png -name "$(APP)"

clean:
	$(RM_BIN)
