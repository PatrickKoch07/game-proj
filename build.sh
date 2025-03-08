# sudo apt install -y gcc-mingw-w64
rm -f game.exe
export CC=x86_64-w64-mingw32-gcc
export CXX=x86_64-w64-mingw32-g++
GOOS=windows GOARCH=amd64 CGO_ENABLED=1 go build -o game.exe main.go