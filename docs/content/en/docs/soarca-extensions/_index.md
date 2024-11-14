---
title: Extensions & Capabilities
description: >
    Extending SOARCA is done by developing a SOARCA-Fin.  
categories: [extensions, architecture, capabilities]
tags: [fin]
weight: 5
date: 2023-01-05
---


{{% alert title="Warning" color="warning" %}}
SOARCA V.1.0.X implements currently the following native capalities: **HTTP capability**, **OpenC2 capability**, **SSH capability**, and **Caldera capability**. Other core capabilities are part of our milestones which can be found [here](https://github.com/COSSAS/SOARCA/milestones).
{{% /alert %}}

SOARCA features a set of [native capabilities](/docs/soarca-extensions/native-capabilities). The HTTP, OpenC2 HTTP, and SSH transport mechanisms are supported by the first release of SOARCA. SOARCA's capabilities can be extended with custom implementations, which is further discussed on this page.

## Extending the native capabilities

The native capabilities supported by SOARCA can be extended through a mechanism we named Fins. Your capability can be integrated with SOARCA by implementing the Fin protocol. This protocol regulates communication between SOARCA and the extension capabilities over an MQTT bus.

MQTT is a lightweight messaging protocol with libraries written in various programming languages. To integrate with SOARCA, you can write your own implementation of the Fin protocol, or use our [python](https://www.python.org/) or [golang](https://go.dev/) libraries for easier integration.

## Fin protocol

The underlying protocol for the SOARCA fins can be found [here](/docs/soarca-extensions/fin-protocol).

