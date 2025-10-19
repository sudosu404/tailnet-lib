---
title: Docker Services
weight: 3
---

To add a service to your Tailnet instance, you need to add a label to your
service container.

{{% steps %}}

### tailnet.enable

Just add the label `tailnet.enable` to true and restart you service. The
container will be started and Tailnet will be enabled.

```yaml
labels:
  tailnet.enable: "true"
```

TSProxy will use container name as Tailscale server, and will use one exposed
port to proxy traffic.

### tailnet.name

If you define a name different from the container name, you can define it with
the label `tailnet.name` and it will be used as the Tailscale server name.

```yaml
labels:
  tailnet.enable: "true"
  tailnet.name: "myserver"
```

### tailnet.container_port

If you want to define a different port than the default one, you can define it
with the label `tailnet.container_port`.
This is useful if the container has multiple exposed ports or if the container
is running in network_mode=host.

```yaml
ports:
  - 8081:8080
  - 8000:8000
labels:
  tailnet.enable: "true"
  tailnet.name: "myserver"
  tailnet.container_port: 8080
```

> [!NOTE]
Note that the port used in the `tailnet.container_port` label is the port used
internal in the container and not the exposed port.

### tailnet.ephemeral

If you want to use an ephemeral container, you can define it with the label `tailnet.ephemeral`.

```yaml
labels:
  tailnet.enable: "true"
  tailnet.name: "myserver"
  tailnet.ephemeral: "true"
```

### tailnet.webclient

If you want to enable the Tailscale webclient (port 5252), you can define it
with the label `tailnet.webclient`.

```yaml
labels:
  tailnet.enable: "true"
  tailnet.name: "myserver"
  tailnet.webclient: "true"
```

### tailnet.tsnet_verbose

If you want to enable Tailscale's verbose mode, you can define it with the label
`tailnet.tsnet_verbose`.

```yaml
labels:
  tailnet.enable: "true"
  tailnet.name: "myserver"
  tailnet.tsnet_verbose: "true"
```

### tailnet.funnel

To enable funnel mode, you can define it with the label `tailnet.funnel`.

```yaml
labels:
  tailnet.enable: "true"
  tailnet.name: "myserver"
  tailnet.funnel: "true"
```

### tailnet.authkey

Enable Tailnet authentication with a different Authkey.
This give the possibility to add tags on your containers if were defined when
created the Authkey.

```yaml
labels:
  tailnet.enable: "true"
  tailnet.authkey: "YOUR_AUTHKEY_HERE"
```

### tailnet.authkeyfile

Authkeyfile is the path to your Authkey. This is useful if you want to use
docker secrets.

```yaml
labels:
  tailnet.enable: "true"
  tailnet.authkey: "/run/secrets/authkey"
```

### tailnet.proxyprovider

If you want to use a proxy provider other than the default one, you can define
it with the label `tailnet.proxyprovider`.

```yaml
labels:
  tailnet.enable: "true"
  tailnet.proxyprovider: "providername"
```

### tailnet.autodetect

Defaults to true, if your having problem with the internal network interfaces
autodetection, set to false.

```yaml
labels:
  tailnet.enable: "true"
  tailnet.autodetect: "false"
```

### tailnet.scheme

Defaults to "http", set to https to enable "https" if the container is running
with TLS.

```yaml
labels:
  tailnet.enable: "true"
  tailnet.scheme: "https"
```

### tailnet.tlsvalidate

Defaults to true, set to false to disable TLS validation.

```yaml
labels:
  tailnet.enable: "true"
  tailnet.scheme: "https"
  tailnet.tlsvalidate: "false"
```

### tailnet.dash.visible

Defaults to true, set to false to hide on Dashboard.

```yaml
labels:
  tailnet.enable: "true"
  tailnet.dash.visible: "false"
```

### tailnet.dash.label

Sets the proxy label on dashboard. Defaults to tailnet.name.

```yaml
labels:
  tailnet.enable: "true"
  tailnet.name: "nas"
  tailnet.dash.label: "Files"
```

### tailnet.dash.icon

Sets the proxy icon on dashboard. If not defined, Tailnet will try to find a
icon based on the image name. See available icons in [icons](/docs/advanced/icons).

```yaml
labels:
  tailnet.enable: "true"
  tailnet.dash.icon: "si/portainer"
```

{{% /steps %}}
