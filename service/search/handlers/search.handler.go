package handlers

import (
	"context"
	"encoding/json"
	"log"

	"github.com/duolacloud/microbase/client/search"
	"github.com/duolacloud/microbase/domain/entity"
	"github.com/duolacloud/microbase/logger"
	"github.com/duolacloud/microbase/proto/pagination"
	pb "github.com/duolacloud/microbase/proto/search"
	"github.com/duolacloud/microbase/service/search/repositories"
	"github.com/golang/protobuf/ptypes"
	"github.com/thoas/go-funk"
	"google.golang.org/protobuf/types/known/emptypb"
)

type searchServiceHandler struct {
	indexRepository    repositories.IndexRepository
	documentRepository repositories.DocumentRepository
}

func NewSearchHandler(indexRepository repositories.IndexRepository, documentRepository repositories.DocumentRepository) pb.SearchServiceHandler {
	return &searchServiceHandler{
		indexRepository,
		documentRepository,
	}
}

func (h *searchServiceHandler) Create(c context.Context, req *pb.Document, rsp *emptypb.Empty) error {
	var fields map[string]interface{}
	err := json.Unmarshal([]byte(req.Fields), &fields)
	if err != nil {
		return err
	}

	log.Printf("Create fields: %v", fields)

	return h.documentRepository.Create(c, &search.Document{
		Index:  req.Index,
		Type:   req.Type,
		Fields: fields,
	})
}

func (h *searchServiceHandler) Upsert(c context.Context, req *pb.Document, rsp *emptypb.Empty) error {
	var fields map[string]interface{}
	err := json.Unmarshal([]byte(req.Fields), &fields)
	if err != nil {
		return err
	}

	return h.documentRepository.Upsert(c, &search.Document{
		Index:  req.Index,
		Type:   req.Type,
		Fields: fields,
	})
}

func (h *searchServiceHandler) Update(c context.Context, req *pb.Document, rsp *emptypb.Empty) error {
	var fields map[string]interface{}
	err := json.Unmarshal([]byte(req.Fields), &fields)
	if err != nil {
		return err
	}

	return h.documentRepository.Update(c, &search.Document{
		Index:  req.Index,
		Type:   req.Type,
		Fields: fields,
	})
}

func (h *searchServiceHandler) Delete(c context.Context, req *pb.DeleteDocumentRequest, rsp *emptypb.Empty) error {
	return h.documentRepository.Delete(c, req.Index, req.Type, req.Id)
}

func (h *searchServiceHandler) BatchUpsert(c context.Context, req *pb.BatchUpsertDocumentRequest, rsp *pb.BatchUpsertDocumentResponse) error {
	/* TODO
	r, err := h.documentRepository.BatchUpsert(c, req.Index, req.Type, req.Id)
	if err != nil {
		return err
	}
	*/
	return nil
}

func (h *searchServiceHandler) Get(c context.Context, req *pb.GetDocumentRequest, rsp *pb.Document) error {
	logger.Infof("Get index: %v", req.Index, req.Type, req.Id)
	doc, err := h.documentRepository.Get(c, req.Index, req.Type, req.Id)
	if err != nil {
		return err
	}

	fieldsB, err := json.Marshal(doc.Fields)
	if err != nil {
		return err
	}

	rsp.Index = doc.Index
	rsp.Type = doc.Type
	rsp.Fields = string(fieldsB)
	return nil
}

func (h *searchServiceHandler) BatchGet(c context.Context, req *pb.BatchGetDocumentRequest, rsp *pb.BatchGetDocumentResponse) error {
	/*
		docs, err := h.documentRepository.BatchGet(c, req.Index, req.Type, req.Ids)
		if err != nil {
			return err
		}
	*/
	return nil
}

func (h *searchServiceHandler) Search(c context.Context, req *pb.SearchRequest, rsp *pb.SearchResponse) error {
	return nil
}

func (h *searchServiceHandler) List(c context.Context, req *pb.ListRequest, rsp *pb.ListResponse) error {
	var query entity.CursorQuery
	query.FromPB(req.Query)

	docs, extra, err := h.documentRepository.List(c, &query, req.Index, req.Type)
	if err != nil {
		return err
	}

	rsp.Documents = funk.Map(docs, func(doc *search.Document) *pb.Document {
		var fieldsB, _ = json.Marshal(doc.Fields)

		return &pb.Document{
			Index:  doc.Index,
			Type:   doc.Type,
			Fields: string(fieldsB),
		}
	}).([]*pb.Document)

	rsp.Total = extra.Total
	rsp.HasNext = extra.HasNext
	rsp.HasPrevious = extra.HasPrevious
	rsp.StartCursor = extra.StartCursor
	rsp.EndCursor = extra.EndCursor

	return nil
}

func (h *searchServiceHandler) Page(c context.Context, req *pb.PageRequest, rsp *pb.PageResponse) error {
	docs, total, err := h.documentRepository.Page(c, &entity.PageQuery{}, req.Index, req.Type)
	if err != nil {
		return err
	}

	rsp.Total = total
	rsp.Documents = funk.Map(docs, func(doc *search.Document) *pb.Document {
		fieldsB, _ := json.Marshal(doc.Fields)

		return &pb.Document{
			Index:  doc.Index,
			Type:   doc.Type,
			Fields: string(fieldsB),
		}
	}).([]*pb.Document)

	return nil
}

func (h *searchServiceHandler) Connection(c context.Context, req *pb.ConnectionRequest, rsp *pagination.Connection) error {
	var query entity.ConnectionQuery
	query.FromPB(req.Query)
	conn, err := h.documentRepository.Connection(c, &query, req.Index, req.Type)
	if err != nil {
		return err
	}

	rsp.Total = conn.Total
	rsp.PageInfo = conn.PageInfo.ToPB()
	rsp.Edges = funk.Map(conn.Edges, func(edge *entity.Edge) *pagination.Edge {
		doc := edge.Node.(*search.Document)

		fieldsB, _ := json.Marshal(doc.Fields)

		anyNode, _ := ptypes.MarshalAny(&pb.Document{
			Index:  doc.Index,
			Type:   doc.Type,
			Fields: string(fieldsB),
		})

		return &pagination.Edge{
			Cursor: edge.Cursor,
			Node:   anyNode,
		}
	}).([]*pagination.Edge)

	return nil
}

func (h *searchServiceHandler) CreateIndex(c context.Context, req *pb.CreateIndexRequest, rsp *emptypb.Empty) error {
	log.Printf("handler.CreateIndex")
	var mapping map[string]interface{}
	err := json.Unmarshal([]byte(req.Index.Mapping), &mapping)
	if err != nil {
		return err
	}

	return h.indexRepository.Create(c, &search.Index{
		Name:    req.Index.Name,
		Mapping: mapping,
	})
}

func (h *searchServiceHandler) DeleteIndex(c context.Context, req *pb.DeleteIndexRequest, rsp *emptypb.Empty) error {
	return h.indexRepository.Delete(c, req.Index)
}

func (h *searchServiceHandler) IndexExists(c context.Context, req *pb.IndexExistsRequest, rsp *pb.IndexExistsResponse) error {
	exists, err := h.indexRepository.IndexExists(c, req.Index)
	if err != nil {
		return err
	}

	rsp.Exists = exists
	return nil
}
