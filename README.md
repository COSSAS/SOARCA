<div align="center">
<a href="https://cossas-project.org/cossas-software/soarca"><img src="img/soarca-logo.svg"/>


[![https://cossas-project.org/portfolio/SOARCA/](https://img.shields.io/badge/website-cossas--project.org-orange)](https://cossas-project.org/portfolio/SOARCA/)
[![Pipeline status](https://github.com/cossas/soarca/actions/workflows/ci.yml/badge.svg?development)](https://github.com/COSSAS/SOARCA/actions)
[![License](https://img.shields.io/badge/License-Apache_2.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)
</div></a>


Automate threat and incident response workflows with CACAO security playbooks	

## Context and backgound 

Organisations are increasingly automating threat and incident response through playbook driven security workflow orchestration. The essence of this concept is that specific security events trigger a predefined series of response actions that are executed with no or only limited human intervention. These automated workflows are captured in machine-readable security playbooks, which are typically executed by a so called Security Orchestration, Automation and Response (SOAR) tool. The market for SOAR solutions has matured significantly over the past years and present day products support sophisticated automation workflows and a wide array of integrations with external security tools and data resources. Typically, however, the technology employed is proprietary and not easily adaptable for research and experimentation purposes. SOARCA aims to offer an open-source alternative for such solutions that is free of vendor dependencies and supports standardized formats and technologies where applicable.

SOARCA was developed for research and innovation purposes and allows SOC, CERT and CTI professionals to experiment with the concept of playbook driven security automation. It is open and extensible and its interfaces are well-defined and elaborately documented. Importantly, it offers native support for the emerging technology standards CACAOv2 and OpenC2, both developed and maintained by OASIS Open. CACAO (Collaborative Automated Course of Action Operations) provides a standardized scheme for machine-readable security playbooks while OpenC2 offers a standardized language for the command and control of cyber defense technologies (e.g. firewalls or IAM solutions).


## Software
SOARCA is a security orchestrator that can ingest, validate and execute CACAOv2 security playbooks. These playbooks and the triggers for their execution are consumed via a JSON API. SOARCA comes with native http(s), SSH and OpenC2 capabilities to interface with external tools and data resources. These native capabilities can be extended via a dedicated MQTT interface, allowing developers to compile additional integrations according their needs.

Development is ongoing. The current version solely supports machine and command line interfaces, but a graphical user interface will be added in the foreseeable future. Furthermore, its current capability to run CACAOv2 playbooks sequentially will evolve towards the ability to run multiple playbooks in parallel. Such further developments will be announced and published on the SOARCA repository on Github.


## Documentation

For the latest documentation we refer to our [Github pages](https://cossas.github.io/SOARCA/).


## Source Project

More information on the source of the project can be found [here](https://cossas.github.io/SOARCA/docs/about/).