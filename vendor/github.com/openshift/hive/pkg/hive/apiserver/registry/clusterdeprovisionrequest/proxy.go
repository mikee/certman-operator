package clusterdeprovisionrequest

import (
	"context"

	hivev1 "github.com/openshift/hive/pkg/apis/hive/v1"
	hivev1client "github.com/openshift/hive/pkg/client/clientset-generated/clientset/typed/hive/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metainternal "k8s.io/apimachinery/pkg/apis/meta/internalversion"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	apirequest "k8s.io/apiserver/pkg/endpoints/request"
	"k8s.io/apiserver/pkg/registry/rest"
	"k8s.io/kubernetes/pkg/printers"
	printerstorage "k8s.io/kubernetes/pkg/printers/storage"

	hiveapi "github.com/openshift/hive/pkg/hive/apis/hive"
	"github.com/openshift/hive/pkg/hive/apiserver/registry"
	"github.com/openshift/hive/pkg/hive/apiserver/registry/util"
	printersinternal "github.com/openshift/hive/pkg/printers/internalversion"
)

type REST struct {
	client hivev1client.HiveV1Interface
	rest.TableConvertor
}

var _ rest.Lister = &REST{}
var _ rest.Getter = &REST{}
var _ rest.CreaterUpdater = &REST{}
var _ rest.GracefulDeleter = &REST{}
var _ rest.Scoper = &REST{}

func NewREST(client hivev1client.HiveV1Interface) registry.NoWatchStorage {
	return registry.WrapNoWatchStorageError(&REST{
		client:         client,
		TableConvertor: printerstorage.TableConvertor{TablePrinter: printers.NewTablePrinter().With(printersinternal.AddHandlers)},
	})
}

func (s *REST) New() runtime.Object {
	return &hiveapi.ClusterDeprovisionRequest{}
}

func (s *REST) NewList() runtime.Object {
	return &hiveapi.ClusterDeprovisionRequestList{}
}

func (s *REST) NamespaceScoped() bool {
	return true
}

func (s *REST) List(ctx context.Context, options *metainternal.ListOptions) (runtime.Object, error) {
	client, err := s.getClient(ctx)
	if err != nil {
		return nil, err
	}

	optv1 := metav1.ListOptions{}
	if err := metainternal.Convert_internalversion_ListOptions_To_v1_ListOptions(options, &optv1, nil); err != nil {
		return nil, err
	}

	clusterDeprovisions, err := client.List(optv1)
	if err != nil {
		return nil, err
	}

	ret := &hiveapi.ClusterDeprovisionRequestList{
		ListMeta: clusterDeprovisions.ListMeta,
		Items:    make([]hiveapi.ClusterDeprovisionRequest, len(clusterDeprovisions.Items)),
	}
	for i, curr := range clusterDeprovisions.Items {
		if err := util.ClusterDeprovisionRequestFromHiveV1(&curr, &ret.Items[i]); err != nil {
			return nil, err
		}
	}
	return ret, nil
}

func (s *REST) Get(ctx context.Context, name string, options *metav1.GetOptions) (runtime.Object, error) {
	client, err := s.getClient(ctx)
	if err != nil {
		return nil, err
	}

	ret, err := client.Get(name, *options)
	if err != nil {
		return nil, err
	}

	clusterDeprovisionRequest := &hiveapi.ClusterDeprovisionRequest{}
	if err := util.ClusterDeprovisionRequestFromHiveV1(ret, clusterDeprovisionRequest); err != nil {
		return nil, err
	}
	return clusterDeprovisionRequest, nil
}

func (s *REST) Delete(ctx context.Context, name string, options *metav1.DeleteOptions) (runtime.Object, bool, error) {
	client, err := s.getClient(ctx)
	if err != nil {
		return nil, false, err
	}

	if err := client.Delete(name, options); err != nil {
		return nil, false, err
	}

	return &metav1.Status{Status: metav1.StatusSuccess}, true, nil
}

func (s *REST) Create(ctx context.Context, obj runtime.Object, _ rest.ValidateObjectFunc, options *metav1.CreateOptions) (runtime.Object, error) {
	client, err := s.getClient(ctx)
	if err != nil {
		return nil, err
	}

	convertedObj := &hivev1.ClusterDeprovision{}
	if err := util.ClusterDeprovisionRequestToHiveV1(obj.(*hiveapi.ClusterDeprovisionRequest), convertedObj); err != nil {
		return nil, err
	}

	ret, err := client.Create(convertedObj)
	if err != nil {
		return nil, err
	}

	clusterDeprovisionRequest := &hiveapi.ClusterDeprovisionRequest{}
	if err := util.ClusterDeprovisionRequestFromHiveV1(ret, clusterDeprovisionRequest); err != nil {
		return nil, err
	}
	return clusterDeprovisionRequest, nil
}

func (s *REST) Update(ctx context.Context, name string, objInfo rest.UpdatedObjectInfo, _ rest.ValidateObjectFunc, _ rest.ValidateObjectUpdateFunc, forceAllowCreate bool, options *metav1.UpdateOptions) (runtime.Object, bool, error) {
	client, err := s.getClient(ctx)
	if err != nil {
		return nil, false, err
	}

	clusterDeprovisionRequest, err := client.Get(name, metav1.GetOptions{})
	if err != nil {
		return nil, false, err
	}

	old := &hiveapi.ClusterDeprovisionRequest{}
	if err := util.ClusterDeprovisionRequestFromHiveV1(clusterDeprovisionRequest, old); err != nil {
		return nil, false, err
	}

	obj, err := objInfo.UpdatedObject(ctx, old)
	if err != nil {
		return nil, false, err
	}

	if err := util.ClusterDeprovisionRequestToHiveV1(obj.(*hiveapi.ClusterDeprovisionRequest), clusterDeprovisionRequest); err != nil {
		return nil, false, err
	}

	ret, err := client.Update(clusterDeprovisionRequest)
	if err != nil {
		return nil, false, err
	}

	new := &hiveapi.ClusterDeprovisionRequest{}
	if err := util.ClusterDeprovisionRequestFromHiveV1(ret, new); err != nil {
		return nil, false, err
	}
	return new, false, err
}

func (s *REST) getClient(ctx context.Context) (hivev1client.ClusterDeprovisionInterface, error) {
	namespace, ok := apirequest.NamespaceFrom(ctx)
	if !ok {
		return nil, apierrors.NewBadRequest("namespace parameter required")
	}
	return s.client.ClusterDeprovisions(namespace), nil
}
