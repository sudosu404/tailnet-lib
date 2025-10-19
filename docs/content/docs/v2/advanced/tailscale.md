---
title: Tailscale
next: /docs/scenarios
---


This document guides you through the different authentication and configuration
options for Tailscale with Tailnet.

## Authentication Methods

Tailnet supports three authentication methods with Tailscale: OAuth,
OAuth (manual), and AuthKey.

### OAuth

{{% steps %}}

#### Prerequisites

1. Generate an OAuth client at [https://login.tailscale.com/admin/settings/oauth](https://login.tailscale.com/admin/settings/oauth).
2. Define tags for services. Tags can be defined in the provider, applying to
all services.

> [!Important]
> All auth keys created from an OAuth client require **tags**. This is a Tailscale requirement.

#### Configuration

Add the OAuth client credentials to the Tailnet configuration:

```yaml {filename="/config/tailnet.yaml"}
tailscale:
  providers:
    default:
      clientId: "your_client_id"
      clientSecret: "your_client_secret"
      tags: "tag:example" # Optional if tags are defined in each proxy
```

#### Restart

Restart Tailnet to apply the changes.

> [!Tip]
> If the proxy fails to authenticate after restarting, check the error logs.
> Ensure the tags are correct and the OAuth client is enabled.

{{% /steps %}}

### OAuth (Manual)

{{% steps %}}

#### Disable AuthKey

OAuth authentication mode is enabled when no AuthKey is set in the Tailscale
provider configuration:

```yaml {filename="/config/tailnet.yaml"}
tailscale:
  providers:
    default:
      authKey: ""
      authKeyFile: ""
```

The proxy will wait for authentication with Tailscale during startup.

#### Dashboard

Access the Tailnet dashboard (e.g., `http://192.168.1.1:8080`).

#### Authentication

Click on the proxy with "Authentication" status.

> [!Tip]
> If "Ephemeral" is set to `true`, authentication is required at each Tailnet restart.

{{% /steps %}}

### AuthKey

{{% steps %}}

#### Generate AuthKey

1. Go to [https://login.tailscale.com/admin/settings/keys](https://login.tailscale.com/admin/settings/keys).
2. Click "Generate auth key".
3. Add a description.
4. Enable "Reusable".
5. Add tags if needed.
6. Click "Generate key".

> [!Warning]
> If tags are added to the key, all proxies initialized with the same AuthKey will receive the same tags. To use different tags, add a new Tailscale provider to the configuration.

#### Configuration

Add the AuthKey to the Tailnet configuration:

```yaml {filename="/config/tailnet.yaml"}
tailscale:
  providers:
    default:
      authKey: "YOUR_GENERATED_KEY_HERE"
      authKeyFile: ""
```

#### Restart

Restart Tailnet to apply the changes.

{{% /steps %}}

## Funnel

In addition to configuring Tailnet to enable Funnel, you need to grant
permissions in the Tailscale ACL. See [Troubleshooting](.././troubleshooting/#funnel-doesnt-work)
for more details. Also read Tailscale's [Funnel documentation](https://tailscale.com/kb/1223/funnel#requirements-and-limitations)
for requirements and limitations.

## Tags

- Tags are required for OAuth authentication.
- Tags only work with OAuth authentication.
- Tags can be configured in the provider or service.
- If tags are defined in the provider, they apply to all services.
- If tags are defined in the service, provider tags are ignored.
