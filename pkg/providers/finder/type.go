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
	FindClient  *find.Finder
	Folder      string
}

func NewDefaultProvider(tMgr *tags.Manager, client *vim25.Client, folder string) *Provider {
	idx := object.NewSearchIndex(client)
	return &Provider{TagManager: tMgr, Client: client, IndexClient: idx, Folder: folder, FindClient: find.NewFinder(client, true)}
}
