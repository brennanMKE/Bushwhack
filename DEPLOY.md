# Deploying to photon

Bushwhack runs on **photon** (`ssh photon`, t4g.small, arm64, AL2023) as the systemd unit `bushwhack`. It sits behind Apache at `https://bushwhack.sstools.co`. It follows the conventions in `~/Developer/AWS`: `AddSite.md` for the vhosts and wildcard cert, and `Ports.md` for the port registry. It is stateless: no database, no env file, no secrets.

| | |
|---|---|
| Binary | `/usr/local/bin/bushwhack` (static, linux/arm64, UI and patterns embedded) |
| Unit | `bushwhack.service`, user `bushwhack` |
| Port | from `/etc/photon/ports.conf`, binds `127.0.0.1` |
| Vhosts | `001-bushwhack.sstools.co.conf`, `001-bushwhack.sstools.co-le-ssl.conf` |
| Cert | existing `sstools.co-wildcard`. **Don't run certbot.** |
| Logs | `journalctl -u bushwhack`, `/var/log/httpd/sites/bushwhack.sstools.co/` |

## First-time setup

DNS needs nothing: `*.sstools.co` already points at photon. Check it:

```sh
dig +short bushwhack.sstools.co       # 98.84.75.184
```

### 1. Reserve a port

On photon:

```sh
sudo photon-ports next 10             # expect the 8110 block to be free
sudo vi /etc/photon/ports.conf
```

Add:

```
# 8110-8119: sstools.co tools
8110    bushwhack              loopback  bushwhack.sstools.co
```

Then:

```sh
sudo photon-ports render              # writes PORT_BUSHWHACK and the unit's drop-in
```

### 2. User, binary and unit

From the Mac:

```sh
make linux
scp bin/bushwhack-linux-arm64 deploy/bushwhack.service photon:/tmp/
```

On photon:

```sh
sudo useradd --system --no-create-home --shell /sbin/nologin bushwhack
sudo install -m 0755 /tmp/bushwhack-linux-arm64 /usr/local/bin/bushwhack
sudo install -m 0644 /tmp/bushwhack.service /etc/systemd/system/bushwhack.service
sudo restorecon -v /usr/local/bin/bushwhack /etc/systemd/system/bushwhack.service
sudo systemctl daemon-reload
sudo systemctl enable --now bushwhack
systemctl status bushwhack --no-pager
curl -s http://127.0.0.1:8110/healthz
```

### 3. Apache

Put the two vhosts into the AWS repo, so they're mirrored with the others:

```sh
cp deploy/001-bushwhack.sstools.co*.conf ~/Developer/AWS/remote/etc/httpd/conf.d/
```

On photon, create the log directory and the docroot. Apache won't start without the log directory:

```sh
sudo mkdir -p /var/log/httpd/sites/bushwhack.sstools.co
sudo chmod 700 /var/log/httpd/sites/bushwhack.sstools.co
sudo restorecon -Rv /var/log/httpd
sudo mkdir -p /var/www/vhosts/bushwhack.sstools.co
```

From `~/Developer/AWS`:

```sh
./deploy-site.sh bushwhack            # copies both confs, configtest, reload
```

### 4. Verify

```sh
curl -sI https://bushwhack.sstools.co/            # 200
curl -sI http://bushwhack.sstools.co/             # 301 to https
curl -s  https://bushwhack.sstools.co/healthz
ssh photon 'sudo photon-ports check'              # 0 failed
```

Then run `./download.sh` in `~/Developer/AWS` and commit `remote/`.

## Shipping a new build

```sh
./deploy/deploy.sh            # tests, builds linux/arm64, installs, restarts, verifies
./deploy/deploy.sh --dry-run
```

## Limits

The app caps each request at:

- 2 MB upload
- 20k elements, 200k path commands, 500k vertices
- a 10 s deadline
- 4 concurrent jobs, or a 503 after 3 s of waiting

The unit caps the service at `MemoryMax=320M`, `GOMEMLIMIT=200MiB` and `CPUQuota=150%`, so a hostile file can't starve the other services on photon.
