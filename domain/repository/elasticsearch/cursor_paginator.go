package elasticsearch

import (
	"context"
	"encoding/json"

	"github.com/duolacloud/microbase/domain/entity"
	"github.com/olivere/elastic/v6"
	// "github.com/olivere/elastic/v7"
)

type CursorPaginator struct {
	client *elastic.Client
}

func NewCursorPaginator(client *elastic.Client) *CursorPaginator {
	return &CursorPaginator{
		client,
	}
}

func (p *CursorPaginator) Paginate(c context.Context, query *entity.CursorQuery, index, typ string) ([]*entity.Document, *entity.CursorExtra, error) {
	// query := elastic.NewTermQuery("", "")

	result, err := p.client.Search().
		Index(index).
		Type(typ).
		// Query(query).
		// Sort("", false).
		Do(c)
	if err != nil {
		return nil, nil, err
	}

	if len(result.Hits.Hits) == 0 {
		return nil, &entity.CursorExtra{
			HasNext:     false,
			HasPrevious: false,
			Total:       0,
		}, nil
	}

	docs := make([]*entity.Document, len(result.Hits.Hits))
	for i, r := range result.Hits.Hits {
		doc := &entity.Document{
			Index: index,
			Type:  typ,
		}

		err = json.Unmarshal(*r.Source, &doc.Data)
		if err != nil {
			return nil, nil, err
		}

		docs[i] = doc
	}

	return docs, &entity.CursorExtra{
		HasNext:     false,
		HasPrevious: false,
		Total:       int(result.Hits.TotalHits),
	}, nil
}
