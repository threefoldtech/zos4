package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/threefoldtech/zos/pkg/schema"

	"github.com/threefoldtech/zos/pkg/gedis/types/directory"
)

type nodeStore struct {
	Nodes []*directory.TfgridNode2 `json:"nodes"`
}

func LoadNodeStore() (nodeStore, error) {
	store := nodeStore{
		Nodes: []*directory.TfgridNode2{},
	}
	f, err := os.OpenFile("nodes.json", os.O_RDONLY, 0660)
	if err != nil {
		if os.IsNotExist(err) {
			return store, nil
		}
		return store, err
	}
	defer f.Close()
	if err := json.NewDecoder(f).Decode(&store); err != nil {
		return store, err
	}
	return store, nil
}

func (s *nodeStore) Save() error {
	f, err := os.OpenFile("nodes.json", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0660)
	if err != nil {
		return err
	}
	defer f.Close()
	if err := json.NewEncoder(f).Encode(s); err != nil {
		return err
	}
	return nil
}

func (s *nodeStore) List() []*directory.TfgridNode2 {
	return s.Nodes
}

func (s *nodeStore) Get(nodeID string) (*directory.TfgridNode2, error) {
	for _, n := range s.Nodes {
		if n.NodeID == nodeID {
			return n, nil
		}
	}
	return nil, fmt.Errorf("node %s not found", nodeID)
}

func (s *nodeStore) Add(node directory.TfgridNode2) error {
	for i, n := range s.Nodes {
		if n.NodeID == node.NodeID {
			s.Nodes[i].FarmID = node.FarmID
			s.Nodes[i].OsVersion = node.OsVersion
			s.Nodes[i].Location = node.Location
			s.Nodes[i].Updated = schema.Date{time.Now()}
			return nil
		}
	}

	node.Created = schema.Date{time.Now()}
	node.Updated = schema.Date{time.Now()}
	s.Nodes = append(s.Nodes, &node)
	return nil
}

func (s *nodeStore) updateTotalCapacity(nodeID string, cap directory.TfgridNodeResourceAmount1) error {
	return s.updateCapacity(nodeID, "total", cap)
}
func (s *nodeStore) updateReservedCapacity(nodeID string, cap directory.TfgridNodeResourceAmount1) error {
	return s.updateCapacity(nodeID, "reserved", cap)
}
func (s *nodeStore) updateUsedCapacity(nodeID string, cap directory.TfgridNodeResourceAmount1) error {
	return s.updateCapacity(nodeID, "used", cap)
}

func (s *nodeStore) updateCapacity(nodeID string, t string, cap directory.TfgridNodeResourceAmount1) error {
	node, err := s.Get(nodeID)
	if err != nil {
		return err
	}

	switch t {
	case "total":
		node.TotalResources = cap
	case "reserved":
		node.ReservedResources = cap
	case "used":
		node.UsedResources = cap
	default:
		return fmt.Errorf("unsupported capacity type: %v", t)
	}

	return nil
}

func (s *nodeStore) SetInterfaces(nodeID string, ifaces []directory.TfgridNodeIface1) error {
	node, err := s.Get(nodeID)
	if err != nil {
		return err
	}

	node.Ifaces = ifaces
	return nil
}

func (s *nodeStore) SetPublicConfig(nodeID string, cfg directory.TfgridNodePublicIface1) error {
	node, err := s.Get(nodeID)
	if err != nil {
		return err
	}

	node.PublicConfig = &cfg
	return nil
}

func (s *nodeStore) SetWGPorts(nodeID string, ports []uint) error {
	node, err := s.Get(nodeID)
	if err != nil {
		return err
	}

	node.WGPorts = ports
	return nil
}
