# Philips Hue Integration

When a message is sent to the hexboard display (via web form or TCP), it can automatically turn on a Philips Hue light. This is useful for wall-mounted displays — the room light comes on so the message is actually visible.

The integration is optional and disabled by default. The service starts and runs normally with no Hue config present.

## Setup

### 1. Find your bridge IP

```bash
curl https://discovery.meethue.com/
# Returns: [{"id":"...","internalipaddress":"192.168.x.x"}]
```

Or check your router's device list for a device named `Philips-hue`.

### 2. Get an API key

Press the physical link button on top of your Hue bridge, then within 30 seconds run:

```bash
curl -X POST http://YOUR_BRIDGE_IP/api \
  -H "Content-Type: application/json" \
  -d '{"devicetype":"hexboard#pi"}'
```

Response:
```json
[{"success":{"username":"ABC123..."}}]
```

The `username` value is your API key.

### 3. Find the light ID

```bash
curl http://YOUR_BRIDGE_IP/api/YOUR_API_KEY/lights
```

This returns all lights with their names. Use the number key (e.g. `"3"`) for the light you want to control.

### 4. Create the config on the Pi

```bash
ssh txt.local "sudo tee /var/lib/hexboard/hue.toml" << 'EOF'
bridge_ip = "192.168.x.x"
api_key   = "your-api-key"
device_id = "3"
EOF
```

### 5. Restart the service

```bash
ssh txt.local sudo systemctl restart hexboard
ssh txt.local sudo journalctl -u hexboard -n 5
# Expected: hue: enabled (bridge=192.168.x.x device=3)
```

### 6. Test it

```bash
echo "hello" | nc txt.local 8080
# The configured light should turn on within a few seconds
```

## Behaviour

- The light turns on every time a message is displayed (web form or TCP)
- The Hue call is fire-and-forget — if the bridge is unreachable, the message still displays immediately (the error appears in logs ~5 seconds later)
- The service starts cleanly if the config file is absent — Hue is silently disabled
- If the config file exists but is missing fields, the service logs an error and starts with Hue disabled

## Disabling

Remove the config file and restart:

```bash
ssh txt.local sudo rm /var/lib/hexboard/hue.toml
ssh txt.local sudo systemctl restart hexboard
```
