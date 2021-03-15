package elasticsearch

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"go/ast"
	"log"
	"reflect"
	"strings"

	_reflect "github.com/duolacloud/microbase/reflect"
	"github.com/olivere/elastic/v6"
)

type IndexModel struct {
	client   *elastic.Client
	model    interface{}
	tenantId string
}

func NewIndexModel(client *elastic.Client, model interface{}, tenantId string) *IndexModel {
	return &IndexModel{
		client,
		model,
		tenantId,
	}
}

func (m *IndexModel) CreateIndex(c context.Context) error {
	reflectType := reflect.ValueOf(m.model).Elem().Type()

	properties := map[string]interface{}{}

	// Get all fields
	for i := 0; i < reflectType.NumField(); i++ {
		if fieldStruct := reflectType.Field(i); ast.IsExported(fieldStruct.Name) {
			settings := parseTagSetting(fieldStruct.Tag)
			name := fieldStruct.Tag.Get("json")
			if len(name) == 0 {
				name = strings.ToLower(fieldStruct.Name)
			}

			properties[name] = settings
		}
	}

	indexMapping := map[string]interface{}{
		"properties": properties,
	}

	var s, _ = json.MarshalIndent(indexMapping, "", "\t")
	log.Printf("%v", string(s))

	structInfo, err := _reflect.GetStructInfo(m.model, nil)
	if err != nil {
		return err
	}

	indexName := indexName(_reflect.TheNamingStrategy.Table(structInfo.Name), m.tenantId)
	// Check if index exists
	indexExists, err := m.client.IndexExists(indexName).Do(c)
	if err != nil {
		return err
	}

	if !indexExists {
		r, err := m.client.CreateIndex(indexName).BodyJson(indexMapping).Do(c)
		if err != nil {
			return err
		}

		if !r.Acknowledged {
			return errors.New(fmt.Sprintf("expected IndicesCreateResult.Acknowledged true; got %v", r.Acknowledged))
		}
	} else {
		// TODO migrate 更新索引？
	}

	return nil
}

func indexName(entityName, tenantId string) string {
	if len(tenantId) == 0 {
		return entityName
	}

	return fmt.Sprintf("%s_%s", entityName, tenantId)
}

func parseTagSetting(tags reflect.StructTag) map[string]string {
	setting := map[string]string{}
	for _, str := range []string{tags.Get("index"), tags.Get("elastic")} {
		if str == "" {
			continue
		}
		tags := strings.Split(str, ";")
		for _, value := range tags {
			v := strings.Split(value, ":")
			k := strings.TrimSpace(strings.ToLower(v[0]))
			if len(v) >= 2 {
				setting[k] = strings.Join(v[1:], ":")
			} else {
				setting[k] = k
			}
		}
	}
	return setting
}
