package pagination

import (
	"google.golang.org/protobuf/types/known/wrapperspb"
)

type connectionQueryBuilder struct {
	req *ConnectionQuery
}

func NewConnectionQueryBuilder() *connectionQueryBuilder {
	return &connectionQueryBuilder{
		req: &ConnectionQuery{},
	}
}

func (b *connectionQueryBuilder) Build() *ConnectionQuery {
	return b.req
}

func (b *connectionQueryBuilder) After(after *string) *connectionQueryBuilder {
	if after != nil {
		b.req.After = wrapperspb.String(*after)
	}
	return b
}

func (b *connectionQueryBuilder) Before(before *string) *connectionQueryBuilder {
	if before != nil {
		b.req.Before = wrapperspb.String(*before)
	}

	return b
}

func (b *connectionQueryBuilder) First(first *int) *connectionQueryBuilder {
	if first != nil {
		b.req.First = wrapperspb.Int32(int32(*first))
	}
	return b
}

func (b *connectionQueryBuilder) Last(last *int) *connectionQueryBuilder {
	if last != nil {
		b.req.Last = wrapperspb.Int32(int32(*last))
	}
	return b
}

func (b *connectionQueryBuilder) Filter(filter *string) *connectionQueryBuilder {
	if filter != nil {
		b.req.Filter = *filter
	}
	return b
}

func (b *connectionQueryBuilder) Reverse(reverse *bool) *connectionQueryBuilder {
	if reverse != nil {
		// b.req.Reverse = *reverse
	}
	return b
}

func (b *connectionQueryBuilder) Orders(orders ...*Order) *connectionQueryBuilder {
	if orders != nil {
		b.req.Orders = orders
	}

	return b
}

func (b *connectionQueryBuilder) NeedTotal(v bool) *connectionQueryBuilder {
	b.req.NeedTotal = v
	return b
}
