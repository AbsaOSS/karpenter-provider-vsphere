package finder

import (
	"github.com/absaoss/karpenter-provider-vsphere/pkg/providers/vsphereclient"
	"github.com/vmware/govmomi/find"
	"github.com/vmware/govmomi/object"
)

type Provider struct {
	Session     *vsphereclient.Session
	IndexClient *object.SearchIndex
	DC          *object.Datacenter
	FindClient  *find.Finder
	Folder      string
	ClusterName string
}

func NewDefaultProvider(session *vsphereclient.Session, findClient *find.Finder, dc *object.Datacenter, folder, cluster string) *Provider {
	idx := object.NewSearchIndex(session.Vim)
	// Set Datacenter globally for find operations
	findClient.SetDatacenter(dc)
	return &Provider{ClusterName: cluster, Session: session, IndexClient: idx, Folder: folder, FindClient: findClient, DC: dc}
}
