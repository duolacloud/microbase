package handlers

import (
	"context"

	"github.com/duolacloud/microbase/proto/pagination"
	"github.com/duolacloud/microbase/proto/search"
)

type searchServiceHandler struct {
}

func NewSearchHandler() search.SearchServiceHandler {
	return &searchServiceHandler{}
}

func (h *searchServiceHandler) Upsert(c context.Context, req *search.UpsertDocumentRequest, rsp *search.UpsertDocumentResponse) error {
	return nil
}

func (h *searchServiceHandler) BatchUpsert(c context.Context, req *search.BatchUpsertDocumentRequest, rsp *search.BatchUpsertDocumentResponse) error {
	return nil
}

func (h *searchServiceHandler) Get(c context.Context, req *search.GetDocumentRequest, rsp *search.Document) error {
	return nil
}

func (h *searchServiceHandler) BatchGet(c context.Context, req *search.BatchGetDocumentRequest, rsp *search.BatchGetDocumentResponse) error {
	return nil
}

func (h *searchServiceHandler) Search(c context.Context, req *search.SearchRequest, rsp *search.SearchResponse) error {
	return nil
}

func (h *searchServiceHandler) Connection(context.Context, *pagination.ConnectionQuery, *pagination.Connection) error {
	return nil
}
