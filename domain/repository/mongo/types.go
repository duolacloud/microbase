package mongo

import (
	"github.com/duolacloud/microbase/domain/model"
	"gopkg.in/mgo.v2"
)

type Indexed interface {
	Indexes() []mgo.Index
	model.Model
}
