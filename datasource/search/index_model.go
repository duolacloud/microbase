package search

import (
	"context"
	"encoding/json"
	"fmt"
	"go/ast"
	"log"
	"reflect"
	"strings"

	_search "github.com/duolacloud/microbase/client/search"
	_reflect "github.com/duolacloud/microbase/reflect"
)

type IndexModel struct {
	searchClient _search.SearchClient
	model        interface{}
	tenantId     string
}

func NewIndexModel(searchClient _search.SearchClient, model interface{}, tenantId string) *IndexModel {
	return &IndexModel{
		searchClient,
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
			jsonTag := fieldStruct.Tag.Get("json")
			ns := strings.Split(jsonTag, ",")
			var name string
			if len(ns) == 0 {
				name = strings.ToLower(fieldStruct.Name)
			} else {
				name = ns[0]
			}

			properties[name] = settings
		}
	}

	var s, _ = json.MarshalIndent(properties, "", "\t")
	log.Printf("%v", string(s))

	structInfo, err := _reflect.GetStructInfo(m.model, nil)
	if err != nil {
		return err
	}

	typ := _reflect.TheNamingStrategy.Table(structInfo.Name)
	indexName := indexName(typ, m.tenantId)
	// Check if index exists
	indexExists, err := m.searchClient.IndexExists(c, indexName)
	if err != nil {
		return err
	}

	if !indexExists {
		mapping := map[string]interface{}{
			"mappings": map[string]interface{}{
				typ: map[string]interface{}{
					"properties": properties,
				},
			},
		}

		err = m.searchClient.CreateIndex(c, &_search.Index{
			Name:    indexName,
			Mapping: mapping,
		})

		if err != nil {
			return err
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
