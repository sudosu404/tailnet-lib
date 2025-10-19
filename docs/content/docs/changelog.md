---
title: Changelog
weight: 500
---

{{% steps %}}

### 2.0.0-beta4

#### New features

- Multiple ports in each tailscale hosts
- Enable and activate multiple redirects
- Proxies can use http and https
- OAuth autentication without using the dashboard
- Assign Tags on Tailscale hosts
- Dashboard gets updated in real-time
- Search in the dashboard
- Dashboard proxies are sorted in alphabetically order
- Add support to swam stacks
- Tailscale user profile in top-right of Dashboard
- Pass Tailscale identity headers to destination service

#### Breaking changes

- Files provider is now Lists ( key in /config/tailnet.yaml changed to
**lists:** instead of files:)
- Lists are now a different yaml file to support multiple ports and redirects,
please [Lists](../v2/providers/lists)

#### Deprecated Docker labels

- tailnet.autodetect
- tailnet.container_port
- tailnet.funnel
- tailnet.scheme
- tailnet.tlsvalidate

### 1.4.0

#### New features

- OAuth authentication using the Dashboard.
- Dashboard has now proxy status.
- Icons and Labels can be used to customize the Dashboard.

#### Fixes

- Error on port when autodetect is disabled.

### 1.3.0

#### Braking changes

Configuration files are now validated and doesn't allow invalid configuration keys
[Verify valid configuration keys](../serverconfig/#sample-configuration-file).

#### New features

- Generate TLS certificates for containers when starting proxies.
- Configuration files are now validated.

### 1.2.0

#### New features

Dashboard finally arrived.

### 1.1.2

#### Fixes

Reload Proxy List Files when changes.

#### New features

- Quicker start with different approach to start proxies in docker
- Add support for targets with self-signed certificates.

### 1.1.1

#### New Docker container labels

##### tailnet.autodetect

If Tailnet, for any reason, can't detect the container's network you can
disable it.

##### tailnet.scheme

If a container uses https, use tailnet.scheme=https label.

### 1.1.0

#### New File Provider

Tailnet now supports a new file provider. It's useful if you want to proxy URL
without Docker.
Now you can use Tailnet even without Docker.

### 1.0.0

#### New Autodetection function for containers network

Tailnet now tries to connect to the container using docker internal
ip addresses and ports. It's more reliable and faster, even in container without
exposed ports.

#### New configuration method

Tailnet still supports the Environment variable method. But there's much more
power with the new configuration yaml file.

#### Multiple Tailscale servers

Tailnet now supports multiple Tailscale servers. This option is useful if you
have multiple Tailscale accounts, if you want to group containers with the same
AUTHKEY or if you want to use different servers for different containers.

#### Multiple Docker servers

Tailnet now supports multiple Docker servers. This option is useful if you have
multiple Docker instances and don't want to deploy and manage Tailnet on each one.

#### New installation scenarios documentation

Now there is a new  [scenarios](/docs/scenarios) section.

#### New logs

Now logs are more readable and easier to read and with context.

#### New Docker container labels

**tailnet.proxyprovider** is the label that defines the Tailscale proxy
provider. It's optional.

#### Tailnet can now run standalone

With the new configuration file, Tailnet can be run standalone.
Just run tailnetd --config ./config .

#### New flag --config

This new flag allows you to specify a configuration file. It's useful if you
want to use as a command line tool instead of a container.

```bash
tailnetd --config ./config/tailnet.yaml
```

{{% /steps %}}
