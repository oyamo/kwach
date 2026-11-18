<h1 align="center">kwach</h1>

<p align="center">
  <img src="https://img.shields.io/badge/Go-1.26-00ADD8?logo=go&logoColor=white" alt="Go 1.26">
  <img src="https://img.shields.io/badge/license-MIT-d6402b" alt="MIT license">
  <img src="https://img.shields.io/badge/status-prototype-6b6861" alt="Status: prototype">
</p>

<p align="center">
  A Go service that generates a short roast from a user's profile and recent news,
  screens it with a safety model, converts it to speech, and stores profiles and chat history for context.
</p>

## Technology

<table align="center">
  <tr>
    <td align="center" width="120"><img src="img/logo/go.svg" width="40" alt="Go"><br>Go</td>
    <td align="center" width="120"><img src="img/logo/docker.svg" width="40" alt="Docker"><br>Docker</td>
    <td align="center" width="120"><img src="img/logo/groq.svg" height="24" alt="Groq"><br>Groq</td>
    <td align="center" width="120">
      <img src="https://eleven-public-cdn.elevenlabs.io/payloadcms/elevenlabs-official-logo.svg" width="110" alt="ElevenLabs"><br>ElevenLabs
    </td>
  </tr>
  <tr>
    <td align="center" width="120"><img src="img/logo/mongodb.svg" width="40" alt="MongoDB"><br>MongoDB</td>
    <td align="center" width="120"><img src="img/logo/redis.svg" width="40" alt="Redis"><br>Redis</td>
    <td align="center" width="120"><img src="img/logo/googlenews.svg" width="40" alt="Google News"><br>Google News</td>
    <td align="center" width="120"><img src="img/logo/nginx.svg" width="40" alt="nginx"><br>nginx</td>
  </tr>
</table>

## Quick start

```sh
export $(grep -v '^#' configs/.env.example | xargs)   # example values; replace with real ones
go run ./cmd/kwach serve                             # JSON API on HTTP_ADDR (default :8080)
go run ./cmd/kwach narrate "your prompt" -out roast.mp3
```

Configuration comes only from environment variables. See [`configs/.env.example`](configs/.env.example) for the full list. The front-end in `web/static/` is served by nginx.

## Docker

```sh
docker build -t kwach .
docker run --rm -p 8080:8080 \
  --read-only --cap-drop=ALL --security-opt=no-new-privileges:true \
  --pids-limit=100 --memory=256m \
  -e GROQ_API_KEY -e ELEVENLABS_API_KEY -e ELEVENLABS_VOICE_ID -e MONGODB_URI -e REDIS_ADDR \
  kwach
```

The `-e NAME` flags pass values from your shell. The image is a static binary on distroless, runs as UID 65532, has no shell, and pins its base images by digest.

## Development

```sh
go vet ./...
go test ./...
```

## License

[MIT](LICENSE) © 2026 Oyamo
