package config

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/mabels/zero-animal/zerotier"
)

type ZeroTierCfg struct {
	zerotier.ZeroTier
	DefaultDomain string
	Networks      arrayFlags
}

type K8sCfg struct {
	Namespace        string
	AnnotationFilter string
	LabelFilter      string
	TTL              int64
	Labels           mapFlags
}

type Config struct {
	Version  bool
	ZeroTier ZeroTierCfg
	K8s      K8sCfg
}

type arrayFlags []string

func (i *arrayFlags) String() string {
	return strings.Join(*i, ",")
}

func (i *arrayFlags) Set(value string) error {
	*i = append(*i, value)
	return nil
}

type mapFlags map[string]string

func (i *mapFlags) String() string {
	out := []string{}
	for k, v := range *i {
		out = append(out, fmt.Sprintf("%s:%s", k, v))
	}
	return strings.Join(out, ",")
}

func (i *mapFlags) Set(value string) error {
	kv := strings.Split(value, "=")
	(*i)[kv[0]] = kv[1]
	return nil
}

func MakeConfig(args []string) *Config {
	cfg := Config{
		Version: false,
		ZeroTier: ZeroTierCfg{
			ZeroTier: zerotier.ZeroTier{
				Host:   "my.zerotier.com",
				Bearer: "",
			},
			DefaultDomain: ".adviser.com.",
			Networks:      []string{},
		},
		K8s: K8sCfg{
			Namespace:        "default",
			TTL:              3600,
			AnnotationFilter: "",
			LabelFilter:      "zero-animal=created",
			Labels: map[string]string{
				"zero-animal": "created",
			},
		},
	}
	bearer, ok := os.LookupEnv("ZERO_TIER_BEARER")
	if !ok {
		bearer = ""
	}
	netStr, ok := os.LookupEnv("ZERO_TIER_NETWORKS")
	if ok {
		for _, net := range strings.Split(netStr, ",") {
			cfg.ZeroTier.Networks = append(cfg.ZeroTier.Networks, net)
		}
	}
	flag.BoolVar(&cfg.Version, "version", false, "display version")

	flag.StringVar(&cfg.ZeroTier.Host, "zeroTierHost", cfg.ZeroTier.Host, "ZeroTier API Host")
	flag.StringVar(&cfg.ZeroTier.Bearer, "zeroTierBearer", bearer, "ZeroTier API Bearer[ZERO_TIER_BEARER]")
	flag.StringVar(&cfg.ZeroTier.DefaultDomain, "zeroTierDefaultDomain", cfg.ZeroTier.DefaultDomain, "ZeroTier Add DefaultDomain")
	flag.Var(&cfg.ZeroTier.Networks, "zeroNetworks", "ZeroTier Networks [ZERO_TIER_NETWORKS]")

	flag.StringVar(&cfg.K8s.Namespace, "k8sNamespace", cfg.K8s.Namespace, "Kubernetes Namespace")
	flag.Int64Var(&cfg.K8s.TTL, "k8sTTL", cfg.K8s.TTL, "DNS TTL")
	flag.StringVar(&cfg.K8s.AnnotationFilter, "k8sAnnotationFilter", cfg.K8s.AnnotationFilter, "AnnotationFilter")
	flag.StringVar(&cfg.K8s.LabelFilter, "k8sLabelFilter", cfg.K8s.LabelFilter, "LabelFilter")
	flag.Var(&cfg.K8s.Labels, "k8sLabels", "K8s Labels")
	return &cfg
}
