# Guestbook Server

This directory contains the Go-based backend server for handling guestbook form submissions with advanced spam filtering and GitHub integration.  It is a homebrew replacement for what I was previously using Staticman for, but the Spam filtering capabilities were not maintained anymore when I was using it.

## Features

- Multi-layered spam protection (Akismet, reCAPTCHA v3, honeypot, heuristics, rate limiting)
- Automatically creates pull requests for new guestbook entries in your GitHub repository
- Compatible with Docker and cloud-native deployments

## Build the Docker Image

```sh
docker build . -t guestbook-server
```

## Run the Server Locally

Make sure you have a `.env.local` file with the required environment variables in your project root.

```sh
docker run -v "$(pwd)/.env.local:/root/.env.local" -p 8080:8080 guestbook-server
```

The server will be available at [http://localhost:8080](http://localhost:8080).

## Environment Variables

See [.env.example](.env.example) for required configuration.
