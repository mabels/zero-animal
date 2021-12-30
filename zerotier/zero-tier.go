package zerotier

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type ZeroTier struct {
	Bearer string
	Host   string
}

func MakeZeroTier(cfg ZeroTier) ZeroTier {
	ret := cfg
	return ret
}

type Member struct {
	Id                  string                 `json:"id"`
	Hidden              bool                   `json:"hidden"`
	Clock               int64                  `json:"clock"`
	NetworkId           string                 `json:"networkId"`
	NodeId              string                 `json:"nodeId"`
	ControllerId        string                 `json:"controllerId"`
	Name                string                 `json:"name"`
	Description         string                 `json:"description"`
	Config              map[string]interface{} `json:"config"`
	LastOnline          int64                  `json:"lastOnline"`
	PhysicalAddress     string                 `json:"physicalAddress"`
	ClientVersion       string                 `json:"clientVersion"`
	ProtocolVersion     int                    `json:"protocolVersion"`
	SupportsRulesEngine bool                   `json:"supportsRulesEngine"`
}

func (zt *ZeroTier) NetworkMember(net string) ([]Member, error) {
	rurl := fmt.Sprintf("https://%s/api/v1/network/%s/member", zt.Host, net)

	req, err := http.NewRequest("GET", rurl, nil)
	if err != nil {
		return nil, err
	}
	if len(zt.Bearer) > 0 {
		req.Header.Add("Authorization", fmt.Sprintf("bearer %s", zt.Bearer))
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var ret []Member
	err = json.NewDecoder(resp.Body).Decode(&ret)
	if err != nil {
		return nil, err
	}
	return ret, nil
}
