package provisioning

import (
	"context"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	apierr "k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type kymaVersionConfigurator struct {
	ctx       context.Context
	k8sClient client.Client

	namespace string
	name      string

	log logrus.FieldLogger
}

func NewKymaVersionConfigurator(ctx context.Context,
	cli client.Client, namespace, name string, log logrus.FieldLogger) KymaVersionConfigurator {

	return &kymaVersionConfigurator{
		ctx:       ctx,
		namespace: namespace,
		name:      name,
		k8sClient: cli,
		log:       log,
	}
}

func (c *kymaVersionConfigurator) ForGlobalAccount(gaID string) (string, bool, error) {
	config := &v1.ConfigMap{}
	err := c.k8sClient.Get(c.ctx, client.ObjectKey{
		Namespace: c.namespace,
		Name:      c.name,
	}, config)

	switch {
	case apierr.IsNotFound(err):
		c.log.Infof("Kyma Version per Global Acocunt configuration %s/%s not found", c.namespace, c.name)
		return "", false, nil
	case err != nil:
		return "", false, errors.Wrap(err, "while getting kyma version config map")
	}

	ver, found := config.Data[gaID]
	return ver, found, nil
}
