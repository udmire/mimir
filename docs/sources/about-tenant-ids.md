---
title: "About tenant IDs"
description: ""
weight: 10
---

# About tenant IDs

Within a Grafana Mimir cluster, the tenant ID is the unique identifier of a tenant.
For information about how Grafana Mimir components use tenant IDs, refer to [About authentication and authorization]({{<relref "./about-authentication-and-authorization.md" >}}).

## Restrictions

Tenant IDs must be less-than or equal-to 150 bytes or characters in length and must comprise only supported characters:

- Alphanumeric characters
  - `0-9`
  - `a-z`
  - `A-Z`
- Special characters
  - Exclamation point (`!`)
  - Hyphen (`-`)
  - Underscore (`_`)
  - Single period (`.`)
  - Asterisk (`*`)
  - Single quote (`'`)
  - Open parenthesis (`(`)
  - Close parenthesis (`)`)

> **Note:** For security reasons, `.` and `..` are not valid tenant IDs.

All other characters, including slashes and whitespace, are not supported.