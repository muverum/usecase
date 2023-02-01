package decoders

import (
	"github.com/swaggest/form/v5"
	"github.com/swaggest/rest"
	"github.com/swaggest/rest/request"
	"net/http"
)

// decoder extracts Go value from *http.Request.
type JsonDecoder struct {
	decoders []ValueDecoderFunc
	in       []rest.ParamIn
}

// Decode populates and validates input with data from http request.
func (d *JsonDecoder) Decode(r *http.Request, input interface{}, validator rest.Validator) error {
	if i, ok := input.(request.Loader); ok {
		return i.LoadFromHTTPRequest(r)
	}

	for i, decode := range d.decoders {
		err := decode(r, input, validator)
		if err != nil {
			// nolint:errorlint // Error is not wrapped, type assertion is more performant.
			if de, ok := err.(form.DecodeErrors); ok {
				errs := make(rest.RequestErrors, len(de))
				for name, e := range de {
					errs[string(d.in[i])+":"+name] = []string{"#: " + e.Error()}
				}

				return errs
			}

			return err
		}
	}

	return nil
}
