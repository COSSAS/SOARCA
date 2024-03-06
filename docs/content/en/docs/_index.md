---
title: SOARCA Documentation
linkTitle: Docs
menu: {main: {weight: 20}}
weight: 20
---


{{% alert title="Warning" color="warning" %}}
SOARCA is currently in its **alpha release**, with ongoing evelopment aimed at expanding its capabilities, improving integration, and enhancing its functionalities. You can track our progress and upcoming milestones at [LINK TO ROADMAP].

We warmly welcome contributions to our repository. You can find the guidelines for contributing [here](/docs/contribution-guidelines).
{{% /alert %}}

SOARCA, an open-source SOAR (Security Orchestration, Automation and Response) tool developed by TNO, is designed be vendor-agnostic, allowing it to orchestrate various security actuators and systems. SOARCA is the first SOAR that aims to be compliant with the [CACAO v2.0](https://docs.oasis-open.org/cacao/security-playbooks/v2.0/security-playbooks-v2.0.html) standard.

SOARCA enables cyber defenders to coordinate and automate their cyber operations, by using executable CACAO playbooks.

SOARCA aims to achieve the following goals:

- **Standard Compliance**: Adhering to the latest standards, including [CACAO v2.0](https://docs.oasis-open.org/cacao/security-playbooks/v2.0/security-playbooks-v2.0.html) and [OpenC2](https://openc2.org/), allows for interopability with a wide range of technologies.
- **Extensibility with Open Interfaces**: Enjoy the flexibility of an extensible tool featuring open and well-defined interfaces, promoting adaptability, customization, and experimentation.
- **Open-Source**: Embrace an open-source model that not only offers cost-effective solutions but also supports unrestricted use and adaptation for research purposes.


Interested in the vision and concepts of SOARCA? Then check the [SOARCA vision and concepts](/docs/concepts/).


## SOARCA capabilities

SOARCA currently supports the following transport mechanisms:

<div class="works-well-with">
{{< cardpane >}}
{{% card header="OpenC2 - Native" %}}
[![OpenC2](/images/logos-external/openc2.svg)](/docs/soarca-extentions/native-capabilities/#openc2-capability)
{{% /card %}}

{{% card header="HTTP - Native" %}}
[![Http](/images/logos-external/http.svg)](/docs/soarca-extentions/native-capabilities/#http-api-capability)
{{% /card %}}

{{% card header="SSH - Native" %}}
[![Ssh](/images/logos-external/ssh.svg)](/docs/soarca-extentions/native-capabilities/#ssh-capability)
{{% /card %}}
{{< /cardpane >}}
</div>


## Features of SOARCA



## Where do I start?

{{% alert title="primary" color="primary" %}}
Following our [Getting started](/docs/getting-started/) guide will help you setup SOARCA and configure the SOAR for your internal security tooling. For more custom requirement 
{{% /alert %}}
