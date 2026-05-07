@echo off
echo DSN=%DATABASE_DSN%
migrate --database "%DATABASE_DSN%" --path .\.. up
pause