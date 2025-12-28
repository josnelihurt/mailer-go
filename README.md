# Mailer-Go

SMS-to-Email gateway for GSM modems running on LXC containers.

## Purpose

Receive text messages from anywhere in the world without paying roaming charges. The system monitors GSM modems connected to LXC containers and forwards received SMS to email.

## Architecture

```
GSM Modem → SMSTools3 → Filesystem Queue → Mailer-Go → Email
```

### Process Flow

1. GSM modem receives SMS message
2. SMSTools3 daemon reads message from modem via serial port
3. Message is written to filesystem queue at `/var/spool/sms/incoming/`
4. Mailer-Go watches the incoming folder
5. New messages are parsed and formatted
6. Email is sent via SMTP (Gmail)
7. Processed messages moved to `/var/spool/sms/done/`
8. Failed messages moved to `/var/spool/sms/err/`

## Deployment

Runs on Proxmox LXC containers:
- Container 751 (gsm-modem-1): 192.168.31.51
- Container 752 (gsm-modem-2): 192.168.31.52

Each container handles one modem independently with identical configuration.

## Hardware

- Huawei GSM Modem (12d1:1c10) - IMEI: XXXXXXXXXXXXXXX
- Huawei GSM Modem (12d1:1c05) - IMEI: XXXXXXXXXXXXXXY

## Dependencies

### System Level
- SMSTools3 (v3.1.21) - SMS gateway daemon
- Docker (docker.io)
- udev rules for persistent device naming

### Application
- Go application (containerized)
- Docker image: `josnelihurt/mailer-go:latest`

### Configuration Files
- `/etc/smsd.conf` - SMSTools3 configuration
- `/opt/mailer-go/config.yaml` - Email credentials and folder paths

## Deployment Scripts

Infrastructure provisioning:
```bash
./scripts/prox2/deploy_gsm_modems.sh
```

Application deployment:
```bash
./scripts/prox2/deploy_mailer_go.sh
```

## Network

Both containers use host networking (network_mode: host) on the 192.168.31.0/24 subnet.
