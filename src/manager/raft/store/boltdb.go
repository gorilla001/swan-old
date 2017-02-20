package store

import (
	"errors"

	"github.com/Dataman-Cloud/swan/src/manager/raft/types"

	"github.com/boltdb/bolt"
	"github.com/coreos/etcd/raft/raftpb"
)

type BoltbDb struct {
	*bolt.DB
}

var (
	bucketKeyStorageVersion = []byte("v1")
	bucketKeyApps           = []byte("apps")
	bucketKeyFramework      = []byte("framework")
	bucketKeyTasks          = []byte("tasks")
	bucketKeyVersions       = []byte("versions")
	bucketKeySlots          = []byte("slots")
	bucketKeyOfferAllocator = []byte("offer_allocator")
	bucketKeyNodes          = []byte("Nodes")
	BucketKeyRaftState      = []byte("raft_hard_state")

	BucketKeyData = []byte("data")
)

var (
	ErrAppUnknown              = errors.New("boltdb: app unknown")
	ErrTaskUnknown             = errors.New("boltdb: task unknown")
	ErrVersionUnknown          = errors.New("boltdb: version unknown")
	ErrSlotUnknown             = errors.New("boltdb: slot unknow")
	ErrNodeUnknown             = errors.New("boltdb: node unknow")
	ErrNilStoreAction          = errors.New("boltdb: nil store action")
	ErrUndefineStoreAction     = errors.New("boltdb: undefined store action")
	ErrUndefineAppStoreAction  = errors.New("boltdb: undefined app store action")
	ErrUndefineFrameworkAction = errors.New("boltdb: undefined framework store action")
	ErrUndefineTaskAction      = errors.New("boltdb: undefined task store action")
	ErrUndefineVersionAction   = errors.New("boltdb: undefined version store action")
	ErrUndefineSlotAction      = errors.New("boltdb: undefined slot store action")
	ErrUndefineNodeAction      = errors.New("boltdb: undefined node store action")
)

func NewBoltbdStore(db *bolt.DB) (*BoltbDb, error) {
	return &BoltbDb{db}, nil
}

func createBucketIfNotExists(tx *bolt.Tx, keys ...[]byte) (*bolt.Bucket, error) {
	bkt, err := tx.CreateBucketIfNotExists(keys[0])
	if err != nil {
		return nil, err
	}

	for _, key := range keys[1:] {
		bkt, err = bkt.CreateBucketIfNotExists(key)
		if err != nil {
			return nil, err
		}
	}

	return bkt, nil
}

func getBucket(tx *bolt.Tx, keys ...[]byte) *bolt.Bucket {
	bkt := tx.Bucket(keys[0])

	for _, key := range keys[1:] {
		if bkt == nil {
			break
		}

		bkt = bkt.Bucket(key)
	}

	return bkt
}

