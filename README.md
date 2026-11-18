<p align="center">
  <img src="img/ico.png" width="200" alt="kwach logo">
</p>

# kwach

<p align="center">
  <a href="https://github.com/oyamo/kwach/actions/workflows/ci.yml"><img src="https://img.shields.io/github/actions/workflow/status/oyamo/kwach/ci.yml?style=for-the-badge&label=CI" alt="CI"></a>
  <a href="https://codecov.io/gh/oyamo/kwach"><img src="https://img.shields.io/codecov/c/github/oyamo/kwach?style=for-the-badge" alt="Coverage"></a>
  <a href="go.mod"><img src="https://img.shields.io/github/go-mod/go-version/oyamo/kwach?style=for-the-badge&logo=go&logoColor=white" alt="Go version"></a>
  <a href="LICENSE"><img src="https://img.shields.io/github/license/oyamo/kwach?style=for-the-badge" alt="License"></a>
</p>

A Go service that generates a short roast from a user's profile and recent news,
screens it with a safety model, converts it to speech, and stores profiles and chat history for context.

## Technology

<div align="center">

<table>
  <tr>
    <td align="center" valign="middle" width="140" height="64"><img src="img/logo/go.svg" height="32" alt="Go"><br>Go</td>
    <td align="center" valign="middle" width="140" height="64"><img src="img/logo/docker.svg" height="32" alt="Docker"><br>Docker</td>
    <td align="center" valign="middle" width="140" height="64"><img src="img/logo/groq.svg" height="32" alt="Groq"><br>Groq</td>
    <td align="center" valign="middle" width="140" height="64"><img src="https://eleven-public-cdn.elevenlabs.io/payloadcms/elevenlabs-official-logo.svg" height="32" alt="ElevenLabs"><br>ElevenLabs</td>
  </tr>
  <tr>
    <td align="center" valign="middle" width="140" height="64"><img src="img/logo/mongodb.svg" height="32" alt="MongoDB"><br>MongoDB</td>
    <td align="center" valign="middle" width="140" height="64"><img src="img/logo/redis.svg" height="32" alt="Redis"><br>Redis</td>
    <td align="center" valign="middle" width="140" height="64"><img src="img/logo/googlenews.svg" height="32" alt="Google News"><br>Google News</td>
    <td align="center" valign="middle" width="140" height="64"><img src="img/logo/nginx.svg" height="32" alt="nginx"><br>nginx</td>
  </tr>
</table>

</div>

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
