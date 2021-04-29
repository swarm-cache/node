go build -o Test1.exe

start "" "Test1.exe" --cl=127.0.0.1:3666 --nl=127.0.0.1:40000 --log-io-msg=1
timeout /t 1
start "" "Test1.exe" --nl=127.0.0.1:40001 --nc=127.0.0.1:40000 --log-io-msg=1
timeout /t 1
start "" "Test1.exe" --cl=127.0.0.1:3667 --nc=127.0.0.1:40001,127.0.0.1:40000 --log-io-msg=1

