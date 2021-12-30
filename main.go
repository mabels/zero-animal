package main

import (
	"flag"
	"os"
	"sort"
	"strings"

	"github.com/mabels/zero-animal/config"
	"github.com/mabels/zero-animal/k8s"
	"github.com/mabels/zero-animal/zerotier"
	"sigs.k8s.io/external-dns/endpoint"

	"k8s.io/klog/v2"
)

type DNSEndpointRef struct {
	DNSEndpoint *endpoint.DNSEndpoint
	Endpoint    *endpoint.Endpoint
}

type BothSides struct {
	Key       string
	ZeroTier  *zerotier.Member
	Endpoints []DNSEndpointRef
}

type EndpointTargetRef struct {
	EndPoint DNSEndpointRef
	Target   int
}

type BothIps struct {
	BothSides         *BothSides
	Ip                string
	ZeroTier          *zerotier.Member
	EndpointTargetRef *EndpointTargetRef
}

func toSortedArray(dup map[string]string) []string {
	ret := make([]string, 0, len(dup))
	for ip, _ := range dup {
		ret = append(ret, ip)
	}
	sort.Strings(ret)
	return ret
}

func (bs *BothSides) zeroTierIps() []string {
	if bs.ZeroTier == nil {
		return []string{}
	}
	ipAs, ok := bs.ZeroTier.Config["ipAssignments"]
	if !ok {
		return []string{}
	}
	dup := map[string]string{}
	for _, aip := range ipAs.([]interface{}) {
		ip := aip.(string)
		dup[ip] = ip
	}
	return toSortedArray(dup)
}

func (bs *BothSides) endpointsIps() []string {
	dup := map[string]string{}
	for _, ep := range bs.Endpoints {
		for _, tip := range ep.Endpoint.Targets {
			dup[tip] = tip
		}
	}
	return toSortedArray(dup)
}

func (bs *BothSides) equal() bool {
	zips := bs.zeroTierIps()
	eips := bs.endpointsIps()
	if len(zips) != len(eips) {
		return false
	}
	for i, _ := range zips {
		if zips[i] != eips[i] {
			return false
		}
	}
	return true
}

func main() {

	klog.InitFlags(nil)
	defer klog.Flush() // flushes all pending log I/O

	ztconfig := config.MakeConfig(os.Args)

	flag.Parse()

	if len(ztconfig.ZeroTier.Networks) == 0 {
		klog.Fatalln("We need atleast one network set")
	}

	zt := zerotier.MakeZeroTier(ztconfig.ZeroTier.ZeroTier)
	members, err := zt.NetworkMember(ztconfig.ZeroTier.Networks[0])
	if err != nil {
		klog.Fatal(err)
		return
	}

	bothSides := map[string]*BothSides{}
	defaultDomain := strings.Trim(strings.TrimSpace(ztconfig.ZeroTier.DefaultDomain), ".")
	for _, member := range members {
		if len(member.Name) > 0 {
			key := strings.TrimRight(strings.TrimSpace(member.Name), ".")
			if len(defaultDomain) > 0 && !strings.Contains(key, ".") {
				key = strings.Join([]string{key, defaultDomain}, ".")
			}
			key = strings.TrimRight(key, ".")
			refMember := member
			bothSides[key] = &BothSides{
				Key:       key,
				ZeroTier:  &refMember,
				Endpoints: []DNSEndpointRef{},
			}
		}
	}

	k8sConfig, err := k8s.GetConfig(ztconfig.K8s)
	if err != nil {
		klog.Fatal(err)
		return
	}

	dn, err := k8s.MakeDNSEndpointApi(ztconfig.K8s, k8sConfig)
	if err != nil {
		klog.Fatal(err)
		return
	}

	eps, err := dn.ReadEndPoints()
	if err != nil {
		klog.Fatal(err)
		return
	}

	for _, dep := range eps.Items {
		for _, ep := range dep.Spec.Endpoints {
			key := strings.TrimRight(strings.TrimSpace(ep.DNSName), ".")
			val, ok := bothSides[key]
			if !ok {
				val = &BothSides{
					Key:       key,
					Endpoints: []DNSEndpointRef{},
				}
				bothSides[key] = val
			}
			val.Endpoints = append(val.Endpoints, DNSEndpointRef{
				DNSEndpoint: &dep,
				Endpoint:    ep,
			})
		}
	}
	for _, both := range bothSides {

		if both.ZeroTier == nil && len(both.Endpoints) > 0 {
			for _, ep := range both.Endpoints {
				klog.Info("K8S-Remove:", both.Key, ep.DNSEndpoint.Name)
				_, err := dn.DeleteEndpoint(ep.DNSEndpoint.Name)
				if err != nil {
					klog.Fatal(err)
				}
			}
		}
		if both.ZeroTier != nil && len(both.Endpoints) > 0 {
			if both.equal() {
			} else {
				klog.Info("K8S-Update:", both.Key, both.zeroTierIps())
				result, err := dn.PatchEndpoint(k8s.ATypeEndPoints{
					Name:   both.Key,
					ATypes: both.zeroTierIps(),
				})
				if err != nil {
					klog.Fatal(err, result)
				}
			}
		}
		if both.ZeroTier != nil && len(both.Endpoints) == 0 {
			klog.Info("K8S-New:", both.Key, both.zeroTierIps())
			_, err := dn.PostEndpoint(k8s.ATypeEndPoints{
				Name:   both.Key,
				ATypes: both.zeroTierIps(),
			})
			if err != nil {
				klog.Fatal(err)
			}
		}
	}
}
