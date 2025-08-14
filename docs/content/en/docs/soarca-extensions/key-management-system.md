---
title: Key management system
description: >
    Organizing keys used in SSH connections
categories: [capabilities]
tags: [native]
weight: 2
date: 2023-01-05
---

This page details the key management system that is built into SOARCA. 
It currently works only in conjunction with the [SSH capability](./native-capabilities/#SSH).

## Activation

The KMS feature of SOARCA is enabled by setting [environment variables](../installation-configuration/#key-management-system).

## Use in SSH commands

The KMS can be referenced inside a playbook inside a [user-auth](https://docs.oasis-open.org/cacao/security-playbooks/v2.0/cs01/security-playbooks-v2.0-cs01.html#_Toc152256508) element.
This element is referenced by the [ssh command](https://docs.oasis-open.org/cacao/security-playbooks/v2.0/cs01/security-playbooks-v2.0-cs01.html#_Toc152256500).
To use the KMS the value kms must be set to true, and the value kms_key_identifier must be set to the name of the key. 

The use of the KMS overrides a specified password.

## Underlying structure

The key management system relies on an underlying folder with public and private keys. 
The name of the key referenced by the user in the kms_key_identifier field is the name of the private key file.
SOARCA caches the keys, and the API allows the user to refresh the system, which loads any potential new key files found in the underlying directory.
Adding a key through the API also creates new files for storing the keys.

The system also allows the user to revoke keys through the API.
This moves the keys to a directory called .revoked inside the key storage, and appends the time and date of the revocation to the key name.
It is up to the user to actually delete these keys or to recover keys.

## API

The API endpoints for the KMS are documented [here](/docs/soarca-api/#key-management-system)

