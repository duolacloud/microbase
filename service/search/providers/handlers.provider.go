package providers

import (
	"github.com/duolacloud/microbase/proto/search"
	"github.com/micro/go-micro/v2"
)

func RegisterHandlers(service micro.Service,
	searchServiceHandler search.SearchServiceHandler,
) {
	search.RegisterSearchServiceHandler(service.Server(), searchServiceHandler)
}
