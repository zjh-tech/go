package frame

type IDDealer struct {
	handler_map map[uint32]interface{}
}

func NewIDDealer() *IDDealer {
	return &IDDealer{
		handler_map: make(map[uint32]interface{}),
	}
}

func (i *IDDealer) RegisterHandler(id uint32, dealer interface{}) bool {
	if dealer == nil {
		return false
	}

	if _, ok := i.handler_map[id]; ok {
		ELog.ErrorAf("IDDealer Error Id = %v Repeat", id)
		return false
	}

	i.handler_map[id] = dealer
	return true
}

func (i *IDDealer) FindHandler(id uint32) interface{} {
	if handler, ok := i.handler_map[id]; ok {
		return handler
	}

	return nil
}
