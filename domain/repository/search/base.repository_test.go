package search

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/duolacloud/microbase/client/search"
	search_datasource "github.com/duolacloud/microbase/datasource/search"
	"github.com/duolacloud/microbase/domain/entity"
	"github.com/duolacloud/microbase/domain/repository"

	"github.com/duolacloud/microbase/multitenancy"
	"github.com/duolacloud/microbase/providers"
	"github.com/micro/go-micro/v2/config"
	"github.com/micro/go-micro/v2/config/source/memory"
	"github.com/micro/go-micro/v2/logger"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli/v2"
)

type User struct {
	ID         string     `json:"id,omitempty" elastic:"type:keyword"`
	Name       *string    `json:"name,omitempty" elastic:"type:keyword"`
	Age        *int       `json:"age,omitempty" elastic:"type:integer"`
	CreateTime *time.Time `json:"ctime,omitempty" elastic:"type:date"`
	UpdateTime *time.Time `json:"mtime,omitempty" elastic:"type:date"`
	DeleteTime *time.Time `json:"dtime,omitempty" elastic:"type:date"`
	Deleted    *bool      `json:"deleted,omitempty" elastic:"type:boolean"`
}

func (u *User) Unique() interface{} {
	return map[string]interface{}{
		"id": u.ID,
	}
}

func getConfig() (config.Config, error) {
	config, err := config.NewConfig()
	if err != nil {
		return nil, err
	}

	data := []byte(`{
		"elasticsearch": {
			"addrs": ["http://localhost:9200"]
		}
	}`)
	source := memory.NewSource(memory.WithJSON(data))

	err = config.Load(source)
	if err != nil {
		return nil, err
	}

	return config, nil
}

type EntityMap struct {
}

func (EntityMap) GetEntities() []interface{} {
	return []interface{}{
		&User{},
	}
}

func getTenancy(config config.Config) (multitenancy.Tenancy, error) {
	globalSet := flag.NewFlagSet("test", 0)
	globalSet.String("registry_address", "http://localhost:8500", "doc")
	cli := cli.NewContext(nil, globalSet, nil)
	client := providers.NewMicroClient(cli, opentracing.GlobalTracer())
	searchService, _ := search_datasource.NewSearchService(client, config)

	searchClient := search.NewSearchClient(searchService)

	tenancy := search_datasource.NewSearchTenancy(searchClient, &EntityMap{})

	return tenancy, nil
}

func getRepo() (repository.BaseRepository, error) {
	config, err := getConfig()
	if err != nil {
		return nil, err
	}

	tenancy, err := getTenancy(config)
	if err != nil {
		return nil, err
	}

	return NewBaseRepository(repository.NewMultitenancyProvider(tenancy)), nil
}

func TestCrud(t *testing.T) {
	assert := assert.New(t)

	ctx := context.Background()
	ctx = context.WithValue(ctx, "tenant-id", "tid27")

	userRepo, err := getRepo()
	if err != nil {
		t.Fatal(err)
	}

	name1 := "??????"
	age1 := 28
	now := time.Now()
	user1 := &User{
		ID:         "1",
		Name:       &name1,
		Age:        &age1,
		CreateTime: &now,
		UpdateTime: &now,
	}

	name2 := "??????"
	age2 := 21
	user2 := &User{
		ID:         "2",
		Name:       &name2,
		Age:        &age2,
		CreateTime: &now,
		UpdateTime: &now,
	}

	{
		err := userRepo.Create(ctx, user1)
		if err != nil {
			t.Fatal(err)
		}

		time.Sleep(time.Second * 3)

		err = userRepo.Create(ctx, user2)

		logger.Info("??????????????????")
	}

	name3 := "??????"
	age3 := 38
	user3 := &User{
		ID:   "3",
		Name: &name3,
		Age:  &age3,
	}
	{
		change, err := userRepo.Upsert(ctx, user3)
		assert.Error(err)
		t.Logf("change: %v", change)
	}

	{
		name4 := "??????"
		user4 := &User{
			ID:   user1.ID,
			Name: &name4,
		}
		err := userRepo.Update(ctx, user4, user4)
		if assert.Error(err) {
			t.Fatal(err)
		}

		data := map[string]interface{}{
			"name": "?????????",
			"age":  10,
		}

		err = userRepo.Update(ctx, &User{}, data)
		if assert.NoError(err) {
			t.Fatal(err)
		}

		// ????????????????????? age??????????????????, &User{ Name: "sunwukong", Age: 0}
		err = userRepo.Update(ctx, user2, data)
		if assert.Error(err) {
			t.Fatal(err)
		}
		logger.Info("??????????????????")
	}

	{
		findUser := &User{ID: user1.ID}
		err := userRepo.Get(ctx, findUser)
		if assert.Error(err) {
			t.Fatal(err)
		}

		b, _ := json.Marshal(findUser)

		logger.Infof("??????????????????: %v", string(b))
	}

	{
		pageQuery := &entity.PageQuery{
			Filter: map[string]interface{}{
				"name": "??????",
				"age": map[string]interface{}{
					"GT": 22,
				},
			},
			PageSize: 10,
			PageNo:   1,
		}

		items := make([]*User, 0)
		total, err := userRepo.Page(ctx, &User{}, pageQuery, &items)
		if assert.Error(err) {
			t.Fatal(err)
		}

		if assert.Equal(1, total) {
			logger.Info("??????????????????")
		} else {
			logger.Info(fmt.Sprintf("??????????????????, ??????1????????????????????????%d???", total))
		}

		b, _ := json.Marshal(items)
		s := string(b)
		t.Log(s)
	}

	testCursorList(t, userRepo)

	{
		err := userRepo.Delete(ctx, &User{ID: user1.ID})
		assert.NoError(err)
		logger.Info("??????????????????")

		items := make([]*User, 0)
		total, err := userRepo.Page(ctx, &User{}, &entity.PageQuery{
			Filter:   map[string]interface{}{},
			PageSize: 10,
			PageNo:   1,
		}, &items)
		if assert.Error(err) {
			t.Fatal(err)
		}

		assert.Equal(1, total)
		assert.Equal(1, len(items))

		err = userRepo.Delete(ctx, &User{ID: user2.ID})
		err = userRepo.Delete(ctx, &User{ID: user3.ID})
		logger.Info("??????????????????")

		items = make([]*User, 0)
		total, err = userRepo.Page(ctx, &User{}, &entity.PageQuery{
			Filter:   map[string]interface{}{},
			PageSize: 10,
			PageNo:   1,
		}, &items)
		assert.NoError(err)
		assert.Equal(t, 0, total)

		logger.Info("??????????????????")
	}
}

