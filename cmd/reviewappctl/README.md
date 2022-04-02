# reviewappctl

reviewappctl is tools for management reviewapp-operator

### manifests-templating

```
Usage:
  reviewappctl manifests-templating [flags]

Flags:
  -h, --help               help for manifests-templating
      --is-candidate       using Candidate template if this flag is true
      --is-stable          using Stable template if this flag is true
  -f, --load string        filename of manifests based on ManifestsTemplate (default "manifests_template.yaml")
      --name string        name of ManifestsTemplate
      --namespace string   namespace of ManifestsTemplate (default "default")
      --validate           validate manifests if this flag is true (default true)
```

This command generate `ManifestsTemplate` CustomResource manifest from some manifests.

Please find below examples.

* called `reviewappctl manifests-templating` in Argo CD Plugin [in this file](https://github.com/cloudnativedaysjp/dreamkast-infra/blob/main/manifests/argocd/overlays/dev/argocd-cm.yaml).
* deployed Argo CD Appliction with reviewappctl Plugin [in this file](https://github.com/cloudnativedaysjp/dreamkast-infra/blob/main/manifests/reviewapps/argocd-apps/dreamkast.yaml)
* above Argo CD Application manage [this directory](https://github.com/cloudnativedaysjp/dreamkast-infra/tree/main/manifests/app/dreamkast/overlays/development/template-dk). Generate `ManifestsTemplate` manifest by run reviewappctl Plugin before Argo CD apply manifests to Kubernetes.
