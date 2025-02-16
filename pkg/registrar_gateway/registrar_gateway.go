package registrargw

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	subTypes "github.com/centrifuge/go-substrate-rpc-client/v4/types"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	substrate "github.com/threefoldtech/tfchain/clients/tfchain-client-go"
	"github.com/threefoldtech/zbus"
	zos4Pkg "github.com/threefoldtech/zos4/pkg"
	"github.com/threefoldtech/zos4/pkg/stubs"
	"github.com/threefoldtech/zos4/pkg/types"
	"github.com/threefoldtech/zosbase/pkg"
	"github.com/threefoldtech/zosbase/pkg/environment"
)

const AuthHeader = "X-Auth"

type registrarGateway struct {
	mu         sync.Mutex
	baseURL    string
	httpClient *http.Client
	privKey    ed25519.PrivateKey
	nodeID     uint64
	twinID     uint64
}

var ErrorRecordNotFound = errors.New("could not find the reqested record")

func NewRegistrarGateway(cl zbus.Client) (zos4Pkg.RegistrarGateway, error) {
	client := http.DefaultClient
	env := environment.MustGet()

	identity := stubs.NewIdentityManagerStub(cl)
	sk := ed25519.PrivateKey(identity.PrivateKey(context.TODO()))

	gw := &registrarGateway{
		httpClient: client,
		baseURL:    env.RegistrarURL,
		mu:         sync.Mutex{},
		privKey:    sk,
	}

	pk := sk.Public().(ed25519.PublicKey)
	twin, err := gw.GetTwinByPubKey(pk)
	if err == nil {
		gw.twinID = twin

		node, err := gw.GetNodeByTwinID(twin)
		if err == nil {
			gw.nodeID = node
		}
	}

	return gw, nil
}

func (r *registrarGateway) GetZosVersion() (string, error) {
	url := fmt.Sprintf("%s/v1/zos/version", r.baseURL)
	log.Debug().Str("url", url).Msg("requesting zos version")

	return r.getZosVersion(url)
}

func (r *registrarGateway) CreateNode(node types.UpdateNodeRequest) (uint64, error) {
	url := fmt.Sprintf("%s/v1/nodes", r.baseURL)
	log.Debug().
		Str("url", url).
		Uint32("farm id", uint32(node.FarmID)).
		Msg("creating node")

	r.mu.Lock()
	defer r.mu.Unlock()

	return r.createNode(url, node)
}

func (r *registrarGateway) CreateTwin(relay string, pk []byte) (types.Account, error) {
	url := fmt.Sprintf("%s/v1/accounts", r.baseURL)
	log.Debug().Str("url", url).Str("relay", relay).Str("pk", hex.EncodeToString(pk)).Msg("creating account")

	r.mu.Lock()
	defer r.mu.Unlock()

	twin, err := r.createTwin(url, []string{relay}, pk)
	if err != nil {
		return types.Account{}, err
	}

	return twin, nil
}

func (r *registrarGateway) EnsureAccount(pk []byte) (twin types.Account, err error) {
	url := fmt.Sprintf("%s/v1/accounts", r.baseURL)

	publicKeyBase64 := base64.StdEncoding.EncodeToString(pk)
	log.Debug().Str("url", url).Bytes("public_key", []byte(publicKeyBase64)).Msg("ensure account")

	r.mu.Lock()
	defer r.mu.Unlock()

	return r.ensureAccount(url, pk)
}

func (r *registrarGateway) GetContract(id uint64) (result substrate.Contract, serr pkg.SubstrateError) {
	// log.Trace().Str("method", "GetContract").Uint64("id", id).Msg("method called")
	// contract, err := g.sub.GetContract(id)
	//
	// serr = buildSubstrateError(err)
	// if err != nil {
	// 	return
	// }
	return
}

func (r *registrarGateway) GetContractIDByNameRegistration(name string) (result uint64, serr pkg.SubstrateError) {
	// log.Trace().Str("method", "GetContractIDByNameRegistration").Str("name", name).Msg("method called")
	// contractID, err := g.sub.GetContractIDByNameRegistration(name)
	//
	// serr = buildSubstrateError(err)
	return
}

func (r *registrarGateway) GetFarm(id uint64) (farm types.Farm, err error) {
	url := fmt.Sprintf("%s/v1/farms/%d", r.baseURL, id)
	log.Debug().Str("url", url).Uint64("id", id).Msg("get farm")

	return r.getFarm(url)
}

