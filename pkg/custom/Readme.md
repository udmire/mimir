# Purpose

* Add admin to manage cluster, tenant, policy, token
    * [ ] Configuration management
    * [x] Cluster manage API.
    * [x] Tenant manage API
    * [x] AccessPolicy manage API
    * [x] Token manage API
    * [ ] Auto create cluster when starting
    * [x] Support generate the default built-in admin policy
* Add auto proxy configuration base ring information.
    * [x] Configuration management
    * [ ] Support Proxy for backends
* Support auto dinaginoze for the whole application status.
    * [x] Configuration management
    * [ ] Support centralized authentication with OIDC
    *
* Add authentication feature.
    * [x] Configuration management
    * [ ] Support override token with config
    * [x] Support integrate with oidc
    * [x] Support enterprise auth type
    * [x] Support trust auth type
    * [x] Support Basic Auth Flow
    * [x] Support Bearer Token Auth Flow
    * [x] Support default policy
    * [ ] Support cache auth result
    * [x] Support validate token
    * [x] Support ignore Public routes
* Token generate with default built-in admin policy.
    * [x] Configuration management
    * [x] Support specify policy name
    * [x] Support multiple time token generate.
    * [x] Store generated tokens to admin object store.
* License management
    * [x] Configuration management