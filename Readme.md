# OAuth Group Sync Operator for Openshift

This Operator synchronizes users, authenticated from different OAuth providers to groups named after their OAuth provider name. This makes permission management easy, as all the necessary Roles can be bound to the groups instead of individual users. Additionally, all users will be put into their groups immediately after logging in for the first time.

## Openshift OAuth Configuration
Let's take a look at the following OAuth configuration in Openshift:

```yaml
oauthConfig:
  identityProviders:
    - name: test1
      mappingMethod: claim
      provider: {...}
    - name: test2
      mappingMethod: claim
      provider: {...}
```

Every user from the `test1` and `test2` identity provider, will be able to log into your openshift cluster. As soon as a new user logs in, a new `User` resource will be created in openshift.  

The operator recognizes the user creation, and will add the user to the appropriate group. For example: If the user uses the `test1` identity provider, the user will be added to the `test1` group.  

## Installation
```bash
helm repo add gp-helm-charts https://gepaplexx.github.io/gp-helm-charts/
helm upgrade --install -n gp-group-sync --create-namespace sync-operator gp-helm-charts/oauth-group-sync-operator
```
