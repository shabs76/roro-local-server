# Installing the RoRo local server on a Raspberry Pi 4

This guide installs the local server on a Raspberry Pi 4, straight from GitHub. When you finish:

- the server answers the tablets on port **4400**,
- MariaDB keeps the data,
- phpMyAdmin runs on port **8080** for administrators,
- everything starts again by itself after a reboot or a power cut.

All three parts run in Docker, so the Pi needs no Go or MariaDB installation of its own.

You need basic command-line skills: typing commands over SSH and editing a file with `nano`. Replace example values such as `192.168.1.50` or `walls` with your own.

## What you need

| Item | Notes |
|---|---|
| Raspberry Pi 4, 4 GB or 8 GB RAM | A 2 GB Pi can run the server, but building it is slow. |
| Official USB-C power supply (5.1 V, 3 A) | A weak supply causes under-voltage, which corrupts the SD card and the database. |
| USB 3 SSD, 128 GB or more (recommended) | Photos and videos add up after every operation. SD cards also wear out under constant database writes. A good microSD card (A2, 64 GB or more) works for a trial. |
| Case with a heatsink or fan | The Pi slows down when it overheats. |
| Network cable to the router the tablets use | Wi-Fi also works; a cable is more reliable. |
| Internet access | Needed to install and at every start, when the server pulls users from the remote server. Tablet logins also go through the remote server (see step 10). |
| Remote login of the initial user | Email and password of a remote account that may read the company's users. |
| A laptop | To prepare the disk and to connect over SSH. |
| Optional: a second USB drive | For backups (step 12). |
| Optional: a small UPS | Protects the database against power cuts. |

## 1. Prepare the system disk

1. On the laptop, install **Raspberry Pi Imager** from https://www.raspberrypi.com/software/.
2. Connect the SSD (through its USB adapter) or the microSD card to the laptop.
3. In Raspberry Pi Imager, choose:
   - **Device:** Raspberry Pi 4.
   - **Operating system:** *Raspberry Pi OS (other)*, then **Raspberry Pi OS Lite (64-bit)**.
     The 64-bit version is required, because the server is built for 64-bit ARM. Lite has no desktop, which leaves more memory for the server.
   - **Storage:** the SSD or the card.
4. When the Imager offers OS customisation, open **Edit settings** and set:
   - **Hostname:** `roro-server`.
   - **Username and password:** for example `walls`, with a strong password.
   - **Wireless LAN:** only if you cannot use a cable.
   - **Locale settings:** your time zone, for example `Africa/Dar_es_Salaam`.
   - **Services tab:** enable SSH with password authentication.
5. Write the disk, then connect it to the Pi. With an SSD, use one of the blue USB 3 ports.

A Pi 4 with an up-to-date bootloader starts from USB on its own. If it does not start from the SSD:

1. Start it once from a microSD card.
2. Run `sudo rpi-eeprom-update -a`, then `sudo raspi-config`.
3. Choose **Advanced Options**, then **Boot Order**, then **USB Boot**.

## 2. First start and system update

1. Connect the network cable and the power supply. Wait about two minutes.
2. From the laptop, log in:

   ```bash
   ssh walls@roro-server.local
   ```

   If the name is not found, look up the Pi's address in the router's list of connected devices and use `ssh walls@<address>`.
3. Update the system and restart:

   ```bash
   sudo apt update && sudo apt full-upgrade -y
   sudo reboot
   ```

4. Log in again and check that the system is 64-bit:

   ```bash
   uname -m
   ```

   It must print `aarch64`. If it prints `armv7l`, the 32-bit system was installed. Go back to step 1 and choose the 64-bit system.

## 3. Give the Pi a fixed address

Each tablet stores the server address. If the Pi gets a new address, the tablets lose the server.

**Preferred: reserve the address in the router.**

1. Find the Pi's network card address:

   ```bash
   ip link show eth0
   ```

   Use `wlan0` instead of `eth0` on Wi-Fi. The address is the value after `link/ether`.
2. In the router's DHCP settings, reserve one IP address for that MAC address.
3. Run `sudo reboot`.

**If you cannot change the router, set the address on the Pi.**

1. Find the connection name:

   ```bash
   nmcli -t -f NAME,DEVICE connection show
   ```

2. Set the address. In this example the connection is `Wired connection 1`, the address is `192.168.1.50`, and the router is `192.168.1.1`:

   ```bash
   sudo nmcli connection modify "Wired connection 1" \
     ipv4.method manual \
     ipv4.addresses 192.168.1.50/24 \
     ipv4.gateway 192.168.1.1 \
     ipv4.dns "1.1.1.1 8.8.8.8"
   sudo nmcli connection up "Wired connection 1"
   ```

   Pick an address outside the router's DHCP range. A wrong value cuts the SSH connection. If that happens, fix the value from a keyboard and screen attached to the Pi.

