package decoders

import (
	"github.com/swaggest/rest"
	"net/http"
	"net/url"
)

type (
	DecoderFunc      func(r *http.Request) (url.Values, error)
	ValueDecoderFunc func(r *http.Request, v interface{}, validator rest.Validator) error
)
