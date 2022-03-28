package kubernetes

import (
	"context"
	"sort"

	"golang.org/x/xerrors"
	batchv1 "k8s.io/api/batch/v1"
	"k8s.io/apimachinery/pkg/labels"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (c Client) GetLatestJobFromLabel(ctx context.Context, namespace, labelKey, labelValue string) (*batchv1.Job, error) {
	var jList batchv1.JobList
	if err := c.List(ctx, &jList, &client.ListOptions{
		Namespace:     namespace,
		LabelSelector: labels.SelectorFromSet(map[string]string{labelKey: labelValue}),
	}); err != nil {
		wrapedErr := xerrors.Errorf("Error to get Job: %w", err)
		return nil, wrapedErr
	}
	sort.Slice(jList.Items, func(i, j int) bool {
		return jList.Items[i].CreationTimestamp.Before(&jList.Items[j].CreationTimestamp)
	})
	return &jList.Items[0], nil
}

func (c Client) CreateJob(ctx context.Context, job *batchv1.Job) error {
	if err := c.Create(ctx, job); err != nil {
		return err
	}
	return nil
}