func (r *registrarGateway) GetNode(id uint64) (node types.Node, err error) {
	url := fmt.Sprintf("%s/v1/nodes/%d", r.baseURL, id)
	log.Trace().Str("url", url).Uint64("id", id).Msg("get node")

	return r.getNode(url)
}

func (r *registrarGateway) GetNodeByTwinID(twin uint64) (result uint64, err error) {
	url := fmt.Sprintf("%s/v1/nodes", r.baseURL)
	log.Trace().Str("url", url).Uint64("twin", twin).Msg("get node by twin_id")

	return r.getNodeByTwinID(url, twin)
}

func (r *registrarGateway) GetNodeContracts(node uint32) ([]subTypes.U64, error) {
	// log.Trace().Str("method", "GetNodeContracts").Uint32("node", node).Msg("method called")
	// return g.sub.GetNodeContracts(node)
	return []subTypes.U64{}, nil
}

func (r *registrarGateway) GetNodeRentContract(node uint32) (result uint64, serr pkg.SubstrateError) {
	// log.Trace().Str("method", "GetNodeRentContract").Uint32("node", node).Msg("method called")
	// contractID, err := g.sub.GetNodeRentContract(node)
	//
	// serr = buildSubstrateError(err)
	return
}

func (r *registrarGateway) GetNodes(farmID uint32) (nodeIDs []uint32, err error) {
	log.Trace().Str("method", "GetNodes").Uint32("farm id", farmID).Msg("method called")

	url := fmt.Sprintf("%s/v1/nodes", r.baseURL)
	return r.getNodesInFarm(url, farmID)
}

func (r *registrarGateway) GetPowerTarget() (power substrate.NodePower, err error) {
	// log.Trace().Str("method", "GetPowerTarget").Uint32("node id", uint32(g.nodeID)).Msg("method called")
	// return g.sub.GetPowerTarget(uint32(g.nodeID))
	return
}

func (r *registrarGateway) GetTwin(id uint64) (result types.Account, err error) {
	url := fmt.Sprintf("%s/v1/accounts/", r.baseURL)
	log.Trace().Str("url", "url").Uint64("id", id).Msg("get account")

	return r.getTwin(url, id)
}

func (r *registrarGateway) GetTwinByPubKey(pk []byte) (result uint64, err error) {
	url := fmt.Sprintf("%s/v1/accounts/", r.baseURL)
	log.Trace().Str("method", "GetTwinByPubKey").Str("pk", hex.EncodeToString(pk)).Msg("method called")

	return r.getTwinByPubKey(url, pk)
}

func (r *registrarGateway) Report(consumptions []substrate.NruConsumption) (subTypes.Hash, error) {
	// contractIDs := make([]uint64, 0, len(consumptions))
	// for _, v := range consumptions {
	// 	contractIDs = append(contractIDs, uint64(v.ContractID))
	// }
	//
	// log.Debug().Str("method", "Report").Uints64("contract ids", contractIDs).Msg("method called")
	// r.mu.Lock()
	// defer r.mu.Unlock()
	//
	// url := fmt.Sprintf("%s/v1/nodes/%d/consumption", r.baseURL, r.nodeID)
	//
	// var body bytes.Buffer
	// _, err := r.httpClient.Post(url, "application/json", &body)
	// if err != nil {
	// 	return subTypes.Hash{}, err
	// }
	//
	// // I need to know what is hash to be able to respond with it
	// return r.sub.Report(r.identity, consumptions)
	return subTypes.Hash{}, nil
}

func (r *registrarGateway) SetContractConsumption(resources ...substrate.ContractResources) error {
	// contractIDs := make([]uint64, 0, len(resources))
	// for _, v := range resources {
	// 	contractIDs = append(contractIDs, uint64(v.ContractID))
	// }
	// log.Debug().Str("method", "SetContractConsumption").Uints64("contract ids", contractIDs).Msg("method called")
	// g.mu.Lock()
	// defer g.mu.Unlock()
	// return g.sub.SetContractConsumption(g.identity, resources...)
	return nil
}

func (r *registrarGateway) SetNodePowerState(up bool) (hash subTypes.Hash, err error) {
	// log.Debug().Str("method", "SetNodePowerState").Bool("up", up).Msg("method called")
	// g.mu.Lock()
	// defer g.mu.Unlock()
	// return g.sub.SetNodePowerState(g.identity, up)
	return subTypes.Hash{}, nil
}

