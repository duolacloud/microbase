package search

import (
	"context"
	"encoding/json"
	"log"

	"github.com/duolacloud/microbase/domain/entity"
	"github.com/duolacloud/microbase/proto/pagination"
	"github.com/duolacloud/microbase/proto/search"
	"github.com/golang/protobuf/ptypes"
	"github.com/micro/go-micro/v2/client"
	"github.com/thoas/go-funk"
)

type SearchClient interface {
	Create(c context.Context, document *Document) error
	Upsert(c context.Context, document *Document) error
	Update(c context.Context, document *Document) error
	Get(c context.Context, index, typ, id string) (*Document, error)
	Delete(c context.Context, index, typ, id string) error
	List(c context.Context, query *entity.CursorQuery, index, typ string) (items []*Document, extra *entity.CursorExtra, err error)
	Connection(c context.Context, query *entity.ConnectionQuery, index, typ string) (*entity.Connection, error)
	Page(c context.Context, query *entity.PageQuery, index, typ string) (docs []*Document, total int64, err error)

	CreateIndex(c context.Context, index *Index) error
	DeleteIndex(c context.Context, index string) error
	IndexExists(c context.Context, index string) (bool, error)
}

type searchClient struct {
	searchService search.SearchService
}

func NewSearchClient(searchService search.SearchService) SearchClient {
	return &searchClient{
		searchService,
	}
}

func (s *searchClient) Create(c context.Context, document *Document) error {
	b, err := json.Marshal(document.Fields)
	if err != nil {
		return err
	}

	_, err = s.searchService.Create(c, &search.Document{
		Index:  document.Index,
		Type:   document.Type,
		Fields: string(b),
	})
	return err
}

func (s *searchClient) Upsert(c context.Context, document *Document) error {
	b, err := json.Marshal(document.Fields)
	if err != nil {
		return err
	}

	_, err = s.searchService.Upsert(c, &search.Document{
		Index:  document.Index,
		Type:   document.Type,
		Fields: string(b),
	})
	return err
}

func (s *searchClient) Update(c context.Context, document *Document) error {
	b, err := json.Marshal(document.Fields)
	if err != nil {
		return err
	}

	_, err = s.searchService.Update(c, &search.Document{
		Index:  document.Index,
		Type:   document.Type,
		Fields: string(b),
	})
	return err
}

func (s *searchClient) Get(c context.Context, index, typ, id string) (*Document, error) {
	rsp, err := s.searchService.Get(c, &search.GetDocumentRequest{
		Index: index,
		Type:  typ,
		Id:    id,
	})
	if err != nil {
		return nil, err
	}

	var fields map[string]interface{}
	err = json.Unmarshal([]byte(rsp.Fields), &fields)
	if err != nil {
		return nil, err
	}

	doc := &Document{
		Index:  rsp.Index,
		Type:   rsp.Type,
		Fields: fields,
	}

	return doc, nil
}

func (s *searchClient) Delete(c context.Context, index, typ, id string) error {
	_, err := s.searchService.Delete(c, &search.DeleteDocumentRequest{
		Index: index,
		Type:  typ,
		Id:    id,
	})
	return err
}

func (s *searchClient) List(c context.Context, query *entity.CursorQuery, index, typ string) (items []*Document, extra *entity.CursorExtra, err error) {
	filterB, err := json.Marshal(query.Filter)
	if err != nil {
		return nil, nil, err
	}

	var cursorDirection pagination.CursorDirection
	if query.Direction == entity.CursorDirectionBefore {
		cursorDirection = pagination.CursorDirection_before
	} else {
		cursorDirection = pagination.CursorDirection_after
	}

	rsp, err := s.searchService.List(c, &search.ListRequest{
		Index: index,
		Type:  typ,
		Query: &pagination.ListQuery{
			Filter: string(filterB),
			Cursor: query.Cursor,
			Orders: funk.Map(query.Orders, func(o *entity.Order) *pagination.Order {
				var direction pagination.OrderDirection
				if o.Direction == entity.OrderDirectionDesc {
					direction = pagination.OrderDirection_DESC
				} else {
					direction = pagination.OrderDirection_ASC
				}

				return &pagination.Order{
					Field:     o.Field,
					Direction: direction,
				}
			}).([]*pagination.Order),
			Size:      int32(query.Size),
			Direction: cursorDirection,
			NeedTotal: query.NeedTotal,
			Fields:    query.Fields,
		},
	})

	if err != nil {
		return nil, nil, err
	}

	return funk.Map(rsp.Documents, func(o *search.Document) *Document {
			var fields map[string]interface{}
			_ = json.Unmarshal([]byte(o.Fields), &fields)

			return &Document{
				Index:  o.Index,
				Type:   o.Type,
				Fields: fields,
			}
		}).([]*Document), &entity.CursorExtra{
			Total:       rsp.Total,
			HasNext:     rsp.HasNext,
			HasPrevious: rsp.HasPrevious,
			StartCursor: rsp.StartCursor,
			EndCursor:   rsp.EndCursor,
		}, nil
}

