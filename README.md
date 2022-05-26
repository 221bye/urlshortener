# urlshortener
A simple url shortener written in Go that uses Redis for storing data.

## Usage
Clone the repository:

> git clone https://github.com/221bye/urlshortener.git

Install and run redis-server:

Download from official website (https://redis.io/download/)
OR pull docker image and run in the container. (https://hub.docker.com/_/redis)

And finally:

> go run main.go

The server runs on localhost:8080.