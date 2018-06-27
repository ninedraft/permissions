package clients

import (
	"context"
	"fmt"
	"net/url"

	"git.containerum.net/ch/permissions/pkg/errors"
	"github.com/containerum/cherry"
	"github.com/containerum/cherry/adaptors/cherrylog"
	"github.com/containerum/kube-client/pkg/model"
	"github.com/containerum/utils/httputil"
	"github.com/json-iterator/go"
	"github.com/sirupsen/logrus"
	"gopkg.in/resty.v1"
)

type KubeAPIClient interface {
	CreateNamespace(ctx context.Context, projectID string, req model.Namespace) error
	SetNamespaceQuota(ctx context.Context, projectID string, ns model.Namespace) error
	DeleteNamespace(ctx context.Context, projectID string, ns model.Namespace) error
	GetNamespace(ctx context.Context, projectID string, name string) (model.Namespace, error)
}

type KubeAPIHTTPClient struct {
	log    *cherrylog.LogrusAdapter
	client *resty.Client
}

func NewKubeAPIHTTPClient(url *url.URL) *KubeAPIHTTPClient {
	log := logrus.WithField("component", "kube_api_client")

	client := resty.New().
		SetLogger(log.WriterLevel(logrus.DebugLevel)).
		SetHostURL(url.String()).
		SetDebug(true).
		SetError(cherry.Err{}).
		SetHeader("Content-Type", "application/json").
		SetHeader("Accept", "application/json")
	client.JSONMarshal = jsoniter.Marshal
	client.JSONUnmarshal = jsoniter.Unmarshal
	return &KubeAPIHTTPClient{
		log:    cherrylog.NewLogrusAdapter(log),
		client: client,
	}
}

func (k *KubeAPIHTTPClient) CreateNamespace(ctx context.Context, projectID string, req model.Namespace) error {
	k.log.WithFields(logrus.Fields{
		"project_id": projectID,
		"cpu":        req.Resources.Hard.CPU,
		"memory":     req.Resources.Hard.Memory,
		"label":      req.Label,
		"name":       req.ID,
		"access":     req.Access,
	}).Debug("create namespace")

	resp, err := k.client.R().
		SetBody(req).
		SetContext(ctx).
		SetHeaders(httputil.RequestXHeadersMap(ctx)).
		SetPathParams(map[string]string{"project": projectID}).
		Post("/projects/{project}/namespaces")
	if err != nil {
		return errors.ErrInternal().Log(err, k.log)
	}
	if resp.Error() != nil {
		return resp.Error().(*cherry.Err)
	}
	return nil
}

func (k *KubeAPIHTTPClient) SetNamespaceQuota(ctx context.Context, projectID string, ns model.Namespace) error {
	k.log.WithFields(logrus.Fields{
		"project_id": projectID,
		"cpu":        ns.Resources.Hard.CPU,
		"memory":     ns.Resources.Hard.Memory,
		"label":      ns.Label,
		"name":       ns.ID,
	}).Debug("set namespace quota")

	resp, err := k.client.R().
		SetBody(ns).
		SetContext(ctx).
		SetHeaders(httputil.RequestXHeadersMap(ctx)).
		SetPathParams(map[string]string{
			"namespace": ns.ID,
			"project":   projectID,
		}).
		Put("/projects/{project}/namespaces/{namespace}")
	if err != nil {
		return errors.ErrInternal().Log(err, k.log)
	}
	if resp.Error() != nil {
		return resp.Error().(*cherry.Err)
	}
	return nil
}

func (k *KubeAPIHTTPClient) DeleteNamespace(ctx context.Context, projectID string, ns model.Namespace) error {
	k.log.WithFields(logrus.Fields{
		"name":       ns.ID,
		"project_id": projectID,
	}).Debugf("delete namespace")

	resp, err := k.client.R().
		SetContext(ctx).
		SetHeaders(httputil.RequestXHeadersMap(ctx)).
		SetPathParams(map[string]string{
			"namespace": ns.ID,
			"project":   projectID,
		}).
		Delete("/projects/{project}/namespaces/{namespace}")
	if err != nil {
		return errors.ErrInternal().Log(err, k.log)
	}
	if resp.Error() != nil {
		return resp.Error().(*cherry.Err)
	}
	return nil
}

func (k *KubeAPIHTTPClient) GetNamespace(ctx context.Context, projectID string, name string) (ret model.Namespace, err error) {
	k.log.WithFields(logrus.Fields{
		"name":       name,
		"project_id": projectID,
	}).Debugf("get namespace")

	resp, err := k.client.R().
		SetResult(&ret).
		SetContext(ctx).
		SetHeaders(httputil.RequestXHeadersMap(ctx)).
		SetPathParams(map[string]string{
			"namespace": name,
			"project":   projectID,
		}).
		Get("/projects/{project}/namespaces/{namespace}")
	if err != nil {
		err = errors.ErrInternal().Log(err, k.log)
		return
	}
	if resp.Error() != nil {
		err = resp.Error().(*cherry.Err)
		return
	}
	return
}

func (k *KubeAPIHTTPClient) String() string {
	return fmt.Sprintf("kube-api http client: url=%s", k.client.HostURL)
}

type KubeAPIDummyClient struct {
	log *logrus.Entry
}

func NewKubeAPIDummyClient() *KubeAPIDummyClient {
	return &KubeAPIDummyClient{
		log: logrus.WithField("component", "kube_api_client"),
	}
}

func (k *KubeAPIDummyClient) CreateNamespace(ctx context.Context, projectID string, req model.Namespace) error {
	k.log.WithFields(logrus.Fields{
		"project_id": projectID,
		"cpu":        req.Resources.Hard.CPU,
		"memory":     req.Resources.Hard.Memory,
		"label":      req.Label,
		"name":       req.ID,
		"access":     req.Access,
	}).Debug("create namespace")

	return nil
}

func (k *KubeAPIDummyClient) SetNamespaceQuota(ctx context.Context, projectID string, ns model.Namespace) error {
	k.log.WithFields(logrus.Fields{
		"project_id": projectID,
		"cpu":        ns.Resources.Hard.CPU,
		"memory":     ns.Resources.Hard.Memory,
		"label":      ns.Label,
		"name":       ns.ID,
	}).Debug("set namespace quota")

	return nil
}

func (k *KubeAPIDummyClient) DeleteNamespace(ctx context.Context, projectID string, ns model.Namespace) error {
	k.log.WithFields(logrus.Fields{
		"name":       ns.ID,
		"project_id": projectID,
	}).Debugf("delete namespace")

	return nil
}

func (k *KubeAPIDummyClient) String() string {
	return "kube-api dummy client"
}

func (k *KubeAPIDummyClient) GetNamespace(ctx context.Context, projectID string, name string) (ret model.Namespace, err error) {
	k.log.WithFields(logrus.Fields{
		"name":       name,
		"project_id": projectID,
	}).Debugf("get namespace")

	return
}