func (r *registrarGateway) UpdateNode(node types.UpdateNodeRequest) (uint64, error) {
	url := fmt.Sprintf("%s/v1/nodes/%d", r.baseURL, r.nodeID)
	log.Debug().Str("url", url).Msg("upate node")
	r.mu.Lock()
	defer r.mu.Unlock()

	return r.updateNode(url, node)
}

func (r *registrarGateway) UpdateNodeUptimeV2(uptime uint64, timestampHint uint64) (err error) {
	url := fmt.Sprintf("%s/v1/nodes/%d/uptime", r.baseURL, r.nodeID)
	log.Debug().
		Str("url", url).
		Uint64("uptime", uptime).
		Uint64("timestamp hint", timestampHint).
		Msg("sending uptime report")
	r.mu.Lock()
	defer r.mu.Unlock()

	return r.updateNodeUptimeV2(r.twinID, url, uptime)
}

func (r *registrarGateway) GetTime() (time.Time, error) {
	// log.Trace().Str("method", "Time").Msg("method called")
	//
	// return g.sub.Time()
	return time.Now(), nil
}

func buildSubstrateError(err error) (serr pkg.SubstrateError) {
	if err == nil {
		return
	}

	serr.Err = err
	serr.Code = pkg.CodeGenericError

	if errors.Is(err, substrate.ErrNotFound) {
		serr.Code = pkg.CodeNotFound
	} else if errors.Is(err, substrate.ErrBurnTransactionNotFound) {
		serr.Code = pkg.CodeBurnTransactionNotFound
	} else if errors.Is(err, substrate.ErrRefundTransactionNotFound) {
		serr.Code = pkg.CodeRefundTransactionNotFound
	} else if errors.Is(err, substrate.ErrFailedToDecode) {
		serr.Code = pkg.CodeFailedToDecode
	} else if errors.Is(err, substrate.ErrInvalidVersion) {
		serr.Code = pkg.CodeInvalidVersion
	} else if errors.Is(err, substrate.ErrUnknownVersion) {
		serr.Code = pkg.CodeUnknownVersion
	} else if errors.Is(err, substrate.ErrIsUsurped) {
		serr.Code = pkg.CodeIsUsurped
	} else if errors.Is(err, substrate.ErrAccountNotFound) {
		serr.Code = pkg.CodeAccountNotFound
	} else if errors.Is(err, substrate.ErrDepositFeeNotFound) {
		serr.Code = pkg.CodeDepositFeeNotFound
	} else if errors.Is(err, substrate.ErrMintTransactionNotFound) {
		serr.Code = pkg.CodeMintTransactionNotFound
	}
	return
}

func (r *registrarGateway) createTwin(url string, relayURL []string, pk []byte) (twin types.Account, err error) {
	publicKeyBase64 := base64.StdEncoding.EncodeToString(pk)
	timestamp := time.Now().Unix()
	signature := getNodeSignature(publicKeyBase64, r.privKey, timestamp)

	account := types.AccountCreationRequest{
		PublicKey: publicKeyBase64,
		Signature: signature,
		Timestamp: timestamp,

		RMBEncKey: "",
		Relays:    relayURL,
	}

	var body bytes.Buffer
	err = json.NewEncoder(&body).Encode(account)
	if err != nil {
		return
	}

	resp, err := r.httpClient.Post(url, "application/json", &body)
	if err != nil {
		return
	}

	if resp.StatusCode != http.StatusCreated {
		err = parseRespError(resp.Body)
		return twin, errors.Wrapf(err, "failed to create twin with status %s", resp.Status)
	}
	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(&twin)

	r.twinID = twin.TwinID
	return twin, err
}

func (r *registrarGateway) ensureAccount(relay string, pk []byte) (twin types.Account, err error) {
	url := fmt.Sprintf("%s/v1/accounts/", r.baseURL)
	twinID, err := r.getTwinByPubKey(url, pk)
	if err != nil {
		if !errors.Is(err, ErrorRecordNotFound) {
			return types.Account{}, errors.Wrap(err, "failed to get twin by public key")
		}

		twin, err := r.createTwin(url, []string{relay}, pk)
		if err != nil {
			return types.Account{}, errors.Wrap(err, "failed to create twin")
		}
		return twin, nil
	}

	return r.getTwin(url, twinID)
}

