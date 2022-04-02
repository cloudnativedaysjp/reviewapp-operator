reviewapp-operator
===

![License](https://img.shields.io/github/license/cloudnativedaysjp/reviewapp-operator)
![Go Report Card](https://goreportcard.com/badge/github.com/cloudnativedaysjp/reviewapp-operator)

[reviewapp-operator](https://github.com/cloudnativedaysjp/reviewapp-operator) works in concert with the [Argo CD](https://github.com/argoproj/argo-cd) to realize "launching a new environment for development (as Review Apps environment) when any PullRequest is opened in the application-repository".
[reviewapp-operator](https://github.com/cloudnativedaysjp/reviewapp-operator) is mainly responsible for "creating and deleting manifests to the manifest-repository when PullRequests in the application-repository are updated," and [Argo CD](https://github.com/argoproj/argo-cd) is responsible for actually applying the manifests from the manifest-repository to Kubernetes.

![https://raw.githubusercontent.com/ShotaKitazawa/zenn-articles/master/images/about-reviewapp-operator/workflow.jpg](workflow of reviewapp-operator)

### Installation

please read [Release](https://github.com/cloudnativedaysjp/reviewapp-operator/releases)

### For more informations

[/docs](https://github.com/cloudnativedaysjp/reviewapp-operator/tree/main/docs)

### Reference

* [PR が出るたびに Kubernetes 上で dev 環境を立ち上げるための Kubernetes Operator](https://zenn.dev/kanatakita/articles/about-reviewapp-operator) - Japanese

### Licence

[MIT](https://github.com/cloudnativedaysjp/reviewapp-operator/tree/main/LICENCE)
