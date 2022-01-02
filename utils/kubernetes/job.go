package kubernetes

import (
	"context"
	"reflect"

	myerrors "github.com/cloudnativedaysjp/reviewapp-operator/errors"
	"golang.org/x/xerrors"
	batchv1 "k8s.io/api/batch/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"
)

func GetJob(ctx context.Context, c client.Client, namespace, name string) (*batchv1.Job, error) {
	var j batchv1.Job
	if err := c.Get(ctx, types.NamespacedName{Name: name, Namespace: namespace}, &j); err != nil {
		wrapedErr := xerrors.Errorf("Error to get %s: %w", reflect.TypeOf(j), err)
		if apierrors.IsNotFound(err) {
			return nil, myerrors.K8sResourceNotFound{Err: wrapedErr}
		}
		return nil, wrapedErr
	}
	return &j, nil
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
