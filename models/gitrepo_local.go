package models

const BaseDir = "/tmp"
const DummyUsername = "dummy"

type GitProject struct {
	DownloadDir     string
	LatestCommitSha string
}
