/*
Copyright 2021 The UnDistro authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package contour

const PatchLocal = `{"spec":{"template":{"spec":{"nodeSelector":{"ingress-ready":"true"},"tolerations":[{"key":"node-role.kubernetes.io/master","operator":"Equal","effect":"NoSchedule"}]}}}}`

var TestResources = `---
apiVersion: projectcontour.io/v1
kind: HTTPProxy
metadata:
  name: undistro-ingress
spec:
  virtualhost:
    fqdn: "*"
    tls:
      passthrough: true
    corsPolicy:
        allowCredentials: true
        allowOrigin:
          - "*" # allows any origin
        allowMethods:
          - "*"
        allowHeaders:
          - "*"
        exposeHeaders:
          - "*"
        maxAge: "10m" # preflight requests can be cached for 10 minutes.
  routes:
    - conditions:
      - prefix: /
      enableWebsockets: true
      services:
        - name: undistro-webhook-service
          port: 2020`
