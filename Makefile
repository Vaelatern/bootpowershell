.PHONY: build

build: bootpowershell bootpowershell.exe

bootpowershell.exe: main.go
	GOOS=windows go build .

bootpowershell: main.go
	go build .

clean:
	rm -f bootpowershell bootpowershell.exe