func (s *searchClient) Connection(c context.Context, query *entity.ConnectionQuery, index, typ string) (*entity.Connection, error) {
	rsp, err := s.searchService.Connection(c, &search.ConnectionRequest{
		Index: index,
		Type:  typ,
		Query: query.ToPB(),
	})

	if err != nil {
		return nil, err
	}

	return &entity.Connection{
		Total: rsp.Total,
		Edges: funk.Map(rsp.Edges, func(e *pagination.Edge) *entity.Edge {
			var doc search.Document
			_ = ptypes.UnmarshalAny(e.Node, &doc)

			var fields map[string]interface{}
			_ = json.Unmarshal([]byte(doc.Fields), &fields)

			return &entity.Edge{
				Cursor: e.Cursor,
				Node: &Document{
					Index:  doc.Index,
					Type:   doc.Type,
					Fields: fields,
				},
			}
		}).([]*entity.Edge),
		PageInfo: entity.PageInfo{
			HasPrevious: rsp.PageInfo.HasPrevious,
			HasNext:     rsp.PageInfo.HasNext,
			EndCursor:   rsp.PageInfo.EndCursor,
			StartCursor: rsp.PageInfo.StartCursor,
		},
	}, nil
}

func (s *searchClient) Page(c context.Context, query *entity.PageQuery, index, typ string) (docs []*Document, total int64, err error) {
	filterB, err := json.Marshal(query.Filter)
	if err != nil {
		return
	}

	var rsp *search.PageResponse
	rsp, err = s.searchService.Page(c, &search.PageRequest{
		Index: index,
		Type:  typ,
		Query: &pagination.PageQuery{
			Filter:   string(filterB),
			PageNo:   int64(query.PageNo),
			PageSize: int32(query.PageSize),
			Orders: funk.Map(query.Orders, func(o *entity.Order) *pagination.Order {
				var direction pagination.OrderDirection
				if o.Direction == entity.OrderDirectionDesc {
					direction = pagination.OrderDirection_DESC
				} else {
					direction = pagination.OrderDirection_ASC
				}

				return &pagination.Order{
					Field:     o.Field,
					Direction: direction,
				}
			}).([]*pagination.Order),
		},
	})
	if err != nil {
		return
	}

	return funk.Map(rsp.Documents, func(doc *search.Document) *Document {
		var fields map[string]interface{}
		_ = json.Unmarshal([]byte(doc.Fields), &fields)

		return &Document{
			Index:  doc.Index,
			Type:   doc.Type,
			Fields: fields,
		}
	}).([]*Document), rsp.Total, nil
}

func (s *searchClient) CreateIndex(c context.Context, index *Index) error {
	log.Printf("client CreateIndex")
	mapping, err := json.Marshal(index.Mapping)
	if err != nil {
		return err
	}

	_, err = s.searchService.CreateIndex(c, &search.CreateIndexRequest{
		Index: &search.Index{
			Name:    index.Name,
			Mapping: string(mapping),
		},
	}, client.WithRetries(0))
	return err
}

func (s *searchClient) DeleteIndex(c context.Context, index string) error {
	_, err := s.searchService.DeleteIndex(c, &search.DeleteIndexRequest{Index: index})
	return err
}

func (s *searchClient) IndexExists(c context.Context, index string) (bool, error) {
	r, err := s.searchService.IndexExists(c, &search.IndexExistsRequest{Index: index})
	if err != nil {
		return false, err
	}

	return r.Exists, err
}
