---
title: SOARCA Extensions
description: >
    Extending SOARCA is done by developing a SOARCA-Fin.  
categories: [extensions, architecture]
tags: [fin]
weight: 5
date: 2023-01-05
---

The native capabilities support by the SOARCA core can be extended through a mechanism known as FINS. Whether you’re working directly with our system or utilizing a third-party library, the key lies in implementing the Fin protocol. This protocol provides for communication between the SOARCA core and the extension capabilities via an MQTT-bus.

The choice of MQTT as the communication protocol allows for seamless integration with libraries written in various programming languages while maintaining a relatively straightforward approach.

## SOARCA Python Fin library

As part of the SOARCA suite there is currently an library that implement the Fin protocol and which can be used within your project. 

## Loading your module
Once you have developed your module you need to load it so SOARCA can use it for the playbooks it executes. You can load your modules in two ways via docker or stand alone.

### Docker
The Docker engine allows for easy loading but requires you to package your capability into a docker container. Once that is done you can add your container to a docker compose file and it will register itself to SOARCA once started. The Fin MUST be loaded after SOARCA's main components otherwise the Fin might not work correctly. 

### Stand alone
SOARCA can be used without Docker. To use it whit your module you need to start it and have an MQTT broker running already before starting your Fin. *The method is for more complex setups and not recommended for first use.*

First set up SOARCA [stand alone](setup.md).


## Protocol

The underlying protocol for the SOARCA fins can be [here](/docs/soarca-extentions/fin-protocol).

