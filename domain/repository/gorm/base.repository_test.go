package gorm

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/duolacloud/microbase/datasource/gorm"
	"github.com/duolacloud/microbase/domain/entity"
	"github.com/duolacloud/microbase/domain/repository"
	"github.com/duolacloud/microbase/multitenancy"
	_gorm "github.com/jinzhu/gorm"
	"github.com/micro/go-micro/v2/config"
	"github.com/micro/go-micro/v2/config/source/memory"
	"github.com/micro/go-micro/v2/logger"
	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
)

type User struct {
	ID         string     `json:"id" gorm:"primary_key"`
	Name       string     `json:"name"`
	Age        int        `json:"age"`
	CreateTime time.Time  `json:"ctime" gorm:"column:ctime"`
	UpdateTime time.Time  `json:"mtime" gorm:"column:utime"`
	DeleteTime *time.Time `json:"dtime" gorm:"column:dtime"`
	Deleted    bool       `json:"deleted"`
}

func (u *User) BeforeCreate(scope *_gorm.Scope) error {
	scope.SetColumn("id", uuid.NewV4().String())
	return nil
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
		"db": {
			"driver": "mysql",
			"connection_string": "root:debezium@tcp(localhost:3306)/test?charset=utf8mb4&parseTime=True&loc=Local"
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
		User{},
	}
}

func getTenancy(config config.Config) (multitenancy.Tenancy, error) {
	entityMap := &EntityMap{}
	tenancy, err := gorm.NewGormTenancy(config, entityMap)
	if err != nil {
		logger.Fatal("数据库连接失败", err)
		return nil, err
	}

	return tenancy, nil
}

