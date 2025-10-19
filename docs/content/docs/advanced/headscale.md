---
title: Headscale
draft: true
---

In case you want to use the Headscale service, please read the following:

{{% steps %}}

### In your Tailnet docker-compose.yaml

Add the following to the `environment` section:

```yaml docker-compose.yml
   environment:
      - TAILNET_CONTROLURL="url of you headscale server"
```

### Restart Tailnet

```bash
docker compose restart
```

{{% /steps %}}