func (r *registrarGateway) getTwin(url string, twinID uint64) (result types.Account, err error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return
	}

	q := req.URL.Query()
	q.Add("twin_id", fmt.Sprint(twinID))
	req.URL.RawQuery = q.Encode()

	resp, err := r.httpClient.Do(req)
	if err != nil {
		return
	}

	if resp == nil {
		return result, errors.New("no response received")
	}

	if resp.StatusCode == http.StatusNotFound {
		return result, ErrorRecordNotFound
	}

	if resp.StatusCode != http.StatusOK {
		err = parseRespError(resp.Body)
		return result, errors.Wrapf(err, "failed to get account by twin id with status code %s", resp.Status)
	}

	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(&result)
	return result, err
}

func (r *registrarGateway) getTwinByPubKey(url string, pk []byte) (result uint64, err error) {
	publicKeyBase64 := base64.StdEncoding.EncodeToString(pk)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return result, err
	}

	q := req.URL.Query()
	q.Add("public_key", publicKeyBase64)
	req.URL.RawQuery = q.Encode()

	resp, err := r.httpClient.Do(req)
	if err != nil {
		return result, err
	}

	if resp == nil {
		return result, errors.New("no response received")
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return result, ErrorRecordNotFound
	}

	if resp.StatusCode != http.StatusOK {
		err = parseRespError(resp.Body)
		return result, errors.Wrapf(err, "failed to get account by public_key with status code %s", resp.Status)
	}

	var account types.Account
	err = json.NewDecoder(resp.Body).Decode(&account)

	return account.TwinID, err
}

func (r *registrarGateway) getZosVersion(url string) (string, error) {
	resp, err := r.httpClient.Get(url)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != http.StatusOK {
		err = parseRespError(resp.Body)
		return "", errors.Wrapf(err, "failed to get zos version with status code %s", resp.Status)
	}

	defer resp.Body.Close()

	var versionString string
	err = json.NewDecoder(resp.Body).Decode(&versionString)
	if err != nil {
		return "", err
	}

	versionBytes, err := base64.StdEncoding.DecodeString(versionString)
	if err != nil {
		return "", err
	}

	correctedJSON := strings.ReplaceAll(string(versionBytes), "'", "\"")

	var version types.ZosVersion
	err = json.NewDecoder(strings.NewReader(correctedJSON)).Decode(&version)
	if err != nil {
		return "", err
	}

	return version.Version, err
}

func (r *registrarGateway) createNode(url string, node types.UpdateNodeRequest) (nodeID uint64, err error) {
	var body bytes.Buffer
	err = json.NewEncoder(&body).Encode(node)
	if err != nil {
		return
	}

	req, err := http.NewRequest("POST", url, &body)
	if err != nil {
		return
	}

	r.authRequest(req, r.twinID)

	resp, err := r.httpClient.Do(req)
	if err != nil {
		return 0, err
	}

	if resp == nil || resp.StatusCode != http.StatusCreated {
		err = parseRespError(resp.Body)
		return 0, errors.Wrapf(err, "failed to update node on the registrar with status code %s", resp.Status)
	}

	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(&nodeID)

	r.nodeID = nodeID
	return nodeID, err
}

func (r *registrarGateway) getFarm(url string) (farm types.Farm, err error) {
	resp, err := r.httpClient.Get(url)
	if err != nil {
		return farm, err
	}

	if resp.StatusCode == http.StatusNotFound {
		return farm, ErrorRecordNotFound
	}

	if resp.StatusCode != http.StatusOK {
		err = parseRespError(resp.Body)
		return farm, errors.Wrapf(err, "failed to get farm with status code %s", resp.Status)
	}
	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(&farm)
	if err != nil {
		return farm, err
	}

	return farm, err
}

func (r *registrarGateway) getNode(url string) (node types.Node, err error) {
	resp, err := r.httpClient.Get(url)
	if err != nil {
		return
	}

	if resp.StatusCode == http.StatusNotFound {
		return node, ErrorRecordNotFound
	}

	if resp.StatusCode != http.StatusOK {
		err = parseRespError(resp.Body)
		return node, errors.Wrapf(err, "failed to get node with status code %s", resp.Status)
	}

	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(&node)
	if err != nil {
		return
	}

	return node, err
}

