---
title: Vision & Concepts
weight: 3
description: >
  The what and why of SOARCA
resources:
- src: "*Slide2.jpg"
  params:
    byline: "*Slide*: © 2024 TNO"
---

## Context and Background

**S**ecurity **O**rchestrator for **A**dvanced **R**esponse to **C**yber **A**ttacks​ - SOARCA


Organisations are increasingly automating threat and incident response through playbook driven security workflow orchestration. The essence of this concept is that specific security events trigger a predefined series of response actions that are executed with no or only limited human intervention. These automated workflows are captured in machine-readable security playbooks, which are typically executed by a so called Security Orchestration, Automation and Response (SOAR) tool. The market for SOAR solutions has matured significantly over the past years and present day products support sophisticated automation workflows and a wide array of integrations with external security tools and data resources. Typically, however, the technology employed is proprietary and not easily adaptable for research and experimentation purposes. SOARCA aims to offer an open-source alternative for such solutions that is free of vendor dependencies and supports standardized formats and technologies where applicable. 

SOARCA, TNO’s open-source SOAR, was developed for research and innovation purposes and allows SOC, CERT and CTI professionals to experiment with the concept of playbook driven security automation. It is open and extensible and its interfaces are well-defined and elaborately documented. It also offers native support for two emerging technology standards, both developed and maintained by OASIS Open:

- [CACAOv2](https://docs.oasis-open.org/cacao/security-playbooks/v2.0/security-playbooks-v2.0.html). The Collaborative Automated Course of Action Operations (CACAO) standard provides a common framework and machine-processable schema for security playbooks that are natively interoperable and can be shared and executed across technological and organizational boundaries.
- [OpenC2](https://openc2.org/). A standardized language for the command and control of cyber defense technologies. In essence it provides a vendor agnostic language and interface through which so called security actuators (e.g. firewalls or IAM solutions) can be reconfigured automatically.

SOARCA is available through [TNO’s](https://www.tno.nl/nl/) community platform [COSSAS](https://cossas-project.org/) (Community for Open Source Security Automation Software) under the [Apache 2.0 license](https://www.apache.org/licenses/LICENSE-2.0). With its release, TNO aims to drive both the adoption and further development of novel technologies for cyber security automation forward. Here we note that open and accessible SOAR functionality is not only relevant for automation in threat and incident response but also for attack & defense simulations, cyber ranges, digital twinning and other emerging innovations that require orchestration of complex (security oriented) workflows.

---
{{% imgproc Slide2 Fill "1280x720" %}}

{{% /imgproc %}}

### Current state of SOARCA

At present, SOARCA is in an Alpha release phase and is intended for Proof of Concepts (PoCs) and research purposes, serving as a platform for demonstrations. The objective of the SOARCA team is to evolve SOARCA into a more mature SOAR orchestration tool suitable for operational environments. For potential applications of SOARCA, please refer to the ‘Use-Cases’ section of our documentation.

## Core Concepts
Several concepts within SOARCA might be important to know.

### Course of Action

A course of action (CoA) refers to a planned sequence of steps or activities taken to achieve a specific cyber security goal. These steps are often collected into "playbooks". Usually in the form of prose in PDFs, internal wikis, or even scattered throughout emails.

### CACAO Playbooks: Streamlining Cybersecurity Operations

The [CACAO Security Playbooks Version 2.0 specification](https://docs.oasis-open.org/cacao/security-playbooks/v2.0/security-playbooks-v2.0.html) provides a standard for writing _executable_ playbooks. These playbooks are stored in a machine-readable form, allowing them to be (semi-)automatically executed by an orchestration tool.

A CACAO playbook is a structured document that outlines a series of orchestrated actions to address specific security events, incidents, or other security-related activities. These playbooks allow for the automation of security steps.

Examples of repetitive tasks that might be automated using a CACAO Playbook might be:

- Investigate the cause of security events.
- Mitigate threats effectively.
- Remediate vulnerabilities.

By following CACAO playbook specifications, organizations can enhance their automated response capabilities, foster collaboration, and ensure consistency of playbooks across diverse technological solutions.

More information can be found in our [primer on playbooks](/docs/concepts/executable-playbooks).

### SOARCA Fin(s): Extending the core capabilities

SOARCA can be extended with custom extensions or rather so-called FIN (inspired by the majestic orca). A fin can be integrated within the SOARCA core. Technical descriptions of the components can be found [here](/docs/soarca-extensions/fin-protocol). Fins communicate with the SOARCA core using a pre-defined MQTT protocol. 


## Join the SOARCA Community

The SOARCA team invites cybersecurity professionals, researchers, and enthusiasts to join our community. Explore, adapt, and contribute to SOARCA’s growth. Let’s fortify cyber defenses together! See our [contribution guidelines](/docs/contribution-guidelines/) on how to make contributions.	🛡️🌐

## Key Details
- Project Name: SOARCA (Security Orchestrator for Advanced Response to Cyber Attacks)
- License: Apache 2.0
- Release Date: March 2024
