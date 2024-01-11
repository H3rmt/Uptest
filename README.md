# Uptest
Simple go program to check if Website is up

## Example
docker-compose.yml
```
version: '3'
services:
  uptest:
    image: image: h3rmt/uptest
    ports:
      - 8080:80
    volumes:
      - ./logs:/logs
      - ./responses:/responses
    environment:
      - URLS=www.google.com=google,www.yahoo.com=yahoo
```
