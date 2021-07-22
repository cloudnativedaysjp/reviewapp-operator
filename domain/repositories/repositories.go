package repositories

type GitRepoApp struct {
	PullRequestAppIFace PullRequestIFace
}

func NewPullRequestApp(pr PullRequestIFace) *GitRepoApp {
	return &GitRepoApp{pr}
}

type GitRepoInfra struct {
	GitCodeIFace GitCodeIFace
}

func NewGitRepoInfra(gc GitCodeIFace) *GitRepoInfra {
	return &GitRepoInfra{gc}
}

type KubernetesIFaces struct {
	ReviewAppConfigIFace   ReviewAppConfigIFace
	ReviewAppInstanceIFace ReviewAppInstanceIFace
	ArgoCDApplictionIFace  ArgoCDApplictionIFace
	SecretIFace            SecretIFace
}

func NewKubernetes(rac ReviewAppConfigIFace, rai ReviewAppInstanceIFace, app ArgoCDApplictionIFace, secret SecretIFace) *KubernetesIFaces {
	return &KubernetesIFaces{rac, rai, app, secret}
}
