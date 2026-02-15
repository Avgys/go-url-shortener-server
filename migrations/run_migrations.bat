@echo off
echo ComSpec=[%ComSpec%]
echo cmdline=[%cmdcmdline%]
echo DSN=%DATABASE_DSN%
migrate --database "%DATABASE_DSN%" --path . up
pause