func (db *BoltbDb) DoStoreActions(actions []*types.StoreAction) error {
	tx, err := db.Begin(true)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for _, storeAction := range actions {
		if err := doStoreAction(tx, storeAction); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func doStoreAction(tx *bolt.Tx, action *types.StoreAction) error {
	if action == nil {
		return ErrNilStoreAction
	}

	actionTarget := action.GetTarget()
	if actionTarget == nil {
		return ErrUndefineStoreAction
	}

	switch actionTarget.(type) {
	case *types.StoreAction_Application:
		return doAppStoreAction(tx, action.Action, action.GetApplication())
	case *types.StoreAction_Framework:
		return doFrameworkStoreAction(tx, action.Action, action.GetFramework())
	case *types.StoreAction_Task:
		return doTaskStoreAction(tx, action.Action, action.GetTask())
	case *types.StoreAction_Version:
		return doVersionStoreAction(tx, action.Action, action.GetVersion())
	case *types.StoreAction_Slot:
		return doSlotStoreAction(tx, action.Action, action.GetSlot())
	case *types.StoreAction_OfferAllocatorItem:
		return doOfferAllocatorItemStoreAction(tx, action.Action, action.GetOfferAllocatorItem())
	case *types.StoreAction_Node:
		return doNodeStoreAction(tx, action.Action, action.GetNode())
	default:
		return ErrUndefineStoreAction
	}
}

func doAppStoreAction(tx *bolt.Tx, action types.StoreActionKind, app *types.Application) error {
	switch action {
	case types.StoreActionKindCreate:
		return createApp(tx, app)
	case types.StoreActionKindUpdate:
		return updateApp(tx, app)
	case types.StoreActionKindRemove:
		return removeApp(tx, app.ID)
	default:
		return ErrUndefineAppStoreAction
	}
}

func doFrameworkStoreAction(tx *bolt.Tx, action types.StoreActionKind, framework *types.Framework) error {
	switch action {
	case types.StoreActionKindCreate, types.StoreActionKindUpdate:
		return putFramework(tx, framework)
	case types.StoreActionKindRemove:
		return removeFramework(tx)
	default:
		return ErrUndefineFrameworkAction
	}
}

func doSlotStoreAction(tx *bolt.Tx, action types.StoreActionKind, slot *types.Slot) error {
	switch action {
	case types.StoreActionKindCreate:
		return createSlot(tx, slot)
	case types.StoreActionKindUpdate:
		return updateSlot(tx, slot)
	case types.StoreActionKindRemove:
		return removeSlot(tx, slot.AppID, slot.ID)
	default:
		return ErrUndefineSlotAction
	}
}

func doTaskStoreAction(tx *bolt.Tx, action types.StoreActionKind, task *types.Task) error {
	switch action {
	case types.StoreActionKindCreate:
		return createTask(tx, task)
	case types.StoreActionKindUpdate:
		return updateTask(tx, task)
	case types.StoreActionKindRemove:
		return removeTask(tx, task.AppID, task.SlotID, task.ID)
	default:
		return ErrUndefineTaskAction
	}
}

func doVersionStoreAction(tx *bolt.Tx, action types.StoreActionKind, version *types.Version) error {
	switch action {
	case types.StoreActionKindCreate:
		return createVersion(tx, version)
	case types.StoreActionKindUpdate:
		return updateVersion(tx, version)
	case types.StoreActionKindRemove:
		return removeVersion(tx, version.AppID, version.ID)
	default:
		return ErrUndefineVersionAction
	}
}

func doOfferAllocatorItemStoreAction(tx *bolt.Tx, action types.StoreActionKind, item *types.OfferAllocatorItem) error {
	switch action {
	case types.StoreActionKindCreate:
		return createOfferAllocatorItem(tx, item)
	case types.StoreActionKindRemove:
		return removeOfferAllocatorItem(tx, item)
	default:
		return ErrUndefineVersionAction
	}
}

func doNodeStoreAction(tx *bolt.Tx, action types.StoreActionKind, node *types.Node) error {
	switch action {
	case types.StoreActionKindCreate:
		return createNode(tx, node)
	case types.StoreActionKindUpdate:
		return updateNode(tx, node)
	case types.StoreActionKindRemove:
		return removeNode(tx, node.ID)
	default:
		return ErrUndefineNodeAction
	}
}

func (db *BoltbDb) SaveRaftState(state raftpb.HardState) error {
	return db.Update(func(tx *bolt.Tx) error {
		return putRaftState(tx, state)
	})
}

func (db *BoltbDb) GetRaftState() (raftpb.HardState, error) {
	var state raftpb.HardState

	if err := db.View(func(tx *bolt.Tx) error {
		var err error
		state, err = getRaftState(tx)
		return err

	}); err != nil {
		return state, err
	}

	return state, nil
}

func (db *BoltbDb) GetNodes() ([]*types.Node, error) {
	var nodes []*types.Node

	if err := db.View(func(tx *bolt.Tx) error {
		nodesBkt := getNodesBucket(tx)
		if nodesBkt == nil {
			nodes = []*types.Node{}
			return nil
		}

		return nodesBkt.ForEach(func(k, v []byte) error {
			nodeBkt := getNodeBucket(tx, string(k))
			if nodeBkt == nil {
				return nil
			}

			node := &types.Node{}
			p := nodeBkt.Get(BucketKeyData)
			if err := node.Unmarshal(p); err != nil {
				return err
			}

			nodes = append(nodes, node)
			return nil
		})

	}); err != nil {
		return nil, err
	}

	return nodes, nil
}
