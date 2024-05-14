---
date: 2024-05-14
title: "Automating Security Responses: Introducing Conditional Logic Support in SOARCA - Demo"
linkTitle: Introducing Conditional Logic Support in SOARCA - Demo
description: >

author: RabbITCybErSeC, Hidde-Jan Jongsma, MaartendeKruijf
resources:
  - src: "**.{png,jpg}"
    title: "Image #:counter"
    params:
      byline: ""
---
Check out our latest demo featuring the (beta) integration of the open-source [CACAO Roaster tool](https://github.com/opencybersecurityalliance/cacao-roaster) available on GitHub, and playbook 'if condition' step support for conditional logic in playbook executions. This enables security operators to create dynamic playbooks that execute appropriate actions based on the outcomes of preceding playbook steps, allowing for precise and efficient responses to security incidents.

Despite the two project being developed by separate teams, the (beta version) integration between CACAO Roaster and SOARCA showcases the advantages of adopting the standardized CACAO V2.0 specification, making security solutions more interoperable. 

In the demo, we present a CACAO playbook for the automated mitigation of a malicious webshell. The playbook allows to scan our infrastructure for malicious webshells, promptly taking action by terminating any identified webshell processes and removing associated binaries. Finally, the playbook reports to Slack.

The integration between SOARCA and third-parties can be facilitated through the [playbook API](https://cossas.github.io/SOARCA/docs/soarca-api/), allowing third-parties to execute stored or uploaded playbooks. SOARCA performs all the steps in the workflow. Depending on the necessary actions, SOARCA can execute system reconfigurations over SSH, HTTP, and [OpenC2](https://openc2.org/), with plans for further integrations and extensions in the pipeline.

Notably, the demo highlights recent enhancements to SOARCA's core features, for example the ["if condition feature"](https://github.com/COSSAS/SOARCA/pull/138). This feature, showcased in the demo, enables the execution of conditional logic embedded within playbooks, enhancing flexibility and customization options for security operations. Another [milestone](https://github.com/COSSAS/SOARCA/milestones) for the SOARCA's 1.1 release is the support for while condition execution steps as described [here](https://github.com/COSSAS/SOARCA/issues/143). 

Example demo files can be found [here](https://github.com/MaartendeKruijf/soarca-webshell-example).


<iframe src="https://player.vimeo.com/video/946107969?h=0114d86628" width="640" height="360" frameborder="0" allow="autoplay; fullscreen; picture-in-picture" allowfullscreen></iframe>
<p><a href="https://vimeo.com/946107969">SOARCA webshell demonstration</a> from <a href="https://vimeo.com/user216437450">COSSAS</a> on <a href="https://vimeo.com">Vimeo</a>.</p>