Check the address with `hostname -I`. This guide uses `192.168.1.50` from here on.

## 4. Clock and time zone

The Pi 4 has no battery-backed clock. After a power cut, it restarts with the time of its last shutdown. It corrects the time only when it reaches the internet.

The wrong time causes two problems:
- inspection times, history and the tablets' live status updates get wrong times;
- the connection to the remote server (HTTPS) fails.

1. Set the time zone:

   ```bash
   sudo timedatectl set-timezone Africa/Dar_es_Salaam
   ```

2. Check the clock:

   ```bash
   timedatectl
   ```

   Check two lines:
   - `System clock synchronized: yes`;
   - the local time is correct.

If the Pi often starts without internet, fit a real-time clock module, for example a DS3231:

1. Add `dtoverlay=i2c-rtc,ds3231` to `/boot/firmware/config.txt`.
2. Reboot.

## 5. Install Docker

1. Install Docker from Docker's own package repository, so you get the current version with the Compose plugin:

   ```bash
   sudo apt install -y ca-certificates curl git
   sudo install -m 0755 -d /etc/apt/keyrings
   sudo curl -fsSL https://download.docker.com/linux/debian/gpg -o /etc/apt/keyrings/docker.asc
   sudo chmod a+r /etc/apt/keyrings/docker.asc
   echo "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.asc] https://download.docker.com/linux/debian $(. /etc/os-release && echo "$VERSION_CODENAME") stable" \
     | sudo tee /etc/apt/sources.list.d/docker.list > /dev/null
   sudo apt update
   sudo apt install -y docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin
   ```

2. Limit the size of Docker's own logs, so they cannot fill the disk:

   ```bash
   sudo tee /etc/docker/daemon.json > /dev/null <<'EOF'
   {
     "log-driver": "json-file",
     "log-opts": { "max-size": "10m", "max-file": "3" }
   }
   EOF
   sudo systemctl restart docker
   ```

3. Let your user run Docker without `sudo`:

   ```bash
   sudo usermod -aG docker $USER
   exit
   ```

4. Log in again over SSH and check the installation:

   ```bash
   docker run --rm hello-world
   docker compose version
   ```

Docker starts automatically at boot.

## 6. Download the server from GitHub

```bash
cd ~
git clone https://github.com/shabs76/roro-local-server.git
cd roro-local-server
```

The repository is public, so no GitHub login is needed. Run all later commands from `~/roro-local-server`.

## 7. Configure

The server uses two settings files. Both hold passwords. Git ignores them, so they are never uploaded to GitHub.

| File | Used by |
|---|---|
| `.env`, next to `docker-compose.yml` | the MariaDB and phpMyAdmin containers |
| `src/.env` | the server |

### 7.1 Create two passwords

```bash
openssl rand -hex 16    # database root password
openssl rand -hex 16    # password of the server's database account
```

Store both in your password manager.

### 7.2 Database settings: `.env`

```bash
cp .env.example .env
nano .env
```

Fill in the values:

```ini
MYSQL_ROOT_PASSWORD=<root password>
MYSQL_DATABASE=roro_local
MYSQL_USER=roro
MYSQL_PASSWORD=<server account password>
TZ=Africa/Dar_es_Salaam
```

### 7.3 Server settings: `src/.env`

```bash
cp src/.env.example src/.env
nano src/.env
```

Fill in the values:

```ini
MYSQL_HOST=localhost
MYSQL_PORT=3306
MYSQL_USER=roro
MYSQL_PASS=<server account password, the same as MYSQL_PASSWORD in .env>
MYSQL_DBNAME=roro_local

REMOTE_SERVER_URL=https://api.walls.co.tz

INIT_USER_EMAIL=<remote email of the initial user>
INIT_USER_PASSWORD=<remote password of the initial user>

PORT=4400
LOG_DIR=logs
APP_ENV=production
PUBLISH_ITEM_WORKERS=3
PUBLISH_MEDIA_WORKERS=6
```

Leave `MYSQL_HOST` as it is: Docker Compose points the server at the database container itself.

These values must match between the two files:

| `.env` | `src/.env` |
|---|---|
| `MYSQL_USER` | `MYSQL_USER` |
| `MYSQL_PASSWORD` | `MYSQL_PASS` |
| `MYSQL_DATABASE` | `MYSQL_DBNAME` |

**The initial user.** The server signs in to the remote server as this user at every start. It then pulls the roles, users and reference lists: makers, body types, models, checklist, package types and clients. Until the first pull succeeds, nobody can log in on the local server. Use an account that may read the company's users.

