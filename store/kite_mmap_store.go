package store

import (
	"container/list"
	"sync"
)

type KiteMMapStore struct {
	datalink *list.List                              //用于LRU
	idx      map[string] /*messageId*/ *list.Element //用于LRU
	lock     sync.RWMutex
	maxcap   int
	path     string
}

func NewKiteMMapStore(path string, initcap, maxcap int) *KiteMMapStore {
	return &KiteMMapStore{
		datalink: list.New(),
		idx:      make(map[string]*list.Element, initcap),
		maxcap:   maxcap,
		path:     path}
}

func (self *KiteMMapStore) Query(messageId string) *MessageEntity {
	self.lock.RLock()
	defer self.lock.RUnlock()
	e, ok := self.idx[messageId]
	if !ok {
		return nil
	}
	//将当前节点放到最前面
	return e.Value.(*MessageEntity)

}
func (self *KiteMMapStore) Save(entity *MessageEntity) bool {
	self.lock.Lock()
	defer self.lock.Unlock()
	cl := self.datalink.Len()
	if cl >= self.maxcap {
		delete(self.idx, entity.Header.GetMessageId())
		self.datalink.Remove(self.datalink.Back())
	}
	e := self.datalink.PushFront(entity)
	self.idx[entity.Header.GetMessageId()] = e
	return true
}
func (self *KiteMMapStore) Commit(messageId string) bool {
	self.lock.Lock()
	defer self.lock.Unlock()
	e, ok := self.idx[messageId]
	if !ok {
		return false
	}
	entity := e.Value.(*MessageEntity)
	entity.Commit = true
	return true
}
func (self *KiteMMapStore) Rollback(messageId string) bool {
	self.lock.Lock()
	defer self.lock.Unlock()
	e, ok := self.idx[messageId]
	if !ok {
		return true
	}
	self.datalink.Remove(e)
	return true
}
func (self *KiteMMapStore) UpdateEntity(entity *MessageEntity) bool {
	self.lock.Lock()
	defer self.lock.Unlock()
	v, ok := self.idx[entity.MessageId]
	if !ok {
		return true
	}

	e := v.Value.(*MessageEntity)
	e.DeliverCount = entity.DeliverCount
	e.NextDeliverTime = entity.NextDeliverTime
	e.SuccGroups = entity.SuccGroups
	e.FailGroups = entity.FailGroups
	return true
}
func (self *KiteMMapStore) Delete(messageId string) bool {
	return self.Rollback(messageId)

}

//根据kiteServer名称查询需要重投的消息 返回值为 是否还有更多、和本次返回的数据结果
func (self *KiteMMapStore) PageQueryEntity(hashKey string, kiteServer string, nextDeliveryTime int64, startIdx, limit int32) (bool, []*MessageEntity) {

	return false, nil
}