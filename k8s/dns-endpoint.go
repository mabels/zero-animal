package k8s

import (
	"context"
	"fmt"
	"strings"

	"github.com/mabels/zero-animal/config"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"

	"k8s.io/client-go/rest"

	"sigs.k8s.io/external-dns/endpoint"

	"github.com/mabels/ipaddress/go/ipaddress"
)

const ApiVersion = "externaldns.k8s.io/v1alpha1"
const Kind = "DNSEndpoint"

// type crdSource struct {
// 	crdClient        rest.Interface
// 	namespace        string
// 	crdResource      string
// 	codec            runtime.ParameterCodec
// 	annotationFilter string
// 	labelSelector    labels.Selector
// }

func addKnownTypes(scheme *runtime.Scheme, groupVersion schema.GroupVersion) error {
	scheme.AddKnownTypes(groupVersion,
		&endpoint.DNSEndpoint{},
		&endpoint.DNSEndpointList{},
	)
	metav1.AddToGroupVersion(scheme, groupVersion)
	return nil
}

// func (cs *crdSource) filterByAnnotations(dnsendpoints *endpoint.DNSEndpointList) (*endpoint.DNSEndpointList, error) {
// 	labelSelector, err := metav1.ParseToLabelSelector(cs.annotationFilter)
// 	if err != nil {
// 		return nil, err
// 	}
// 	selector, err := metav1.LabelSelectorAsSelector(labelSelector)
// 	if err != nil {
// 		return nil, err
// 	}

// 	// empty filter returns original list
// 	if selector.Empty() {
// 		return dnsendpoints, nil
// 	}

// 	filteredList := endpoint.DNSEndpointList{}

// 	for _, dnsendpoint := range dnsendpoints.Items {
// 		// convert the dnsendpoint' annotations to an equivalent label selector
// 		annotations := labels.Set(dnsendpoint.Annotations)

// 		// include dnsendpoint if its annotations match the selector
// 		if selector.Matches(annotations) {
// 			filteredList.Items = append(filteredList.Items, dnsendpoint)
// 		}
// 	}

// 	return &filteredList, nil
// }

type ATypeEndPoints struct {
	Name   string
	ATypes []string
}

func toEndpointTargets(a map[string]string) endpoint.Targets {
	targets := endpoint.Targets{}
	for ip, _ := range a {
		targets = append(targets, ip)
	}
	return targets
}

type DNSEndpointApi struct {
	restClient *rest.RESTClient
	cfg        config.K8sCfg
	scheme     *runtime.Scheme
}

func MakeDNSEndpointApi(cfg config.K8sCfg, k8sConfig *rest.Config) (*DNSEndpointApi, error) {
	groupVersion, err := schema.ParseGroupVersion(ApiVersion)
	if err != nil {
		return nil, err
	}

	scheme := runtime.NewScheme()
	addKnownTypes(scheme, groupVersion)

	crdConfig := *k8sConfig
	crdConfig.ContentConfig.GroupVersion = &groupVersion
	crdConfig.APIPath = "/apis"
	crdConfig.NegotiatedSerializer = serializer.WithoutConversionCodecFactory{CodecFactory: serializer.NewCodecFactory(scheme)}
	crdConfig.UserAgent = rest.DefaultKubernetesUserAgent()

	crdClient, err := rest.UnversionedRESTClientFor(&crdConfig)
	if err != nil {
		return nil, err
	}
	return &DNSEndpointApi{
		restClient: crdClient,
		cfg:        cfg,
		scheme:     scheme,
	}, nil
}

func (dn *DNSEndpointApi) List(opts *metav1.ListOptions) (result *endpoint.DNSEndpointList, err error) {
	result = &endpoint.DNSEndpointList{}
	err = dn.restClient.Get().
		Namespace(dn.cfg.Namespace).
		Resource(strings.ToLower(Kind) + "s").
		// VersionedParams(opts, cs.codec).
		Do(context.TODO()).
		Into(result)
	return
}

func (dn *DNSEndpointApi) GetEndpoint(name string) (*endpoint.DNSEndpoint, error) {
	// var result runtime.Object
	result := &endpoint.DNSEndpoint{}
	err := dn.restClient.Get().
		Namespace(dn.cfg.Namespace).
		Resource(strings.ToLower(Kind) + "s").
		Name(name).
		Do(context.TODO()).
		Into(result)
	// log.Print(dn.cfg.Namespace, Kind, result)
	return result, err
}

