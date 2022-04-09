package kubernetes

import (
	"github.com/go-logr/logr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Client struct {
	client.Client

	logger logr.Logger
}

func NewClient(l logr.Logger, c client.Client) *Client {
	return &Client{c, l}
}