func TestCrud(t *testing.T) {
	assert := assert.New(t)

	config, err := getConfig()
	if err != nil {
		t.Fatal(err)
	}

	tenancy, err := getTenancy(config)
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	ctx = context.WithValue(ctx, "tenant-id", "tenantId1")

	userRepo := NewBaseRepository(repository.NewMultitenancyProvider(tenancy))

	user1 := &User{
		Name: "吕布",
		Age:  28,
	}

	user2 := &User{
		Name: "貂蝉",
		Age:  21,
	}

	{
		err := userRepo.Create(ctx, user1)
		if assert.Error(err) {
			t.Fatal(err)
		}

		time.Sleep(time.Second * 3)

		err = userRepo.Create(ctx, user2)
		if assert.Error(err) {
			t.Fatal(err)
		}

		logger.Info("插入记录成功")
	}

	user3 := &User{
		Name: "关羽",
		Age:  38,
	}
	{
		change, err := userRepo.Upsert(ctx, user3)
		assert.Error(err)
		t.Logf("change: %v", change)
	}

	{
		user4 := &User{
			ID:   user1.ID,
			Name: "赵云",
		}
		err := userRepo.Update(ctx, user4, user4)
		if assert.Error(err) {
			t.Fatal(err)
		}

		data := map[string]interface{}{
			"name": "孙悟空",
			"age":  0,
		}

		err = userRepo.Update(ctx, &User{}, data)
		if assert.NoError(err) {
			t.Fatal(err)
		}

		// 如果这么更新， age将不会被设置, &User{ Name: "sunwukong", Age: 0}
		err = userRepo.Update(ctx, user2, data)
		if assert.Error(err) {
			t.Fatal(err)
		}
		logger.Info("选择更新成功")
	}

	{
		findUser := &User{ID: user1.ID}
		err := userRepo.Get(ctx, findUser)
		if assert.Error(err) {
			t.Fatal(err)
		}
		logger.Info("找到对应记录")
	}

	{
		pageQuery := &entity.PageQuery{
			Filter: map[string]interface{}{
				"name": "赵云",
				"age": map[string]interface{}{
					"GT": 22,
				},
			},
			PageSize: 10,
			PageNo:   1,
		}

		items := make([]*User, 0)
		total, pageCount, err := userRepo.Page(ctx, &User{}, pageQuery, &items)
		if assert.Error(err) {
			t.Fatal(err)
		}

		if assert.Equal(1, total) {
			logger.Info("翻页查询正确")
		} else {
			logger.Info(fmt.Sprintf("翻页查询错误, 期望1条记录，实际返回%d条", total))
		}

		if assert.Equal(1, pageCount) {
			logger.Info("翻页查询正确")
		} else {
			logger.Info(fmt.Sprintf("翻页查询错误 期望1页, 实际返回%d页", pageCount))
		}

		b, _ := json.Marshal(items)
		s := string(b)
		t.Log(s)
	}

	testCursorList(t, userRepo)

	{
		err := userRepo.Delete(ctx, &User{ID: user1.ID})
		assert.NoError(err)
		logger.Info("删除记录成功")

		items := make([]*User, 0)
		total, pageCount, err := userRepo.Page(ctx, &User{}, &entity.PageQuery{
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
		logger.Info("删除记录成功")

		items = make([]*User, 0)
		total, pageCount, err = userRepo.Page(ctx, &User{}, &entity.PageQuery{
			Filter:   map[string]interface{}{},
			PageSize: 10,
			PageNo:   1,
		}, &items)
		assert.NoError(err)
		assert.Equal(t, 0, total)
		assert.Equal(t, 0, pageCount)

		logger.Info("翻页核对成功")
	}
}

func TestCursorList(t *testing.T) {
	config, err := getConfig()
	if err != nil {
		t.Fatal(err)
	}

	tenancy, err := getTenancy(config)
	if err != nil {
		t.Fatal(err)
	}

	userRepo := NewBaseRepository(repository.NewMultitenancyProvider(tenancy))

	testCursorList(t, userRepo)
}

func testCursorList(t *testing.T, repo repository.BaseRepository) {
	/*for i := 0; i < 100; i++ {
		user := &User{
			Name: "关羽",
			Age:  38,
		}

		change, _ := repo.Upsert(context.Background(), user)
		t.Logf("change: %v", change)
		time.Sleep(time.Second * 2)
	}*/
	// h, _ := time.ParseDuration("1s")
	// t1 := user1.Ctime.Add(h)

	var cursor string
	for {
		cursorQuery := &entity.CursorQuery{
			NeedTotal: true,
			Cursor:    cursor,
			Fields:    []string{"name", "age"},
			Direction: entity.CursorDirectionBefore,
			Filter:    map[string]interface{}{},
			Orders: []*entity.Order{
				{
					Field: "name",
				},
				{
					Field: "ctime",
				},
			},
			Size: 7,
		}

		items := make([]*User, 0)
		extra, err := repo.List(context.Background(), cursorQuery, &User{}, &items)
		if err != nil {
			t.Fatal(err)
		}

		// b, _ := json.Marshal(items)
		// s := string(b)
		// log.Printf("=== items: %s", s)
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

	logger.Info("游标查询成功")
}

func TestConnectionPaginate(t *testing.T) {
	config, err := getConfig()
	if err != nil {
		t.Fatal(err)
	}

	tenancy, err := getTenancy(config)
	if err != nil {
		t.Fatal(err)
	}

	userRepo := NewBaseRepository(repository.NewMultitenancyProvider(tenancy))

	testConnectionPaginate(t, userRepo)
}

func testConnectionPaginate(t *testing.T, repo repository.BaseRepository) {
	/*for i := 0; i < 100; i++ {
		user := &User{
			Name: "关羽",
			Age:  38,
		}

		change, _ := repo.Upsert(context.Background(), user)
		t.Logf("change: %v", change)
		time.Sleep(time.Second * 2)
	}*/
	// h, _ := time.ParseDuration("1s")
	// t1 := user1.Ctime.Add(h)

	var after string
	first := 7
	for {
		connQuery := &entity.ConnectionQuery{
			NeedTotal: true,
			After:     &after,
			Fields:    []string{"name", "ctime"},
			Filter:    map[string]interface{}{},
			Orders: []*entity.Order{
				{
					Field: "name",
				},
				{
					Field: "ctime",
				},
			},
			First: &first,
		}

		conn, err := repo.Connection(context.Background(), connQuery, &User{})
		if err != nil {
			t.Fatal(err)
		}

		b, _ := json.Marshal(conn.Edges)
		s := string(b)
		log.Printf("=== edges: %s", s)
		//log.Printf("=== extra: %v", extra)
		log.Printf("=== after: %v", after)
		log.Printf("=== edges count: %d", len(conn.Edges))
		log.Printf("=== total: %d", conn.Total)
		log.Printf("=== startCursor: %s", conn.PageInfo.StartCursor)
		log.Printf("=== endCursor: %s", conn.PageInfo.EndCursor)
		log.Printf("=== hasPrevious: %v", conn.PageInfo.HasPrevious)
		log.Printf("=== hasNext: %v", conn.PageInfo.HasNext)

		after = conn.PageInfo.EndCursor

		if !conn.PageInfo.HasPrevious || !conn.PageInfo.HasNext {
			break
		}
	}

	logger.Info("游标查询成功")
}
