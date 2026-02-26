# Installers

Each Sonar application has a dedicated installer that wraps the upstream project's install process.

## Customer Portal

- **Repo:** https://github.com/SonarSoftwareInc/customer_portal
- **Upstream method:** `git clone` → `sudo ./install.sh`
- **Type:** Docker-based (Laravel app)
- **OS:** Ubuntu 18 or 22
- **Prerequisites:** Public IP, valid domain name pointing to server
- **Config prompts:** Sonar URL, API username, API password, domain, email for SSL

### Steps
1. Install git, unzip
2. Clone repository
3. Run `install.sh` (installs Docker, builds containers, prompts for config)
4. Access at `https://<domain>/settings` with generated settings key

---

## Netflow On-Prem

- **Repo:** https://github.com/SonarSoftwareInc/netflow-onprem
- **Upstream method:** `git clone` → configure `.env` → `sudo ./install.sh`
- **Type:** Docker-based
- **OS:** Ubuntu 24.04 LTS (recommended), also Debian 10-12
- **Prerequisites:** Dedicated host, accurate NTP
- **Config prompts:** Sonar URL, API token, netflow name, public IP, retention settings, DB password

### Steps
1. Install git, make, unzip
2. Clone repository
3. Copy `.env.example` → `.env` and populate values
4. Run `install.sh` (installs Docker, builds images — 15-30 min)
5. Reboot

---

## FreeRADIUS Genie v3

- **Repo:** https://github.com/SonarSoftwareInc/freeradius_genie-v3
- **Upstream method:** `git clone` → run `./genie`
- **Type:** Native install (PHP CLI tool)
- **OS:** Ubuntu 24.04 64-bit (recommended)
- **Prerequisites:** None beyond base Ubuntu
- **Config prompts:** Sonar instance details (see Sonar KB)

### Steps
1. Clone repository
2. Copy `.env.example` → `.env` and configure
3. Run `./genie` to install and configure freeRADIUS

---

## Poller

- **Repo:** https://github.com/SonarSoftwareInc/poller
- **Upstream method:** `wget setup.sh` → `sudo ./setup.sh`
- **Type:** Native install (PHP/Composer, supervisord)
- **OS:** Ubuntu 24 Server
- **Prerequisites:** Bare metal or VM
- **Config prompts:** Sonar URL, poller API key (from Sonar Settings > Monitoring > Pollers)

### Steps
1. Download `setup.sh` from raw GitHub
2. Run `setup.sh` (installs PHP, composer deps, nginx, supervisord)
3. Access web UI at `https://<server_ip>`
4. Login with Sonar credentials, configure Settings tab
