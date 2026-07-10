---
title: LLM Playbook Generation
weight: 4
description: >
  The how to on playbook generation using LLMs
tags: [playbooks, cacaov2, soarca, LLM, AI]
---

Creating playbooks has long been a bottleneck on automation. With the introduction LLMs the effort of generating text and code has been significantly reduced. Now this generation comes to SOARCA. In the repo one can find the [grammar](https://github.com/COSSAS/SOARCA/blob/development/SOARCA-PLAYBOOK-GRAMMAR-CACAO-V2.md) which can be ingested into an LLM as starting prompt. After that natural language can be used to describe the playbook. In a few minutes the playbook will be produced. The grammar has been tested with Claude Opus 4.7/4.8 and ChatGPT 5.4. But we expect it to work with other models. 