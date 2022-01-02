package kubernetes

import (
	"context"

	"golang.org/x/xerrors"
	batchv1 "k8s.io/api/batch/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"
)

func PickJobObjectFromObjectStr(ctx context.Context, objectStr string) (*batchv1.Job, error) {
	var job batchv1.Job
	err := yaml.Unmarshal([]byte(objectStr), &job)
	if err != nil {
		return nil, xerrors.Errorf("%w", err)
	}
	return &job, nil
}

func CreateJob(ctx context.Context, c client.Client, job *batchv1.Job) error {
	if err := c.Create(ctx, job); err != nil {
		return err
	}
	return nil
}
