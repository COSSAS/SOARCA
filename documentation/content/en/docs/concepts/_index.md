---
title: SOARCA vision & Concepts
weight: 3
description: >
  Why SOARCA?
resources:
- src: "*Slide2.jpg"
  params:
    byline: "*Slide*: © 2024 TNO"
---

## Background of SOARCA

**S**ecurity **O**rchestrator for **A**dvanced **R**esponse to **C**yber **A**ttacks​ - SOARCA

SOARCA is [TNO’s](https://www.tno.nl/nl/) new Open Source SOAR (Security Orchestration, Automation and Response) tool, which has been developed for research and demonstration purposes. With SOARCA, TNO’s goal is to realise and stimulate advanced cyber security innovations and empower end users and organizations by providing a vendor-agnostic, extensible, and standards-compliant solution for security orchestration. SOARCA will be open sourced  on COSSAS (Community for Open Source Security Automation Software – see also [COSSAS](https://cossas-project.org/) with the [Apache 2.0 licence](https://www.apache.org/licenses/LICENSE-2.0).​ 

While there are already several mature SOAR tools available on the market, many of them are commercial closed-source products, and none of them complies with the new emerging OASIS Open standards. TNO’s easily accessible “SOARCA” bridges this gap to let end users and organisations get hands-on experience with SOAR tooling and enable new innovations: it is vendor-agnostic, extensible and has open and well-defined interfaces. SOARCA will be available Open Source and in it's first phase can be used and adapted freely for research, demonstrations and PoC purposes. The goal is to grow the SOARCA community and getting SOARCA more mature. SOARCA is designed to fully comply with the newest standards [CACAO v2.0](https://docs.oasis-open.org/cacao/security-playbooks/v2.0/security-playbooks-v2.0.html) and [OpenC2](https://openc2.org/).​

Note that that open and accessible SOAR functionality is relevant not only for automation in cyber incident response handling, but also attack & defense simulations, cyber ranges, digital twinning and other growing innovation topics that have a strong dependence on the orchestration of complex workflows.


{{% imgproc Slide2 Fill "1280x720" %}}

{{% /imgproc %}}

## Vision of SOARCA

### Why Soarca?

Both inside and outside of TNO there is a strong need for interoperable workflow orchestration tooling that aids (cybersecurity) innovation. High-quality SOAR (Security Orchestration, Automation and Response) tools are widely available in the market, however these are commercial products with significant license costs and that employ proprietary technologies rather than the emerging innovative standards.


- **Vendor-Agnostic Compatibility**: Our solution ensures seamless integration with various vendors, eliminating reliance on a single provider.
- **Standard Compliance**: Adhering to the latest standards, including [CACAO v2.0](https://docs.oasis-open.org/cacao/security-playbooks/v2.0/security-playbooks-v2.0.html) and [OpenC2](https://openc2.org/), guarantees up-to-date and secure operations.
- **Extensibility with Open Interfaces**: Enjoy the flexibility of an extensible tool featuring open and well-defined interfaces, promoting adaptability and customization.
- **Open-Source Affordability**: Embrace an open-source model that not only offers cost-effective solutions but also supports unrestricted use and adaptation for research purposes.


SOAR functionality is relevant not only for automation in incident response handling, but also attack & defense simulations, cyber ranges, digital twinning and other (TNO research) topics that have a strong dependence on the orchestration of complex workflows. 

### Current state of Soarca

At present, SOARCA is in an Alpha release phase and is intended for Proof of Concepts (PoCs) and research purposes, serving as a platform for demonstrations. The objective of the SOARCA team is to evolve SOARCA into a more mature SOAR orchestration tool suitable for operational environments. For potential applications of SOARCA, please refer to the ‘Use-Cases’ section of our documentation.

### Why making Soarca open-source?

- SOARCA has been publicly funded and should therefore ideally be made publicly available.
- The target audience of SOC, CERT/CSIRT and CTI teams has a very strong affinity with Open Source solutions and embraces them to a great extent. (see also the success of MISP, OpenCTI, The-Hive...)
- OS SOARCA software provides a low barrier for partner organisations to collaborate with TNO and contribute to further development.
- Open Source software and tooling can easily be brought in as background into projects and partnerships such as HEU, EDF, TKI projects and others. The use of Open Source tooling is explicitly encouraged by the European Commission.


## Core Concepts

There are several concepts within SOARCA that might be important to know. 

### SoC Playbooks

`to be added`

### CACAO Playbooks: Streamlining Cybersecurity Operations
A CACAO playbook is a structured and standardized document that outlines a series of orchestrated actions to address specific security events, incidents, or other security-related activities. These playbooks allow for the automation of security steps.

Example of repetive tasks that might be automated using a CACAO Playbook might be:

- Investigate the cause of security events.
- Mitigate threats effectively.
- Remediate vulnerabilities.

By following CACAO playbooks specification, organizations can enhance their automated response capabilities, foster collaboration, and ensure consistentcy of playbooks across diverse technological solutions.

Learn more about CACAO and its schema in the [CACAO Security Playbooks Version 2.0 specification](https://docs.oasis-open.org/cacao/security-playbooks/v2.0/security-playbooks-v2.0.html).

### SOARCA Fin(s): Extending the core Soarca capabilities

SOARCA can be extended with custom extensions or rather so-called FIN (inspired by the majestic orca). A fin can integrate with the SOARCA core. (Technical descriptions of the components can be found [here]()). Fins communicate with our SOARCA core using pre-defined MQTT protocol. 

### Coarse of Action

A course of action step refers to a planned sequence of steps or activities taken to achieve a specific cyber security goal.

## Join the SOARCA Community

The SOARCA team invites cybersecurity professionals, researchers, and enthusiasts to join our community. Explore, adapt, and contribute to SOARCA’s growth. Let’s fortify cyber defenses together! See our [contribution guidelines](/contribution-guidelines/) on how to make contributions.	🛡️🌐

## Key Details
- Project Name: SOARCA (Security Orchestrator for Advanced Response to Cyber Attacks)
- License: Apache 2.0
- Release Date: March 2024
