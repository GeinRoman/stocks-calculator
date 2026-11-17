# Release build and deployment

[Русский](release.ru.md)

This guide builds and runs the Go backend, PostgreSQL, migrations, and SvelteKit frontend with Docker Compose on the local machine. It is separate from the development stack: the Compose project is `stocks-calculator-release` and its database uses a separate persistent volume. It does **not** automatically import development data.

The release gateway listens on `127.0.0.1:8080` by default. Configure an existing HTTPS reverse proxy on the same host before exposing the application to other machines. The gateway serves the frontend and forwards `/api/` to the backend, removing the prefix. Neither the database nor the backend is published directly.

## Requirements

- Docker Engine with the Compose plugin on Linux, or Docker Desktop with Compose on Windows.
- Linux: a POSIX shell. Windows: PowerShell and Docker Desktop running with Linux containers.
- A checkout or release source directory containing the frontend, backend, migrations and Dockerfiles.
- Internet access for base images, Go modules, npm packages, and MOEX. The existing external proxy must provide a valid HTTPS certificate for the public hostname.

Commands below run from the project root. The release images are built from source; Node.js and Go are not required on the host.

## Private configuration

Create **both** files in the project root before deploying. They are ignored by Git and excluded from Docker build contexts. Do not commit them, attach them to bug reports, or share their contents.

### `release.env`

Create this file using your own values:

```dotenv
DATABASE_USER=stocksapp
DATABASE_PASSWORD=REPLACE_WITH_RANDOM_HEX_PASSWORD
RELEASE_PORT=8080
```

- `DATABASE_USER` and `DATABASE_PASSWORD` initialize and authenticate to the release database. Use a unique, strong password. The current backend and migration connection strings embed these values without URL escaping: use an ASCII alphanumeric username and a **hexadecimal password**, not punctuation, whitespace, `$`, or URI-special characters.
- `RELEASE_PORT` is the **host loopback** port for the gateway; choose a free port and use the same one in your HTTPS proxy. It is not the backend port.
- Docker Compose gives variables already exported in the shell precedence over `release.env`. Avoid conflicting `DATABASE_USER`, `DATABASE_PASSWORD`, and `RELEASE_PORT` environment variables.
- PostgreSQL uses these credentials only on first initialization of its volume. Changing this file later does **not** change the database password; coordinate credential rotation in PostgreSQL and the application separately.

### `release.config.yaml`

Create this separate backend configuration:

```yaml
max_connections: 100
jwt_key: "REPLACE_WITH_A_DIFFERENT_RANDOM_HEX_KEY"
jwt_expiration_time_min: 10
```

- `jwt_key` signs and verifies authentication tokens. Generate a unique, unpredictable secret of at least 32 random bytes, encoded as hexadecimal. The deployment scripts require this hex value on its own YAML line, optionally enclosed in double quotes. Never use the development key from `config.yaml`, and never reuse the database password. Anyone with this key can forge tokens.
- Keep the key stable across deployments and restarts. Changing it invalidates previously issued tokens and requires users to sign in again. If compromised, rotate it deliberately and protect backups containing it.
- `jwt_expiration_time_min` controls access-token lifetime in minutes. `max_connections` sets the backend database connection limit; select a positive value appropriate for your PostgreSQL capacity.
- The backend reads this file on startup. After editing it, restart the server service as described below. The image does not contain the production key.
- Restrict file access to the deployment administrator; on Linux, for example, `chmod 600 release.env release.config.yaml`. Ensure the Docker daemon can read the mounted configuration. The backend container runs as root so it can read a host-owned file with these permissions.

Generate the password and signing key **separately**. For example, run `openssl rand -hex 32` twice on Linux, or use this PowerShell command twice:

```powershell
$bytes = New-Object byte[] 32
[System.Security.Cryptography.RandomNumberGenerator]::Create().GetBytes($bytes)
-join ($bytes | ForEach-Object { $_.ToString('x2') })
```

Place each result in its corresponding file without quotes in `release.env` and inside the YAML quotes for `jwt_key`. Treat terminal output, history, and backups as sensitive.

## Build and start

Linux:

```sh
./deploy.sh
```

Windows PowerShell:

```powershell
.\deploy.ps1
```

Each script builds the release images, starts PostgreSQL, runs existing up migrations, then starts the backend, frontend and gateway. It returns a nonzero status if Compose fails. It does not delete database data.

For a local-only smoke test, open `http://127.0.0.1:8080` (or the configured port). This is **not** HTTPS and should not be used for remote access or real credentials. Point your existing HTTPS reverse proxy at `http://127.0.0.1:8080`, preserving the request path and host and forwarding the original scheme. Configure its upload limit to at least 33 MiB for snapshot imports, and confirm both the site and `/api/` are forwarded to the gateway. Terminate TLS at that proxy; do not expose the loopback HTTP port externally. Refresh tokens are stored in browser local storage, so HTTPS and careful access control are important.

For example, an existing host Nginx installation could use the following server block after replacing the hostname, certificate paths and, if needed, port. Obtain and renew certificates separately, validate the configuration, and reload Nginx using your host's normal procedure:

```nginx
server {
    listen 443 ssl;
    server_name portfolio.example.com;
    ssl_certificate /path/to/fullchain.pem;
    ssl_certificate_key /path/to/private-key.pem;
    client_max_body_size 33m;

    location / {
        proxy_pass http://127.0.0.1:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    }
}
```

## Operations

Once the images have been built using the deployment script, start the release stack without rebuilding:

```sh
./start.sh
```

```powershell
.\start.ps1
```

To start it if needed and then restart the backend, frontend, and gateway, use `./start.sh restart` on Linux or `.\start.ps1 restart` on Windows. PostgreSQL is left running; the scripts do not remove volumes. Both actions ensure the database is healthy and up migrations have completed before application services start. These scripts do not build changed source: rerun the deployment script for updates.

Use the same project and env file for every Compose command:

```sh
docker compose --env-file release.env -p stocks-calculator-release -f compose.release.yaml ps
docker compose --env-file release.env -p stocks-calculator-release -f compose.release.yaml logs --tail=100
```

After updating source, back up data and rerun the deployment script. Migrations are applied before the server starts; they are not automatically rolled back if a later step fails. After changing `release.config.yaml`, restart the server:

```sh
docker compose --env-file release.env -p stocks-calculator-release -f compose.release.yaml restart server
```

To stop the stack without deleting data:

```sh
docker compose --env-file release.env -p stocks-calculator-release -f compose.release.yaml down
```

Never add `-v` unless you deliberately intend to delete the release database. Before upgrades, back up PostgreSQL using your usual tested database backup procedure and store the dump and private configuration securely. The application's profile tab can also export a JSON snapshot of an individual account; this is not a substitute for a full database backup. Test restoration on a separate disposable instance, never against production without a recovery plan.

If startup fails, check `ps` and service logs, confirm the database credentials and configuration file, and ensure the selected port is free. A failed migration prevents server startup; investigate and back up before repairing migration state. Network or MOEX failures can affect market-data requests. Do not paste credentials or tokens into logs or support requests.