### 7.4 Protect the files

```bash
chmod 600 .env src/.env
```

## 8. Build and start

```bash
docker compose up -d --build
```

The first build downloads the base images and compiles the server on the Pi. It can take half an hour or more. Later builds are faster.

On the first start, these steps happen in order:

1. MariaDB creates the `roro_local` database and the `roro` account. It then loads `roro_local.sql`: all tables, no data.
2. The server waits until the database is ready.
3. The server applies its migrations and starts listening on port 4400.
4. In the background, the server signs in to the remote server as the initial user and pulls users and lists.

Check that the three containers run:

```bash
docker compose ps
```

You should see:
- `roro-mariadb` marked `(healthy)`;
- `roro_app` and `roro_phpmyadmin` marked `Up`.

## 9. Check the installation

1. Check that the server answers:

   ```bash
   curl http://localhost:4400/health
   ```

   Expected: `{"status":"ok","time":...}`.

2. Check the user pull:

   ```bash
   docker compose logs app | grep -i "initial user"
   ```

   | Message | Meaning and action |
   |---|---|
   | `Initial user sync: roles and users pulled from the remote server` | Done. Users can log in. |
   | `Initial user sync failed; retrying` | The remote server cannot be reached. The server retries up to 8 times over about 30 minutes. Check the internet connection and the clock (step 4). Then run `docker compose restart app`. |
   | `Initial user sync stopped: the remote server refused the initial user's login` | `INIT_USER_EMAIL` or `INIT_USER_PASSWORD` is wrong. Fix `src/.env`, then run `docker compose up -d --force-recreate app`. |
   | `No users on this server and no initial user configured` | `INIT_USER_EMAIL` and `INIT_USER_PASSWORD` are empty. Fill them in, then recreate the app as above. |

   After a change to `src/.env`, run `docker compose up -d --force-recreate app`. A plain `restart` keeps the old settings.

3. Count the users the server received:

   ```bash
   docker exec roro-mariadb sh -c 'mariadb -uroot -p"$MYSQL_ROOT_PASSWORD" roro_local -e "SELECT COUNT(*) FROM users"'
   ```

4. From a laptop or tablet on the same network, open `http://192.168.1.50:4400/health` in a browser. You should see the same `ok` answer as in point 1.

## 10. Connect the tablets

On each tablet, in the Walls app login screen:

1. Turn on **Use local Server**.
2. Enter `http://192.168.1.50:4400`. Include `http://` and leave out any slash at the end.
3. Log in with the user's normal Walls account.

The local server checks the password, then confirms the login with the remote server. **Logins therefore need internet on the Pi.** Once logged in, inspections and publishing to the local server work without internet.

Staff added on the remote server later, and changed passwords, work at once: the server stores the user when the remote server accepts the login.

## 11. Daily operation

- **Automatic restart.** The containers start again by themselves after a reboot or power cut. Test it once:
  1. Run `sudo reboot`.
  2. Wait two minutes.
  3. Repeat step 9.1.
- **Server logs:**
  - live: `docker compose logs -f app`;
  - as a file: `sudo tail -f src/logs/server.log`. The server rotates this file at 10 MB and keeps 5 old copies.
- **Disk space:**
  - free space: `df -h /`;
  - space used by photos and videos: `du -sh src/media_data`.

  Plan more storage before the disk is 80 % full.
- **Power and heat:** `vcgencmd get_throttled` must print `throttled=0x0`. Any other value means the Pi lacked power or overheated. Fix the power supply or the cooling.
- **phpMyAdmin:**
  - Open `http://192.168.1.50:8080`.
  - Log in as `roro` or `root` with the passwords from `.env`.
  - Use it with care: changes take effect immediately.
- **Network exposure:**
  - Port 4400 must be reachable by the tablets.
  - The database port 3306 is open to the Pi itself only.
  - Never forward port 4400 or 8080 from the internet to the Pi.

## 12. Backups

The data lives in two places:
- the database (Docker volume `roro-local-server_db_data`);
- the uploaded files (`src/media_data`).

Back up both.

### 12.1 Prepare a USB drive for backups

1. Format the drive as ext4, for example with `sudo mkfs.ext4 -L roro-backup /dev/sda1`.
   This erases the drive. Check the device name with `lsblk` first. When the system itself runs from an SSD, the backup drive is usually `/dev/sdb1`, not `/dev/sda1`.
2. Mount the drive at every start:

   ```bash
   sudo mkdir -p /mnt/backup
   echo 'LABEL=roro-backup /mnt/backup ext4 defaults,nofail 0 2' | sudo tee -a /etc/fstab
   sudo systemctl daemon-reload
   sudo mount -a
   sudo chown $USER: /mnt/backup
   ```

