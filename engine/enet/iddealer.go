package enet

type IDDealer struct {
	handlerMap map[uint32]interface{}
}

func NewIDDealer() *IDDealer {
	return &IDDealer{
		handlerMap: make(map[uint32]interface{}),
	}
}

func (i *IDDealer) RegisterHandler(id uint32, dealer interface{}) bool {
	if dealer == nil {
		return false
	}

	if _, ok := i.handlerMap[id]; ok {
		ELog.ErrorAf("IDDealer Error Id = %v Repeat", id)
		return false
	}

	i.handlerMap[id] = dealer
	return true
}

func (i *IDDealer) FindHandler(id uint32) interface{} {
	if handler, ok := i.handlerMap[id]; ok {
		return handler
	}

	return nil
}
