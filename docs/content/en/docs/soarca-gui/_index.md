---
title: SOARCA-GUI
description: >
      
categories: [extensions, architecture, soarca-gui]
tags: []
weight: 7
---


{{% alert title="Warning" color="warning" %}}
SOARCA-GUI is currently in its **first release**, with ongoing development aimed at expanding its capabilities, improving integration, and enhancing its functionalities. 

We warmly welcome contributions to our [repository](https://github.com/COSSAS/SOARCA-GUI). You can find the guidelines for contributing [here](/docs/contribution-guidelines).
{{% /alert %}}

SOARCA can now work with a front-end interface called the SOARCA-GUI (written in the [GoTTH stack](https://github.com/TomDoesTech/GOTTH) for simplicity), which can be found in a separate [repository](https://github.com/COSSAS/SOARCA-GUI). The SOARCA-GUI is designed to assist administrators and analysts in tracking executions and providing manual inputs when specific action steps require decision-making. In its first version, the SOARCA-GUI allows users to track the execution of playbooks. 

Our long-term vision for the SOARCA-GUI includes enabling users to configure SOARCA directly, test integrations using tools like the SOARCA Fin library, and manage these tasks without requiring terminal commands or interventions. Additionally, we plan to introduce functionality for viewing and managing playbooks in a future version of the interface.

The SOARCA-GUI features OIDC-based login for authentication and authorization. Similar to SOARCA, the SOARCA-GUI uses the [gauth](https://github.com/COSSAS/gauth) library as authentication & authorization middleware. This middleware is known to work with [Authentik](https://goauthentik.io/). For more information on setting up authentication for SOARCA, please refer to the documentation [here](/docs/installation-configuration/authentication.md). Authentication only works when enabling OIDC, as such if you want to have authentication you are required to setup Authentik or a different OIDC-providers. Note, that other OIDC-providers have not been tested yet. 