### 12.2 Backup script

1. Create `~/roro-backup.sh`:

   ```bash
   #!/bin/bash
   set -euo pipefail
   DEST=/mnt/backup
   STAMP=$(date +%Y%m%d-%H%M)

   mountpoint -q "$DEST" || { echo "backup drive not mounted at $DEST"; exit 1; }
   mkdir -p "$DEST/db"

   # Database: a consistent copy, taken while the server keeps running.
   docker exec roro-mariadb sh -c \
     'exec mariadb-dump --single-transaction --routines -uroot -p"$MYSQL_ROOT_PASSWORD" roro_local' \
     | gzip > "$DEST/db/roro_local-$STAMP.sql.gz"

   # Photos, videos and documents. Files are only ever added, so a mirror is enough.
   rsync -a "$HOME/roro-local-server/src/media_data/" "$DEST/media_data/"

   # Keep 14 days of database copies.
   find "$DEST/db" -name 'roro_local-*.sql.gz' -mtime +14 -delete
   echo "backup $STAMP done"
   ```

2. Make the script executable and run it once:

   ```bash
   chmod +x ~/roro-backup.sh
   ~/roro-backup.sh
   ```

3. Run it every night at 02:30:
   1. Run `crontab -e`.
   2. Add this line:

      ```
      30 2 * * * /home/walls/roro-backup.sh >> /home/walls/roro-backup.log 2>&1
      ```

### 12.3 Restore

1. Restore the database:

   ```bash
   gunzip -c /mnt/backup/db/roro_local-<stamp>.sql.gz \
     | docker exec -i roro-mariadb sh -c 'exec mariadb -uroot -p"$MYSQL_ROOT_PASSWORD" roro_local'
   ```

2. Restore the files:

   ```bash
   sudo rsync -a /mnt/backup/media_data/ ~/roro-local-server/src/media_data/
   ```

3. Restart the server: `docker compose restart app`.

## 13. Updating to a new version

```bash
cd ~/roro-local-server
~/roro-backup.sh
git pull
docker compose up -d --build
docker compose logs app | grep -i migration
```

- **Database changes:** the server applies new migrations by itself at start. The log shows one `Migration file applied` line per file.
- **`roro_local.sql`:** used only when the database is created for the first time.
- **Local edits:** edit only `.env` and `src/.env` on the Pi. Changes to any other file block `git pull`.

**Pis set up before October 2026.** Older images stored uploaded files inside the container, in `/media_data`, instead of in `src/media_data` on the Pi. Rebuilding the container deletes that folder. Before the first update of such a Pi:

1. Check for files:

   ```bash
   docker exec roro_app ls /media_data
   ```

2. If the command lists folders such as `images`, copy them out:

   ```bash
   sudo docker cp roro_app:/media_data/. src/media_data/
   ```

3. Then run the update commands above.

## 14. Starting again from an empty database

> **Warning:** these commands permanently delete every manifest, inspection, user session and uploaded photo on this Pi. They cannot be undone. Take a backup first (step 12) and make sure all tablets have published.

```bash
cd ~/roro-local-server
docker compose down -v
sudo find src/media_data -mindepth 1 ! -name folder_holder.txt -delete
docker compose up -d
```

The next start loads `roro_local.sql` again and pulls users as in step 8.

## Troubleshooting

| Symptom | Cause and fix |
|---|---|
| `exec format error` | The 32-bit system is installed. Reinstall with Raspberry Pi OS Lite (64-bit). |
| `roro-mariadb` is not `healthy`; `roro_app` stays `Created` | The server waits for the database. Run `docker compose logs db`. `password option is not specified` means `.env` is missing or incomplete. |
| App log repeats `Waiting for database before running migrations` with `Access denied for user 'roro'` | `MYSQL_PASS` in `src/.env` differs from `MYSQL_PASSWORD` in `.env`. Put back the password used at the first start, then run `docker compose up -d --force-recreate app`. A later change in `.env` does not change the database account. |
| Tablet login says `Invalid email or password` for a valid account | The remote server refused the login. Check the account on the remote server: it must be active, and the password must be the current one. |
| Tablet login fails with `failed to login to the remote server` | The Pi has no internet, or its clock is wrong (step 4). |
| Tablets cannot reach the server | Check that the tablet uses the same network as the Pi. Open `http://192.168.1.50:4400/health` in the tablet's browser. Check that the Pi kept its address (`hostname -I`) and that the address in the app has no slash at the end. |
| Slow or failing uploads | Run `vcgencmd get_throttled` (step 11) and `df -h /`. A full disk or under-voltage causes both. |
