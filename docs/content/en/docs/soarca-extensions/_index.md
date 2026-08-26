---
title: Extensions & Capabilities
description: >
    Extending SOARCA is done by developing a SOARCA-Fin.  
categories: [extensions, architecture, capabilities]
tags: [fin]
weight: 6
date: 2023-01-05
---


{{% alert title="Warning" color="warning" %}}
SOARCA V.1.0.X implements currently the following native capabilities: **HTTP capability**, **OpenC2 capability**, **SSH capability**, **Manual capability** and **PowerShell (WINRM)**. Other core capabilities are part of our milestones which can be found [here](https://github.com/COSSAS/SOARCA/milestones).
{{% /alert %}}

SOARCA features a set of [native capabilities](/docs/soarca-extensions/native-capabilities). The HTTP, OpenC2 HTTP, and SSH transport mechanisms are supported by the first release of SOARCA. SOARCA's capabilities can be extended with custom implementations, which is further discussed on this page.

## Extending the native capabilities

The native capabilities supported by SOARCA can be extended through a mechanism we named Fins. Your capability can be integrated with SOARCA by implementing the Fin protocol.

{{% alert title="Note" color="info" %}}
The Fin protocol is being redesigned as a simple, pull-based HTTP/JSON
protocol, replacing the previous MQTT-based implementation. Documentation
for the new protocol will be published here once it lands.
{{% /alert %}}

