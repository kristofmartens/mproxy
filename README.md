# MPROXY
A simple reverse proxy that focusses purely on OIDC based service authorization. This proxy does not handle the OIDC 
authentication flows and assumes that the access tokens are already available in the incoming requests. This is often
the case if you are running services in a kubernetes cluster where all access to these services is authenticated via a 
load balancer or API gateway. All finer grained access needs to be authorized at the service level.

The rationale for building this proxy was that for some reason I did not find a flexible enough reverse proxy that 
could easily authorize access to micro services for different OIDC providers (more specifically AWS Cognito) that 
use custom claims in their access tokens.

## How to build and install
Make sure that the go tools and your go path are set correctly. Building and installing should be as simple as:
```shell script
# Building from local version
cd cmd/mproxy && go get && go install

# Build directly from github, put binary somewhere in $PATH
go get github.com/kristofmartens/mproxy/cmd/mproxy
go build github.com/kristofmartens/mproxy/cmd/mproxy
```
There is also a Dockerfile available for running this image in your kubernetes cluster. Using this proxy as a side-car
to authorize your services in a kubernetes cluster.

For convenience I also added a helm chart to easily install the reverse proxy with your micro services in kubernetes.
For more information about how to use and install helm please go to https://helm.sh/

## Running the proxy
MProxy requires a config file in yaml format only requiring a small set of parameters. You start the proxy like this:
```shell script
mproxy --config <config-file>
```
Here is an example configuration file for use with AWS Cognito:
```yaml
# Local port where the mproxy will listen on (default 8080)
localPort: 8080
# Location to forward all traffic to (mandatory)
destination: http://localhost:80
# Discovery URL of where to retrieve all OAUTH/OIDC endpoints (mandatory)
discoveryUrl: https://cognito-idp.<region>.amazonaws.com/<cognito-user-pool-id>
# HTTP header that contains the access token (mandatory)
accessTokenHeader: X-Amzn-Oidc-Accesstoken
# Defines the authorization proxy rules (default will allow all authenticated traffic)
proxyRules:
    # URL pattern to authorise: This rule will allaw all authenticated users that are in the cognito group admin OR user 
    # access to all paths to the microservice 
  - pattern: /
    # Checks the claims of the access token to verify whether all are valid
    claims:
        # The name of the claim to verify
      - claimName: cognito:groups
        # Allowed values in the claim, regular expressions are also allowed!
        allowedClaims:
          - admin
          - user
          - regex_group_[a-z|0-9]{3}
        # If true, all claims need to be present. Otherwise only one claim is required
        requireAllClaims: false
# Liveliness path to check whether service is still up
livelinessPath: /alive
```

An example values.yaml file for authorizing microservices on kubernetes with AWS Cognito as OIDC provider:
```yaml
replicaCount: 1

image:
  repository: martensk/mproxy
  tag: v0.1.0
  pullPolicy: Always

mproxy:
  config:
    destination: http://<service-name>.<namespace>.svc.cluster.local:<service-port>
    discoveryUrl: https://cognito-idp.<region>.amazonaws.com/<<cognito-user-pool-id>>
    accessTokenHeader: X-Amzn-Oidc-Accesstoken
    proxyRules:
      - pattern: /
        claims:
          - claimName: cognito:groups
            allowedClaims:
              - admin
              - user
            requireAllClaims: false

ingress:
  enabled: true
  annotations: {}
  hosts:
    - host: <service-name>.yourdomain.com
      paths:
        - /
  tls: []

resources:
  limits:
    cpu: 50m
    memory: 64Mi
  requests:
    cpu: 50m
    memory: 64Mi
```
