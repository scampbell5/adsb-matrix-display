# Public Repository Privacy Notes

This repository is structured so deployment-specific identity and location data can stay outside Git.

## Not committed

Do not commit real values for:

- receiver latitude/longitude
- street address or ZIP code
- controller LAN IP/address
- API tokens or webhook URLs
- feeder UUIDs or account identifiers
- email addresses or phone numbers
- private station branding if you want to keep it separate from your Git identity

Use Portainer/Docker environment variables for those values.

## Runtime branding

The public default is:

```text
DISPLAY_NAME=ADS-B Matrix
```

A private deployment can set a different name without modifying source files.

## Before publishing changes

Recommended checks:

```bash
git status
git diff --cached
```

Make sure `.env` or other local configuration files are not staged. If privacy separation matters, also check your Git commit author name/email before the first public push.
