package handler

import (
  "encoding/json"
  "fmt"
  "github.com/gin-gonic/gin"
  "github.com/stretchr/testify/assert"
  "github.com/stretchr/testify/require"
  "math"
  "net/http"
  "testing"
)

// marshal marshals a value to JSON and logs a fatal error if marshaling fails.
// Only intended for use in testing.
func marshal(t *testing.T, value any) []byte {
  var data, err = json.Marshal(value)
  if nil != err {
    t.Fatal(err)
  }
  return data
}

type dummy struct {
  StringField  string  `json:"string_field,omitempty"`
  IntField     int     `json:"int_field"`
  Int8Field    int8    `json:"int8_field"`
  Int16Field   int16   `json:"int16_field"`
  Int32Field   int32   `json:"int32_field"`
  Int64Field   int64   `json:"int64_field"`
  Float32Field float32 `json:"float32_field"`
  Float64Field float64 `json:"float64_field"`
  BoolField    bool    `json:"bool_field,omitempty"`
  unexported   string
  NotWanted    string
}

func wrap(s string) string {
  return "  \t\n\n\t  " + s + "  \t\n\n\t  "
}

func Test_bindPostForm(t *testing.T) {
  t.Run("success", func(t *testing.T) {
    expectedStruct := dummy{
      StringField:  "Lorem ipsum dolor sit amet, consectetur adipiscing elit. Aenean tincidunt suscipit nibh, eu facilisis tortor imperdiet ut.",
      IntField:     math.MaxInt32,
      Int8Field:    math.MaxInt8,
      Int16Field:   math.MaxInt16,
      Int32Field:   math.MaxInt32,
      Int64Field:   math.MaxInt64,
      Float32Field: math.MaxFloat32,
      Float64Field: math.MaxFloat64,
      BoolField:    true,
    }

    c, _ := gin.CreateTestContext(nil)
    c.Request = &http.Request{}
    _ = c.Request.ParseForm()

    c.Request.PostForm.Add("string_field", wrap(expectedStruct.StringField))
    c.Request.PostForm.Add("int_field", wrap(fmt.Sprintf("%d", math.MaxInt32)))
    c.Request.PostForm.Add("int8_field", wrap(fmt.Sprintf("%d", math.MaxInt8)))
    c.Request.PostForm.Add("int16_field", wrap(fmt.Sprintf("%d", math.MaxInt16)))
    c.Request.PostForm.Add("int32_field", wrap(fmt.Sprintf("%d", math.MaxInt32)))
    c.Request.PostForm.Add("int64_field", wrap(fmt.Sprintf("%d", math.MaxInt64)))
    c.Request.PostForm.Add("float32_field", wrap(fmt.Sprintf("%f", math.MaxFloat32)))
    c.Request.PostForm.Add("float64_field", wrap(fmt.Sprintf("%f", math.MaxFloat64)))
    c.Request.PostForm.Add("bool_field", wrap("true"))
    c.Request.PostForm.Add("ignored_field", wrap("foo bar"))

    s := dummy{}
    err := bindPostForm(c, &s)

    assert.NoError(t, err)
    assert.Equal(t, expectedStruct, s)
  })

  t.Run("accepts no nil parameters", func(t *testing.T) {
    err1 := bindPostForm(nil, struct{}{})
    err2 := bindPostForm(&gin.Context{}, nil)
    msg := "got an unacceptable nil parameter"
    assert.ErrorContains(t, err1, msg)
    assert.ErrorContains(t, err2, msg)
  })

  t.Run("accepts only a pointer to a struct", func(t *testing.T) {
    values := []any{
      1,
      1.1,
      "str",
      func() {},
      make(chan int),
      struct{}{},
    }

    c := &gin.Context{}
    for _, value := range values {
      err := bindPostForm(c, value)
      assert.ErrorContains(t, err, "type of parameter \"val\" is not a pointer to a struct")
    }

    err := bindPostForm(c, &struct{}{})
    assert.NoError(t, err)
  })

  t.Run("invalid syntax", func(t *testing.T) {
    errs := []map[string]string{
      {
        "field_name":  "int_field",
        "field_value": "int",
      },
      {
        "field_name":  "int8_field",
        "field_value": "int8",
      },
      {
        "field_name":  "int16_field",
        "field_value": "int16",
      },
      {
        "field_name":  "int32_field",
        "field_value": "int32",
      },
      {
        "field_name":  "int64_field",
        "field_value": "int64",
      },
      {
        "field_name":  "float32_field",
        "field_value": "float32",
      },
      {
        "field_name":  "float64_field",
        "field_value": "float64",
      },
      {
        "field_name":  "bool_field",
        "field_value": "bool",
      },
    }

    for _, e := range errs {
      c, _ := gin.CreateTestContext(nil)
      c.Request = &http.Request{}
      _ = c.Request.ParseForm()
      c.Request.PostForm.Add(e["field_name"], wrap("foo"))
      s := dummy{}
      err := bindPostForm(c, &s)
      require.Error(t, err)
      assert.ErrorContains(t, err, fmt.Sprintf("Failed to parse the provided value as: %s. Please make sure the value is valid according to its type.", e["field_value"]))
    }
  })

  t.Run("out of range", func(t *testing.T) {
    errs := []map[string]string{
      {
        "field_name": "int8_field",
        "field_type": fmt.Sprintf("%d", math.MaxInt64),
      },
      {
        "field_name": "int16_field",
        "field_type": fmt.Sprintf("%d", math.MaxInt64),
      },
      {
        "field_name": "int32_field",
        "field_type": fmt.Sprintf("%d", math.MaxInt64),
      },
    }

    for _, e := range errs {
      c, _ := gin.CreateTestContext(nil)
      c.Request = &http.Request{}
      _ = c.Request.ParseForm()
      c.Request.PostForm.Add(e["field_name"], e["field_type"])
      s := dummy{}
      err := bindPostForm(c, &s)
      require.Error(t, err)
      assert.ErrorContains(t, err, "is out of range for the specified type")
    }
  })
}

func Test_getArticleFilter(t *testing.T) {
  // t.Run("success with search", func(t *testing.T) {
  //   expectedArticles := make([]*model.Article, 3)
  //   expectedNeedle := "20 www xxx yyy zzz zzz"
  //
  //   needle := ">> = 20 www? xxx! yyy... zzz_zzz \" ' ° <<"
  //
  //   r := mocks.NewArchiveRepository()
  //   r.On(routine, ctx, expectedNeedle, true, false).Return(expectedArticles, nil)
  //
  //   articles, err := NewArticlesService(r).GetHidden(ctx, needle)
  //
  //   assert.Equal(t, expectedArticles, articles)
  //   assert.NoError(t, err)
  // })

}
