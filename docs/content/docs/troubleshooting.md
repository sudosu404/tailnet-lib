---
title: Troubleshooting
prev: /docs/advanced
weight: 300
---

## How to troubleshoot Tailnet

### Global

1. Verify your tailnet.yaml file. Configuration files are Case sensitive.
[Verify your files](../serverconfig/#sample-configuration-file)

### Docker provider

1. Verify if you added the label with tailnet.enable=true
2. Force use of the port adding tailnet.container_port=xxx to the container
3. If your container is using https add tailnet.scheme="https" to your container
4. If case of self certificates also add tailnet.tlsvalidate=false
5. Check if your firewall isn't blocking the traffic
6. Add your container to the same Tailnet docker network
7. Disable autodetection with tailnet.autodetect="false" in your container
8. Verify if your case isn't in the next [Common errors](#common-errors)
9. Still having problems? Send a [Bug report](https://github.com/sudosu404/tailnet-lib/issues/new/choose)

### Proxies List provider

1. Configuration files are case sensitive. [Verify your files](../list/#proxy-list-file-options)

## Common Errors

{{% steps %}}

### http: proxy error: tls: failed to verify certificate: x509: certificate

The actual error is a TLS error. The most common cause is that the target has a
self-signed certificate.

```yaml
tailnet.enable: true
tailnet.scheme: https
tailnet.tlsvalidate: false
```

### http: proxy error: dial tcp 172.18.0.1:8001: i/o timeout

This error is caused by the target not being reachable. It's a network error.

#### Cause: Firewall

Most likely the firewall is blocking the traffic. If using UFW, execute this command:

```bash
sudo ufw allow in from 172.17.0.0/16
```

#### Cause: Failed docker autodetection

Try to disable autodetection and define the port:

```yaml
labels:
  tailnet.enable: "true"
  tailnet.autodetect: "false"
  tailnet.container_port: 8001
```

### Funnel doesn't work

#### Cause: Funnel not enabled

Visit <https://tailscale.com/kb/1223/funnel#funnel-node-attribute> to enable Funnel in ACL

#### Cause: Using tags with Funnel

If using tags, edit the attribute to include your tag(s), e.g.:

```json
"nodeAttrs": [
 {
  "target": ["autogroup:member", "tag:server"],
  "attr":   ["funnel"],
 },
],
```

{{%/ steps %}}
