package zk

import (
	"github.com/Dataman-Cloud/swan/types"
)

// As Nested Field of AppHolder, CreateVersion Require Transaction Lock
func (zk *ZKStore) CreateVersion(aid string, version *types.Version) error {
	if zk.GetVersion(aid, version.ID) != nil {
		return errVersionAlreadyExists
	}

	bs, err := encode(version)
	if err != nil {
		return err
	}

	path := keyApp + "/" + aid
	return zk.createAll(path, bs)
}

func (zk *ZKStore) GetVersion(aid, vid string) *types.Version {
	return nil
}

func (zk *ZKStore) ListVersions(aid string) []*types.Version {
	return nil
}