func TestCursorList(t *testing.T) {
	userRepo, err := getRepo()
	if err != nil {
		t.Fatal(err)
	}

	testCursorList(t, userRepo)
}

func testCursorList(t *testing.T, repo repository.BaseRepository) {
	ctx := context.Background()
	ctx = context.WithValue(ctx, "tenant-id", "tid27")

	for i := 0; i < 100; i++ {
		age := i + 10
		name := fmt.Sprintf("??????%d", i)
		user := &User{
			ID:   fmt.Sprintf("%d", i),
			Name: &name,
			Age:  &age,
		}

		_, err := repo.Upsert(ctx, user)
		if err != nil {
			t.Fatal(err)
		}
	}

	var cursor string
	for {
		cursorQuery := &entity.CursorQuery{
			NeedTotal: true,
			Cursor:    cursor,
			Fields:    []string{"name", "age"},
			Direction: entity.CursorDirectionAfter,
			Filter:    map[string]interface{}{
				// "name": "3",
			},
			// Orders: []*entity.Order{
			// {
			//	Field: "name",
			// },
			// {
			//	Field: "age",
			//},
			//},
			Size: 7,
		}

		items := make([]*User, 0)
		extra, err := repo.List(ctx, cursorQuery, &User{}, &items)
		if err != nil {
			t.Fatal(err)
		}

		b, _ := json.Marshal(items)
		s := string(b)
		log.Printf("=== items: %s", s)
		//log.Printf("=== extra: %v", extra)
		log.Printf("=== cursor: %v", cursor)
		log.Printf("=== items count: %d", len(items))
		log.Printf("=== total: %d", extra.Total)
		log.Printf("=== startCursor: %s", extra.StartCursor)
		log.Printf("=== endCursor: %s", extra.EndCursor)
		log.Printf("=== hasPrevious: %v", extra.HasPrevious)
		log.Printf("=== hasNext: %v", extra.HasNext)

		cursor = extra.EndCursor

		if !extra.HasPrevious || !extra.HasNext {
			break
		}
	}

	logger.Info("??????????????????")
}

func TestConnectionPaginate(t *testing.T) {
	userRepo, err := getRepo()
	if err != nil {
		t.Fatal(err)
	}

	testConnectionPaginate(t, userRepo)
}

func testConnectionPaginate(t *testing.T, repo repository.BaseRepository) {
	for i := 0; i < 5; i++ {
		name := "??????"
		age := 1 + i
		now := time.Now()
		user := &User{
			ID:         fmt.Sprintf("%d", i+1),
			Name:       &name,
			Age:        &age,
			CreateTime: &now,
			UpdateTime: &now,
		}

		// err := repo.Delete(context.Background(), user)
		_, err := repo.Upsert(context.Background(), user)
		if err != nil {
			t.Fatal(err)
		}
	}

	// h, _ := time.ParseDuration("1s")
	// t1 := user1.Ctime.Add(h)

	var before string
	var after string
	last := 3
	// first := 3
	round := 0
	for {
		round += 1

		connQuery := &entity.ConnectionQuery{
			NeedTotal: true,
			// After:     &after,
			Before: &before,
			Fields: []string{"name", "ctime"},
			Filter: map[string]interface{}{},
			/*Orders: []*entity.Order{
				{
					Field: "name",
				},
				{
					Field: "ctime",
				},
			},*/
			Last: &last,
			// First: &first,
		}

		conn, err := repo.Connection(context.Background(), connQuery, &User{})
		if err != nil {
			t.Fatal(err)
		}

		b, _ := json.Marshal(conn.Edges)
		s := string(b)
		log.Printf("==================== round: %d", round)
		log.Printf("=== edges: %s", s)
		//log.Printf("=== extra: %v", extra)
		log.Printf("=== after: %v, before: %v", after, before)
		log.Printf("=== edges count: %d", len(conn.Edges))
		log.Printf("=== total: %d", conn.Total)
		log.Printf("=== startCursor: %s", conn.PageInfo.StartCursor)
		log.Printf("=== endCursor: %s", conn.PageInfo.EndCursor)
		log.Printf("=== hasPrevious: %v", conn.PageInfo.HasPrevious)
		log.Printf("=== hasNext: %v", conn.PageInfo.HasNext)

		after = conn.PageInfo.EndCursor
		before = conn.PageInfo.StartCursor

		if !conn.PageInfo.HasPrevious || !conn.PageInfo.HasNext {
			break
		}
	}

	logger.Info("??????????????????")
}
