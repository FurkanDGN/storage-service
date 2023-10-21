package zookeeper

import (
	"encoding/json"
	"errors"
	"github.com/go-zookeeper/zk"
	"time"
)

type NodeDescription struct {
	ID            string `json:"id"`
	ServerAddress string `json:"serverAddress"`
}

type Zookeeper struct {
	conn *zk.Conn
}

func Connect(servers []string) (*Zookeeper, error) {
	conn, _, err := zk.Connect(servers, time.Second*5)
	if err != nil {
		return nil, err
	}

	for conn.State() != zk.StateHasSession {
		time.Sleep(time.Second)
	}

	return &Zookeeper{conn: conn}, nil
}

func (z *Zookeeper) CreateBaseNode() error {
	_, err := z.conn.Create("/videohub", []byte{}, 0, zk.WorldACL(zk.PermAll))
	if err != nil && !errors.Is(err, zk.ErrNodeExists) {
		return err
	}

	return nil
}

func (z *Zookeeper) CreateChildNode(nodeDesc NodeDescription) error {
	nodeData, err := json.Marshal(nodeDesc)
	if err != nil {
		return err
	}

	_, err = z.conn.Create("/videohub/"+nodeDesc.ID, nodeData, zk.FlagEphemeral, zk.WorldACL(zk.PermAll))
	if err != nil {
		return err
	}

	return nil
}

func (z *Zookeeper) fetchChildNodes() ([]string, error) {
	children, _, err := z.conn.Children("/videohub")
	if err != nil {
		return nil, err
	}
	return children, nil
}

func (z *Zookeeper) fetchSpecificChildNode(childNodeName string) (*NodeDescription, error) {
	data, _, err := z.conn.Get("/videohub/" + childNodeName)
	if err != nil {
		return nil, err
	}

	var nodeDesc NodeDescription
	if err := json.Unmarshal(data, &nodeDesc); err != nil {
		return nil, err
	}

	return &nodeDesc, nil
}