func (r *registrarGateway) getNodeByTwinID(url string, twin uint64) (result uint64, err error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return
	}

	q := req.URL.Query()
	q.Add("twin_id", fmt.Sprint(twin))
	req.URL.RawQuery = q.Encode()

	resp, err := r.httpClient.Do(req)
	if err != nil {
		return
	}

	if resp == nil {
		return result, errors.New("no response received")
	}

	if resp.StatusCode == http.StatusNotFound {
		return result, ErrorRecordNotFound
	}

	if resp.StatusCode != http.StatusOK {
		err = parseRespError(resp.Body)
		return result, errors.Wrapf(err, "failed to get node by twin id with status code %s", resp.Status)
	}

	defer resp.Body.Close()

	var nodes []types.Node
	err = json.NewDecoder(resp.Body).Decode(&nodes)
	if err != nil {
		return
	}
	if len(nodes) == 0 {
		return result, ErrorRecordNotFound
	}

	return nodes[0].NodeID, nil
}

func (r *registrarGateway) getNodesInFarm(url string, farmID uint32) (nodeIDs []uint32, err error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return
	}

	q := req.URL.Query()
	q.Add("farm_id", fmt.Sprint(farmID))
	req.URL.RawQuery = q.Encode()

	resp, err := r.httpClient.Do(req)
	if err != nil {
		return
	}

	if resp == nil || resp.StatusCode != http.StatusOK {
		err = parseRespError(resp.Body)
		return nodeIDs, errors.Wrapf(err, "failed to get nodes in farm %d with status code %s", farmID, resp.Status)
	}

	defer resp.Body.Close()

	var nodes []types.Node
	err = json.NewDecoder(resp.Body).Decode(&nodes)
	if err != nil {
		return
	}

	for _, node := range nodes {
		nodeIDs = append(nodeIDs, uint32(node.NodeID))
	}

	return nodeIDs, nil
}

func (r *registrarGateway) updateNode(url string, node types.UpdateNodeRequest) (uint64, error) {
	var body bytes.Buffer
	err := json.NewEncoder(&body).Encode(node)
	if err != nil {
		return 0, err
	}

	req, err := http.NewRequest("PATCH", url, &body)
	if err != nil {
		return 0, err
	}

	r.authRequest(req, r.twinID)
	log.Info().Any("request", req.Body).Msg("request body is ")

	resp, err := r.httpClient.Do(req)
	if err != nil {
		return 0, err
	}

	if resp == nil || resp.StatusCode != http.StatusOK {
		err = parseRespError(resp.Body)
		return 0, errors.Wrapf(err, "failed to update node with twin id %d with status code %s", r.twinID, resp.Status)
	}
	defer resp.Body.Close()

	res, err := r.getNode(url)
	return res.NodeID, err
}

func (r *registrarGateway) updateNodeUptimeV2(twinID uint64, url string, uptime uint64) (err error) {
	report := types.UptimeReportRequest{Uptime: time.Duration(uptime), Timestamp: time.Now()}

	var body bytes.Buffer
	err = json.NewEncoder(&body).Encode(report)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", url, &body)
	if err != nil {
		return err
	}

	r.authRequest(req, twinID)

	resp, err := r.httpClient.Do(req)
	if err != nil {
		return err
	}

	if resp == nil || resp.StatusCode != http.StatusCreated {
		err = parseRespError(resp.Body)
		return errors.Wrap(err, "failed to send node up time report")
	}

	return
}

func getNodeSignature(pubKey string, privKey []byte, timestamp int64) (signatureBase64 string) {
	// Create challenge
	challenge := []byte(fmt.Sprintf("%d:%s", timestamp, pubKey))

	signature := ed25519.Sign(privKey, challenge)
	signatureBase64 = base64.StdEncoding.EncodeToString(signature)
	return
}

func (r *registrarGateway) authRequest(req *http.Request, twinID uint64) {
	// Create authentication challenge
	timestamp := time.Now().Unix()
	challenge := []byte(fmt.Sprintf("%d:%v", timestamp, twinID))
	signature := ed25519.Sign(r.privKey, challenge)
	authHeader := fmt.Sprintf(
		"%s:%s",
		base64.StdEncoding.EncodeToString(challenge),
		base64.StdEncoding.EncodeToString(signature),
	)

	req.Header.Set("X-Auth", authHeader)
	req.Header.Set("Content-Type", "application/json")
}

func parseRespError(body io.ReadCloser) error {
	errResp := struct {
		Error string
	}{}
	err := json.NewDecoder(body).Decode(&errResp)
	if err != nil {
		return err
	}
	return errors.New(errResp.Error)
}
