package models

type GitCredential struct {
	username string
	token    string
}

func NewGitCredential(username, token string) GitCredential {
	return GitCredential{username, token}
}

func (m GitCredential) GetUsername() string {
	return m.username
}

func (m GitCredential) GetToken() string {
	return m.token
}