func (dn *DNSEndpointApi) DeleteEndpoint(name string) (runtime.Object, error) {
	var result runtime.Object
	err := dn.restClient.Delete().
		Namespace(dn.cfg.Namespace).
		Resource(strings.ToLower(Kind) + "s").
		Name(name).
		Do(context.TODO()).
		Into(result)
	// log.Print(dn.cfg.Namespace, Kind, result)
	return result, err
}

func (dn *DNSEndpointApi) buildEndPoints(atyp ATypeEndPoints) ([]*endpoint.Endpoint, error) {
	if len(atyp.ATypes) == 0 {
		return nil, nil
	}
	eps := []*endpoint.Endpoint{}
	aaaas := map[string]string{}
	as := map[string]string{}
	for _, ip := range atyp.ATypes {
		res := ipaddress.Parse(ip)
		if res.IsErr() {
			return nil, fmt.Errorf(*res.UnwrapErr())
		}
		if res.Unwrap().Is_ipv4() {
			as[ip] = ip
		}
		if res.Unwrap().Is_ipv6() {
			aaaas[ip] = ip
		}
	}
	if len(as) > 0 {
		eps = append(eps, &endpoint.Endpoint{
			DNSName:    atyp.Name,
			Targets:    toEndpointTargets(as),
			RecordType: "A",
			RecordTTL:  endpoint.TTL(dn.cfg.TTL),
		})
	}
	if len(aaaas) > 0 {
		eps = append(eps, &endpoint.Endpoint{
			DNSName:    atyp.Name,
			Targets:    toEndpointTargets(aaaas),
			RecordType: "AAAA",
			RecordTTL:  endpoint.TTL(dn.cfg.TTL),
		})
	}
	return eps, nil
}

func (dn *DNSEndpointApi) PatchEndpoint(atyp ATypeEndPoints) (runtime.Object, error) {
	eps, err := dn.buildEndPoints(atyp)
	if err != nil {
		return nil, err
	}

	name := strings.ReplaceAll(atyp.Name, ".", "-")
	ep, err := dn.GetEndpoint(name)
	if err != nil {
		return ep, err
	}
	ep.Spec.Endpoints = eps
	var result runtime.Object
	err = dn.restClient.Put().
		Namespace(dn.cfg.Namespace).
		Resource(strings.ToLower(Kind) + "s").
		Name(name).
		// SubResource("status").
		Body(ep).
		Do(context.TODO()).
		Into(result)
	// log.Print(dn.cfg.Namespace, Kind, result)
	return result, err
}

func (dn *DNSEndpointApi) PostEndpoint(atyp ATypeEndPoints) (runtime.Object, error) {
	eps, err := dn.buildEndPoints(atyp)
	if err != nil {
		return nil, err
	}
	newEP := &endpoint.DNSEndpoint{
		ObjectMeta: metav1.ObjectMeta{
			Name:      strings.ReplaceAll(atyp.Name, ".", "-"),
			Namespace: dn.cfg.Namespace,
			Labels:    dn.cfg.Labels,
		},
		Spec: endpoint.DNSEndpointSpec{
			Endpoints: eps,
		},
	}
	var result runtime.Object
	err = dn.restClient.Post().
		Namespace(dn.cfg.Namespace).
		Resource(strings.ToLower(Kind) + "s").
		Name(newEP.ObjectMeta.Name).
		// SubResource("status").
		Body(newEP).
		Do(context.TODO()).
		Into(result)
	// log.Print(dn.cfg.Namespace, Kind, result)
	return result, err
}

func (dn *DNSEndpointApi) ReadEndPoints() (*endpoint.DNSEndpointList, error) {

	labelSelector, err := labels.Parse(dn.cfg.LabelFilter)
	if err != nil {
		return nil, err
	}

	// crdSource := &crdSource{
	// 	crdResource:      strings.ToLower(Kind) + "s",
	// 	namespace:        dn.cfg.Namespace,
	// 	annotationFilter: dn.cfg.AnnotationFilter,
	// 	labelSelector:    labelSelector,
	// 	crdClient:        dn.restClient,
	// 	codec:            runtime.NewParameterCodec(dn.scheme),
	// }

	eps, err := dn.List(&metav1.ListOptions{LabelSelector: labelSelector.String()})
	return eps, err

}
