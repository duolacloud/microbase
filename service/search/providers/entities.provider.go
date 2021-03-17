package providers

import "github.com/duolacloud/microbase/datasource"

type EntityMap struct {
}

func (m *EntityMap) GetEntities() []interface{} {
	return []interface{}{}
}

func NewEntityMap() datasource.EntityMap {
	return &EntityMap{}
}
