package finder

import (
	"github.com/vmware/govmomi/find"
	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/vapi/tags"
	"github.com/vmware/govmomi/vim25"
)

type Provider struct {
	TagManager  *tags.Manager
	Client      *vim25.Client
	IndexClient *object.SearchIndex
	DC          *object.Datacenter
	FindClient  *find.Finder
	Folder      string
	ClusterName string
}

func NewDefaultProvider(tMgr *tags.Manager, client *vim25.Client, findClient *find.Finder, dc *object.Datacenter, folder, cluster string) *Provider {
	idx := object.NewSearchIndex(client)
	// Set Datacenter globally for find operations
	findClient.SetDatacenter(dc)
	return &Provider{ClusterName: cluster, TagManager: tMgr, Client: client, IndexClient: idx, Folder: folder, FindClient: findClient, DC: dc}
}
