<p align="center">
  <img src="img/ico.png" width="200" alt="kwach logo">
</p>

<h1 align="center">kwach</h1>

<p align="center">
  <a href="https://github.com/oyamo/kwach/actions/workflows/ci.yml"><img src="https://img.shields.io/github/actions/workflow/status/oyamo/kwach/ci.yml?style=for-the-badge&label=CI" alt="CI"></a>
  <a href="https://codecov.io/gh/oyamo/kwach"><img src="https://img.shields.io/codecov/c/github/oyamo/kwach?style=for-the-badge" alt="Coverage"></a>
  <a href="go.mod"><img src="https://img.shields.io/github/go-mod/go-version/oyamo/kwach?style=for-the-badge&logo=go&logoColor=white" alt="Go version"></a>
  <a href="LICENSE"><img src="https://img.shields.io/github/license/oyamo/kwach?style=for-the-badge" alt="License"></a>
</p>

kwach is an AI harness for voice roasts. It builds a prompt from a user's profile and recent headlines, checks the generated script with a guard model, then turns it into speech. The chat and guard models use the OpenAI API format, so you can switch providers with config. Profiles live in MongoDB and chat history in Redis. The front end runs the intake interview by voice.

<p align="center">
  <img src="img/screenshot.png" alt="kwach interview screen">
</p>

## Technology

<div align="center">

<table>
  <tr>
    <td align="center" valign="middle" width="180" height="64"><img src="img/logo/go.svg" height="32" alt="Go"><br>Go</td>
    <td align="center" valign="middle" width="180" height="64"><img src="img/logo/docker.svg" height="32" alt="Docker"><br>Docker</td>
    <td align="center" valign="middle" width="180" height="64"><img src="img/logo/groq.svg" height="32" alt="Groq"><br>Groq</td>
  </tr>
  <tr>
    <td align="center" valign="middle" width="180" height="64"><img src="https://eleven-public-cdn.elevenlabs.io/payloadcms/elevenlabs-official-logo.svg" height="32" alt="ElevenLabs"><br>ElevenLabs</td>
    <td align="center" valign="middle" width="180" height="64"><img src="img/logo/mongodb.svg" height="32" alt="MongoDB"><br>MongoDB</td>
    <td align="center" valign="middle" width="180" height="64"><img src="img/logo/redis.svg" height="32" alt="Redis"><br>Redis</td>
  </tr>
  <tr>
    <td align="center" valign="middle" width="180" height="64"><img src="img/logo/googlenews.svg" height="32" alt="Google News"><br>Google News</td>
    <td align="center" valign="middle" width="180" height="64"><img src="img/logo/nginx.svg" height="32" alt="nginx"><br>nginx</td>
    <td align="center" valign="middle" width="180" height="64"><img src="img/logo/codecov.png" height="24" alt="Codecov"><br>Codecov</td>
  </tr>
</table>

</div>

## Run

```sh
go build -o kwach ./cmd/kwach
./kwach serve
```

## Docker

```sh
docker build -t kwach .
docker run --rm -p 8080:8080 \
  --read-only --cap-drop=ALL --security-opt=no-new-privileges:true \
  --pids-limit=100 --memory=256m \
  -e OPENAI_BASE_URL -e OPENAI_API_KEY -e OPENAI_MODEL -e ELEVENLABS_API_KEY -e ELEVENLABS_VOICE_IDS -e MONGODB_URI -e REDIS_URL \
  kwach
```

## Development

```sh
go vet ./...
go test ./...
```

## License

[MIT](LICENSE) © 2026 Oyamo
