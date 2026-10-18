# telegram-gateway-auth

Phone number verification service. Sends a one-time code through the
[Telegram Gateway](https://core.telegram.org/gateway) and falls back to SMS via
[Eskiz](https://eskiz.uz) when Telegram is unavailable or cannot reach the number.

## How it works

1. `POST /v1/send-code` generates a 6-digit code and sends it via Telegram Gateway.
2. If Telegram is unavailable or the number is unreachable, the code is sent by SMS.
3. Only an HMAC-SHA256 hash of the code is stored in Redis/Valkey, with a TTL of 150 seconds.
4. `POST /v1/verify-code` checks the code. 3 attempts are allowed, then the code is discarded.

Sends are rate limited per phone number: a 1-minute cooldown between sends and a cap per hour.

## Run

Requires Docker and a [Telegram Gateway](https://gateway.telegram.org) API token.

Generate a secret key:

```sh
go run ./cmd/genkey
```

Create `.env`:

```env
TOKEN=<telegram gateway token>
SECRET_KEY=<generated key>
HTTP_PORT=8080
REDIS_PASSWORD=<password>
ESKIZ_EMAIL=<optional>
ESKIZ_PASSWORD=<optional>
```

Start:

```sh
docker compose up --build
```

A demo page is served at `http://localhost:8080/`.

## Configuration

| Variable         | Required | Description                                     |
|------------------|----------|-------------------------------------------------|
| `TOKEN`          | yes      | Telegram Gateway API token                      |
| `SECRET_KEY`     | yes      | HMAC key used to hash codes                     |
| `HTTP_ADDR`      | yes      | Listen address, e.g. `:8080`                    |
| `REDIS_ADDR`     | yes      | Redis/Valkey address, e.g. `valkey:6379`        |
| `REDIS_PASSWORD` | no       | Redis/Valkey password                           |
| `ESKIZ_EMAIL`    | no       | Eskiz account email. SMS fallback is disabled if empty |
| `ESKIZ_PASSWORD` | no       | Eskiz account password                          |

## API

### `POST /v1/send-code`

```json
{ "phone": "+998901234567" }
```

### `POST /v1/verify-code`

```json
{ "phone": "+998901234567", "code": "123456" }
```

Success response for both endpoints:

```json
{ "ok": true }
```

Error response:

```json
{ "code": "FLOOD_WAIT", "message": "flood wait: retry after 45s", "retry_after": 45 }
```

`retry_after` (seconds) is present only for `FLOOD_WAIT`, which also sets the `Retry-After` header.

| HTTP | `code`                  | Meaning                                    |
|------|-------------------------|--------------------------------------------|
| 400  | `INVALID_REQUEST`       | Malformed JSON body                        |
| 400  | `INVALID_PHONE_NUMBER`  | Phone number is invalid                    |
| 400  | `INVALID_CODE`          | Wrong code                                 |
| 400  | `CODE_EXPIRED`          | Code expired                               |
| 404  | `OTP_NOT_FOUND`         | No active code for this number             |
| 409  | `MESSAGE_ALREADY_SENT`  | An active code already exists              |
| 422  | `PHONE_NOT_REACHABLE`   | No sender can reach this number            |
| 429  | `FLOOD_WAIT`            | Send rate limit hit                        |
| 429  | `TOO_MANY_REQUESTS`     | Too many requests                          |
| 429  | `MAX_ATTEMPTS_EXCEEDED` | Verification attempts exhausted            |
| 502  | `SENDER_UNAVAILABLE`    | Telegram and SMS are both unavailable      |
| 500  | `INTERNAL_ERROR`        | Unexpected error                           |

## License

[MIT](LICENSE)
