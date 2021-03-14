# Architecture

The overarching architecture of UnDistro is centered around a "management plane". This plane is expected to serve as a single interface upon which administrators can create, scale, upgrade, and delete Kubernetes clusters. At a high level view, the management plane + created clusters should look something like this:
![Image of Architecture](./assets/arch.png)