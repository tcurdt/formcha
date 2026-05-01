# formcha

`formcha` receives HTML form submissions, verifies the [ALTCHA](https://altcha.org) captcha, and dispatches the data to one or more notification backends.

## How it works

1. Your page loads the ALTCHA widget, which fetches a challenge from formcha.
2. The user solves the challenge client-side (no external service needed).
3. On submit, the form POSTs to formcha with the solved captcha token.
4. formcha verifies the token and forwards the form data to every configured backend.

## Endpoints

| Method | Path                  | Description                                            |
| ------ | --------------------- | ------------------------------------------------------ |
| `GET`  | `/ping`               | Health check — returns `pong`                          |
| `GET`  | `/metrics`            | Prometheus metrics                                     |
| `GET`  | `/altcha`             | Challenge endpoint for the ALTCHA widget               |
| `POST` | `/submit`             | Form submission with standard ALTCHA proof-of-work     |
| `POST` | `/submit_spam_filter` | Form submission with ALTCHA server-side spam filtering |

## HTML form integration

Add the [ALTCHA widget](https://altcha.org/docs/website-integration/) to your form and point both the `challengeurl` and the form `action` at your formcha instance.

```html
<script
  async
  defer
  src="https://unpkg.com/altcha/dist/altcha.min.js"
  type="module"
></script>

<form action="https://formcha.example.com/submit" method="POST">
  <label>
    Name
    <input type="text" name="name" required />
  </label>
  <label>
    Email
    <input type="email" name="email" required />
  </label>
  <label>
    Message
    <textarea name="message" required></textarea>
  </label>

  <!-- ALTCHA widget — fetches challenge from the same formcha instance -->
  <altcha-widget
    challengeurl="https://formcha.example.com/altcha"
  ></altcha-widget>

  <button type="submit">Send</button>
</form>
```

Use `/submit_spam_filter` instead of `/submit` if you have ALTCHA server-side spam filtering enabled on your account.

## Configuration

All configuration is done through environment variables.

### Core

| Variable               | Required | Description                                                              |
| ---------------------- | -------- | ------------------------------------------------------------------------ |
| `ALTCHA_HMAC_KEY`      | yes      | Secret key used to sign and verify ALTCHA challenges                     |
| `PORT`                 | no       | Port to listen on (default: `3000`)                                      |
| `FORMCHA_IDLE_TIMEOUT` | no       | Shut down after this period of inactivity, e.g. `5m` (default: disabled) |

### Backends

Each backend is enabled automatically when its required variables are set. You can enable multiple backends at once.

#### Log to stdout

Always active — every submission is printed to standard output. No configuration required.

#### Webhook

Posts form data as JSON to an HTTP endpoint.

| Variable      | Required | Description              |
| ------------- | -------- | ------------------------ |
| `WEBHOOK_URL` | yes      | URL to POST form data to |

#### SMTP

Sends an email via a plain SMTP server.

| Variable        | Required | Description                                      |
| --------------- | -------- | ------------------------------------------------ |
| `SMTP_HOST`     | yes      | SMTP server hostname                             |
| `SMTP_PORT`     | yes      | SMTP server port (`25`, `587`, or `465` for TLS) |
| `SMTP_FROM`     | yes      | Sender address                                   |
| `SMTP_TO`       | yes      | Recipient address                                |
| `SMTP_USERNAME` | no       | SMTP username (if authentication is required)    |
| `SMTP_PASSWORD` | no       | SMTP password (if authentication is required)    |

#### Brevo

Sends an email via the [Brevo](https://brevo.com) transactional email API.

| Variable             | Required | Description                    |
| -------------------- | -------- | ------------------------------ |
| `BREVO_API_KEY`      | yes      | Brevo API key                  |
| `BREVO_SENDER_EMAIL` | yes      | Verified sender address        |
| `BREVO_TO_EMAIL`     | yes      | Recipient address              |
| `BREVO_SENDER_NAME`  | no       | Display name for the sender    |
| `BREVO_TO_NAME`      | no       | Display name for the recipient |

#### Pushover

Sends a push notification via [Pushover](https://pushover.net).

| Variable            | Required | Description                |
| ------------------- | -------- | -------------------------- |
| `PUSHOVER_TOKEN`    | yes      | Pushover application token |
| `PUSHOVER_USER_KEY` | yes      | Pushover user or group key |

#### ntfy

Sends a push notification via [ntfy](https://ntfy.sh) (self-hosted or ntfy.sh).

| Variable     | Required | Description                                     |
| ------------ | -------- | ----------------------------------------------- |
| `NTFY_URL`   | yes      | Full topic URL, e.g. `https://ntfy.sh/my-topic` |
| `NTFY_TOKEN` | no       | Access token for protected topics               |

## Running

```sh
ALTCHA_HMAC_KEY=changeme \
NTFY_URL=https://ntfy.sh/my-topic \
./formcha
```

### Systemd socket activation

formcha supports systemd socket activation. When launched by systemd with a socket unit, it will use the provided socket instead of binding to `PORT`.
