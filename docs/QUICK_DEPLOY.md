# Quick Deploy Reference

## One-Command Deploy

```bash
./scripts/deploy-catalyst.sh
```

---

## Manual Deploy (Step-by-Step)

### On Local Machine

```bash
# 1. Verify package exists
ls -lh build/catalyst-deployment.tar.gz

# 2. Upload to server
scp build/catalyst-deployment.tar.gz user@ads.thenexusengine.com:/tmp/

# 3. Deploy test page
scp assets/test-magnite.html user@ads.thenexusengine.com:/tmp/
```

### On Remote Server

```bash
# SSH to server
ssh user@ads.thenexusengine.com

# Deploy application
cd /opt/catalyst
sudo systemctl stop catalyst
sudo tar xzf /tmp/catalyst-deployment.tar.gz --strip-components=1
sudo mv build/catalyst-server ./catalyst-server
sudo chmod +x catalyst-server
sudo systemctl start catalyst
sleep 3

# Deploy test page
sudo cp /tmp/test-magnite.html /var/www/html/

# Verify
curl -s http://localhost:8000/health | jq .
sudo journalctl -u catalyst -n 50
```

---

## Quick Tests

### Health Check
```bash
curl https://ads.thenexusengine.com/health
```

### Bid Request Test
```bash
curl -X POST https://ads.thenexusengine.com/v1/bid \
  -H "Content-Type: application/json" \
  -d '{
    "accountId": "icisic-media",
    "timeout": 2800,
    "slots": [{
      "divId": "test",
      "sizes": [[728,90]],
      "adUnitPath": "totalprosports.com/leaderboard"
    }]
  }' | jq .
```

### Browser Test
```
https://ads.thenexusengine.com/test-magnite.html
```

---

## Monitor Logs

```bash
# Live logs
ssh user@ads.thenexusengine.com 'sudo journalctl -u catalyst -f'

# Last 100 lines
ssh user@ads.thenexusengine.com 'sudo journalctl -u catalyst -n 100'

# Search for errors
ssh user@ads.thenexusengine.com 'sudo journalctl -u catalyst -n 1000 | grep -i error'
```

---

## Rollback (if needed)

```bash
ssh user@ads.thenexusengine.com
cd /opt/catalyst
sudo systemctl stop catalyst
sudo cp catalyst-server.backup.YYYYMMDD-HHMMSS catalyst-server
sudo systemctl start catalyst
```

---

## Key Log Messages

**Startup:**
```
✓ Loaded bidder mapping: 10 ad units for publisher icisic-media
✓ Configured bidders: rubicon, kargo, sovrn, oms, aniview, pubmatic, triplelift
✓ Catalyst MAI Publisher endpoint registered: /v1/bid
```

**Bid Request:**
```
✓ Found mapping for ad unit: totalprosports.com/leaderboard
✓ Injected parameters for 7 bidders
✓ Catalyst bid request completed: 2 bids in 312ms
```

**Warnings (OK):**
```
⚠ No mapping found for ad unit: unknown.com/test
```

---

## Quick Troubleshooting

| Issue | Command |
|-------|---------|
| Service down | `sudo systemctl start catalyst` |
| Check status | `sudo systemctl status catalyst` |
| View config | `cat config/bizbudding-all-bidders-mapping.json \| jq .` |
| Test locally | `curl http://localhost:8000/health` |
| Check ports | `sudo netstat -tlnp \| grep 8000` |
| Disk space | `df -h` |
| Memory usage | `free -h` |

---

## File Locations on Server

```
/opt/catalyst/
├── catalyst-server              # Main binary
├── config/
│   └── bizbudding-all-bidders-mapping.json
├── assets/
│   ├── catalyst-sdk.js
│   └── tne-ads.js
└── catalyst-server.backup.*     # Backups

/var/www/html/
└── test-magnite.html           # Test page

/etc/systemd/system/
└── catalyst.service            # Service config
```

---

## Environment Variables (on server)

Edit: `/opt/catalyst/.env`

```bash
PORT=8000
HOST_URL=https://ads.thenexusengine.com
LOG_LEVEL=info
PBS_TIMEOUT=2500ms
```

After changes:
```bash
sudo systemctl restart catalyst
```

---

## Metrics & Monitoring

```bash
# All metrics
curl https://ads.thenexusengine.com/metrics

# Catalyst metrics only
curl https://ads.thenexusengine.com/metrics | grep catalyst

# Bidder metrics
curl https://ads.thenexusengine.com/metrics | grep bidder_requests
```

---

## Success Checklist

- [ ] Health endpoint returns 200
- [ ] SDK loads without 404
- [ ] Bid endpoint accepts POST requests
- [ ] Logs show "Loaded bidder mapping"
- [ ] Logs show "Injected parameters for 7 bidders"
- [ ] Browser test page loads
- [ ] No errors in journalctl logs
- [ ] Metrics endpoint responding

---

**Deployment Time:** ~15 minutes
**Testing Time:** ~20 minutes

**Ready?** Run: `./scripts/deploy-catalyst.sh`
