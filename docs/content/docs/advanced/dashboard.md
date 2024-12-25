---
title: Dashboard
prev: /docs/advanced
---

{{% steps %}}

### Dashboard in docker

#### TSDProxy docker compose

Update docker-compose.yml with the following:

```yaml docker-compose.yml
    labels:
      - tsdproxy.enable=true
      - tsdproxy.name=dash
```

#### Restart TSDProxy

```bash
docker compose restart
```

### Standalone

#### Configure files provider

Configure a new files provider or configure it in any existing files provider.

```yaml  {filename="/config/tsdproxy.yaml"}
files:
  proxies:
    filename: /config/proxies.yaml
```

#### Add Dashboard entry on your files provider

```yaml {filename="/config/proxies.yaml"}
dash:
  url: http://127.0.0.1:8080
```

### Test Dashboard access

```bash
curl https://dash.FUNNY-NAME.ts.net
```

{{< callout type="info" >}}
Note that you need to replace `FUNNY-NAME` with the name of your network.
{{< /callout >}}

{{% /steps %}}
