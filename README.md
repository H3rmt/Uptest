# Uptest
Simple go program to check if Website is up

## Example
docker-compose.yml
```
version: '3'
services:
  uptest:
    image: h3rmt/uptest
    ports:
      - 8080:8080
    volumes:
      - ./logs:/logs
      - ./responses:/responses
      # mount timezone file
      - /etc/localtime:/etc/localtime:ro
    environment:
      - URLS="www.google.com:google,www.yahoo.com?s=1test: yahoo"
```
