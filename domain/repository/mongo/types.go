package mongo

import (
	"github.com/duolacloud/microbase/domain/entity"
	"gopkg.in/mgo.v2"
)

type Indexed interface {
	Indexes() []mgo.Index
	entity.Entity
}
