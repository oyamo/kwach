<p align="center">
  <img src="img/ico.png" width="200" alt="kwach logo">
</p>

# kwach

[![CI](https://img.shields.io/github/actions/workflow/status/oyamo/kwach/ci.yml?style=for-the-badge&label=CI)](https://github.com/oyamo/kwach/actions/workflows/ci.yml)
[![Coverage](https://img.shields.io/codecov/c/github/oyamo/kwach?style=for-the-badge)](https://codecov.io/gh/oyamo/kwach)
[![Go Reference](https://img.shields.io/badge/Go%20Reference-pkg.go.dev-00ADD8?style=for-the-badge&logo=go&logoColor=white)](https://pkg.go.dev/github.com/oyamo/kwach)
[![Go Report Card](https://goreportcard.com/badge/github.com/oyamo/kwach?style=for-the-badge)](https://goreportcard.com/report/github.com/oyamo/kwach)
[![Go version](https://img.shields.io/github/go-mod/go-version/oyamo/kwach?style=for-the-badge&logo=go&logoColor=white)](go.mod)
[![License](https://img.shields.io/github/license/oyamo/kwach?style=for-the-badge)](LICENSE)

A Go service that generates a short roast from a user's profile and recent news,
screens it with a safety model, converts it to speech, and stores profiles and chat history for context.

## Technology

<table>
  <tr>
    <td valign="bottom"><img src="img/logo/go.svg" width="40" alt="Go"><br>Go</td>
    <td valign="bottom"><img src="img/logo/docker.svg" width="40" alt="Docker"><br>Docker</td>
    <td valign="bottom"><img src="img/logo/groq.svg" height="24" alt="Groq"><br>Groq</td>
    <td valign="bottom"><img src="https://eleven-public-cdn.elevenlabs.io/payloadcms/elevenlabs-official-logo.svg" width="110" alt="ElevenLabs"><br>ElevenLabs</td>
  </tr>
  <tr>
    <td valign="bottom"><img src="img/logo/mongodb.svg" width="40" alt="MongoDB"><br>MongoDB</td>
    <td valign="bottom"><img src="img/logo/redis.svg" width="40" alt="Redis"><br>Redis</td>
    <td valign="bottom"><img src="img/logo/googlenews.svg" width="40" alt="Google News"><br>Google News</td>
    <td valign="bottom"><img src="img/logo/nginx.svg" width="40" alt="nginx"><br>nginx</td>
  </tr>
</table>

## Run

Build the binary and provide the configuration through the environment: your service manager, secret store, or container runtime. Never use the placeholder values in `configs/.env.example` for a real deployment.

```sh
go build -o kwach ./cmd/kwach
./kwach serve   # JSON API on HTTP_ADDR (default :8080)
```

The app reads only environment variables. See [`configs/.env.example`](configs/.env.example) for the full list of names. The front-end in `web/static/` is served by nginx.

For a one-off roast from the command line: `./kwach narrate "your prompt" -out roast.mp3`.

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
