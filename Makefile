DEST?=bin
GOOS?=linux
GOARCH?=arm
GOARM?=6
TARGET?=stm32f4disco
FLAGS=-ldflags "-w"

GOENV=GOOS=$(GOOS) GOARCH=$(GOARCH) $(if $(GOARM),GOARM=$(GOARM))
ENV=env $(GOENV)

.PHONY: all camera-reader clean go-binaries motor-control pi-zero pi-zero-2 pi-zero-2-64 sonar-reader test tinygo-hello tinygo-sonar web-bridge

all: pi-zero tinygo-sonar tinygo-hello

go-binaries: motor-control sonar-reader camera-reader web-bridge

pi-zero: GOARCH=arm
pi-zero: GOARM=6
pi-zero: go-binaries

pi-zero-2: GOARCH=arm
pi-zero-2: GOARM=7
pi-zero-2: go-binaries

pi-zero-2-64: GOARCH=arm64
pi-zero-2-64: GOARM=
pi-zero-2-64: go-binaries

motor-control:
	$(ENV) go build $(FLAGS) -o $(DEST)/motor-control cmd/motor-control/main.go

sonar-reader:
	$(ENV) go build $(FLAGS) -o $(DEST)/sonar-reader cmd/sonar-reader/main.go

camera-reader:
	$(ENV) go build $(FLAGS) -o $(DEST)/camera-reader cmd/camera-reader/main.go

tinygo-sonar:
	tinygo build -target=$(TARGET) -o $(DEST)/sonar-$(TARGET).bin ./firmware/sonar

tinygo-hello:
	tinygo build -target=$(TARGET) -o $(DEST)/hello-$(TARGET).bin ./firmware/hello

web-bridge:
	$(ENV) go build $(FLAGS) -o $(DEST)/web-bridge cmd/web-bridge/main.go
	@rm -rf $(DEST)/static
	cp -R cmd/web-bridge/static $(DEST)/static

test:
	go test ./...

clean:
	rm -rf ${DEST}/{motor-control,sonar-reader,camera-reader,sonar-*.bin,hello-*.bin,web-bridge,web}
