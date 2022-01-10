package kubernetes

import (
	"context"
	"sort"

	myerrors "github.com/cloudnativedaysjp/reviewapp-operator/errors"
	"golang.org/x/xerrors"
	batchv1 "k8s.io/api/batch/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"
)

func GetLatestJobFromLabel(ctx context.Context, c client.Client, namespace, labelKey, labelValue string) (*batchv1.Job, error) {
	var jList batchv1.JobList
	if err := c.List(ctx, &jList, &client.ListOptions{
		Namespace:     namespace,
		LabelSelector: labels.SelectorFromSet(map[string]string{labelKey: labelValue}),
	}); err != nil {
		wrapedErr := xerrors.Errorf("Error to get Job: %w", err)
		if apierrors.IsNotFound(err) {
			return nil, myerrors.K8sResourceNotFound{Err: wrapedErr}
		}
		return nil, wrapedErr
	}
	sort.Slice(jList.Items, func(i, j int) bool {
		return jList.Items[i].CreationTimestamp.Before(&jList.Items[j].CreationTimestamp)
	})
	return &jList.Items[0], nil
}

func CreateJob(ctx context.Context, c client.Client, job *batchv1.Job) error {
	if err := c.Create(ctx, job); err != nil {
		return err
	}
	return nil
}

func PickJobObjectFromObjectStr(ctx context.Context, objectStr string) (*batchv1.Job, error) {
	var job batchv1.Job
	err := yaml.Unmarshal([]byte(objectStr), &job)
	if err != nil {
		return nil, xerrors.Errorf("%w", err)
	}
	return &job, nil
}
