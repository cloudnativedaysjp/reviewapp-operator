package models

type GitCredential struct {
	username string
	token    string
}

func NewGitCredential(username, token string) GitCredential {
	return GitCredential{username, token}
}

func (m GitCredential) Username() string {
	return m.username
}

func (m GitCredential) Token() string {
	return m.token
}
