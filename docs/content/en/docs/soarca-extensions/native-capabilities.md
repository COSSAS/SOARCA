---
title: Native capabilities
description: >
    Capabilities and transport mechanisms baked right into SOARCA
categories: [capabilities]
tags: [native]
weight: 2
date: 2023-01-05
---

This page contains a list of capabilities that are natively implemented in SOARCA see details [here](/docs/core-components/modules). For MQTT-message-based capabilities, check [here](/docs/soarca-extensions/).


## OpenC2 capability

The OpenC2 HTTP capability uses the http(s) transport layer as specified in [OpenC2 HTTPS](https://docs.oasis-open.org/openc2/open-impl-https/v1.0/open-impl-https-v1.0.html). It allows executing actions on an OpenC2-compatible security actuator.

CACAO documentation: [OpenC2 HTTP Command](https://docs.oasis-open.org/cacao/security-playbooks/v2.0/cs01/security-playbooks-v2.0-cs01.html#_Toc152256498)

## HTTP API capability

The HTTP capability allows sending arbitrary HTTP requests to other servers.

CACAO documentation: [HTTP API Command](https://docs.oasis-open.org/cacao/security-playbooks/v2.0/cs01/security-playbooks-v2.0-cs01.html#_Toc152256495)

## SSH capability

The SSH capability allows executing commands on systems running an SSH-server.

CACAO documentation: [SSH Command](https://docs.oasis-open.org/cacao/security-playbooks/v2.0/cs01/security-playbooks-v2.0-cs01.html#_Toc152256500)

## Powershell capability

The PowerShell capability allows executing commands on systems running an WinRM server.

CACAO documentation: [PowerShell Command](https://docs.oasis-open.org/cacao/security-playbooks/v2.0/cs01/security-playbooks-v2.0-cs01.html#_Toc152256499)

## Caldera capability

The Caldera capability allows for interoperability between SOARCA and [Caldera](https://caldera.mitre.org/).

Caldera documentation: [caldera Command](https://docs.oasis-open.org/cacao/security-playbooks/v2.0/cs01/security-playbooks-v2.0-cs01.html#_Toc152256493)
