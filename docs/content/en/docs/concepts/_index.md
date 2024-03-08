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

## Background of SOARCA

**S**ecurity **O**rchestrator for **A**dvanced **R**esponse to **C**yber **A**ttacks​ - SOARCA

SOARCA is [TNO’s](https://www.tno.nl/nl/) new open-source SOAR (Security Orchestration, Automation and Response) tool, which is developed for research and demonstration purposes. With SOARCA, TNO’s goal is to realise and stimulate advanced cyber security innovations and empower end users and organizations by providing a vendor-agnostic, extensible, and standards-compliant solution for security orchestration. SOARCA is made available on [COSSAS](https://cossas-project.org/) (Community for Open Source Security Automation Software) under the [Apache 2.0 licence](https://www.apache.org/licenses/LICENSE-2.0).​

While there are already several mature SOAR tools available on the market, many of them are commercial closed-source products, and do not comply with the new emerging OASIS Open standards. SOARCA is designed to fully comply with the newest standards [CACAO v2.0](https://docs.oasis-open.org/cacao/security-playbooks/v2.0/security-playbooks-v2.0.html) and [OpenC2](https://openc2.org/).

TNO’s SOARCA bridges this gap to let end users and organisations get hands-on experience with SOAR tooling and enable innovations: it is vendor-agnostic, extensible and has open and well-defined interfaces. SOARCA will be freely available and geared toward research and demonstrations. The goal is to foster a healthy community around SOARCA. ​

Note that open and accessible SOAR functionality is relevant not only for automation in cyber incident response handling but also for attack & defense simulations, cyber ranges, digital twinning and other growing innovation topics that have a strong dependence on the orchestration of complex workflows.


{{% imgproc Slide2 Fill "1280x720" %}}

{{% /imgproc %}}

## Vision of SOARCA

### Why SOARCA?

Both inside and outside of TNO there is a strong need for interoperable workflow orchestration tooling that aids (cybersecurity) innovation. High-quality SOAR (Security Orchestration, Automation and Response) tools are widely available in the market, however, these are commercial products with significant license costs and that employ proprietary technologies rather than emerging innovative standards.


- **Vendor-Agnostic Compatibility**: Our solution ensures seamless integration with various vendors, eliminating reliance on a single provider.
- **Standard Compliance**: Adhering to the latest standards, including [CACAO v2.0](https://docs.oasis-open.org/cacao/security-playbooks/v2.0/security-playbooks-v2.0.html) and [OpenC2](https://openc2.org/), guarantees up-to-date and secure operations.
- **Extensibility with Open Interfaces**: Enjoy the flexibility of an extensible tool featuring open and well-defined interfaces, promoting adaptability and customization.
- **Open-Source**: Embrace an open-source model that not only offers cost-effective solutions but also supports unrestricted use and adaptation for research purposes.


SOAR functionality is relevant not only for automation in incident response handling but also attack & defense simulations, cyber ranges, digital twinning and other (TNO research) topics that have a strong dependence on the orchestration of complex workflows.

### Current state of SOARCA

At present, SOARCA is in an Alpha release phase and is intended for Proof of Concepts (PoCs) and research purposes, serving as a platform for demonstrations. The objective of the SOARCA team is to evolve SOARCA into a more mature SOAR orchestration tool suitable for operational environments. For potential applications of SOARCA, please refer to the ‘Use-Cases’ section of our documentation.

### Why make SOARCA open-source?

- SOARCA has been publicly funded and should therefore ideally be made publicly available.
- The target audience of SOC, CERT/CSIRT and CTI teams has a very strong affinity with open-source solutions and embraces them to a great extent. (see also the success of MISP, OpenCTI, The-Hive, ...)
- Open-source software provides a low barrier for partner organisations to collaborate and contribute. 
- Open Source software and tooling can easily be brought in as background into projects and partnerships such as HEU, EDF, or National funded projects and others. The use of open-source tooling is explicitly encouraged by the European Commission.


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

SOARCA can be extended with custom extensions or rather so-called FIN (inspired by the majestic orca). A fin can integrate with the SOARCA core. (Technical descriptions of the components can be found [here]()). Fins communicate with our SOARCA core using a pre-defined MQTT protocol. 



## Join the SOARCA Community

The SOARCA team invites cybersecurity professionals, researchers, and enthusiasts to join our community. Explore, adapt, and contribute to SOARCA’s growth. Let’s fortify cyber defenses together! See our [contribution guidelines](/docs/contribution-guidelines/) on how to make contributions.	🛡️🌐

## Key Details
- Project Name: SOARCA (Security Orchestrator for Advanced Response to Cyber Attacks)
- License: Apache 2.0
- Release Date: March 2024
