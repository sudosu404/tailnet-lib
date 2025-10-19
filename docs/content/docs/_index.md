---
linkTitle: "Documentation"
title: Introduction
weight: 1
---

👋 Welcome to the Tailnet documentation!

## What is Tailnet?

Tailnet is an application that automatically creates a proxy to
virtual addresses in your Tailscale network.
Easy to configure and deploy, based on Docker container labels or a simple proxy
list file.
It simplifies traffic redirection to services running inside Docker containers,
without the need for a separate Tailscale container for each service.

> [!NOTE]
> Tailnet just needs a label in your new docker service or a proxy list file and
> it will be automatically created in your Tailscale network and the proxy will be
> ready to be used.

## Why another proxy?

Tailnet was created to address the need for a proxy that can handle multiple services
without the need for a dedicated Tailscale container for each service and without configuring
virtual hosts in Tailscale network.

![how Tailnet works](/images/tailnet.svg)

## What's different with Tailnet?

Tailnet differs from other Tailscale proxies in that it does not require a separate Tailscale container per service.

![how Tailnet works](/images/Tailnet-compare.svg)

## Features

- **Easy to Use** - Creates virtual Tailscale addresses using Docker container labels.
- **Really Easy to Use** - Creates virtual Tailscale addresses using a proxy list.
- **Lightweight** - No need to spin up a dedicated Tailscale container for every service.
- **Quick deploy** - No need to configure virtual hosts in Tailscale network.
- **Automatically supports TLS** - Automatically supports Tailscale/LetsEncrypt certificates
with MagicDNS.

## Questions or Feedback?

> [!IMPORTANT]
  Tailnet is still in active development.
  Have a question or feedback? Feel free to [open an issue](https://github.com/sudosu404/tailnet-lib/issues)!

## Next

Dive right into the following section to get started:

{{< cards >}}
  {{< card link="getting-started" title="Getting Started" icon="document-text"
    subtitle="Learn how to get started with Tailnet"
  >}}
{{< /cards >}}
