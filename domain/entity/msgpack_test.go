package entity_test

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/vmihailenco/msgpack"
)

type Foo struct {
	ID    interface{} `msgpack:"i"`
	Value interface{} `msgpack:"v"`
}

func (c *Foo) Marshal(w io.Writer) error {
	wc := base64.NewEncoder(base64.RawStdEncoding, w)
	defer wc.Close()
	return msgpack.NewEncoder(wc).Encode(c)
}

func (c *Foo) Unmarshal(s string) error {
	if err := msgpack.NewDecoder(
		base64.NewDecoder(
			base64.RawStdEncoding,
			strings.NewReader(s),
		),
	).Decode(c); err != nil {
		return fmt.Errorf("cannot decode cursor: %w", err)
	}
	return nil
}

func TestMsgPack(t *testing.T) {
	vv := reflect.ValueOf(time.Now())
	c := &Foo{
		ID:    "37433482-5a06-4db5-91c6-591b1d15f6af",
		Value: vv.Interface(),
	}

	w := new(bytes.Buffer)

	err := c.Marshal(w)
	if err != nil {
		t.Fatal(err)
	}
	s := w.String()

	log.Printf("s: %s", s)

	var c1 Foo
	err = c1.Unmarshal(s)
	if err != nil {
		t.Fatal(err)
	}
	var v = c1.Value
	t.Logf("c1: %v, v: %v", c1, v)
}
