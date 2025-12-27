package github

import (
	"bytes"
	"encoding/json"
	"errors"
)

type GraphqlError struct {
	Query     string `json:"-"`
	Message   string `json:"message"`
	Locations []struct {
		Line   int `json:"line"`
		Column int `json:"column"`
	} `json:"locations"`
}

func (e *GraphqlError) Error() string {
	return "github: graphql: " + e.Message
}

func GraphqlRequest(query string, variables map[string]any, out any) error {
	buf := &bytes.Buffer{}
	if err := json.NewEncoder(buf).Encode(map[string]any{
		"query":     query,
		"variables": variables,
	}); err != nil {
		return err
	}

	headers := map[string]string{
		"Content-Type": "application/json",
	}
	body, err := Request("POST", "graphql", headers, buf)
	if err != nil {
		return err
	}
	defer body.Close()

	o := struct {
		Data   any             `json:"data"`
		Errors []*GraphqlError `json:"errors"`
	}{
		Data: &out,
	}
	if err := json.NewDecoder(body).Decode(&o); err != nil {
		return err
	}

	if len(o.Errors) > 0 {
		errs := []error{}
		for _, err := range o.Errors {
			err.Query = query
			errs = append(errs, err)
		}
		return errors.Join(errs...)
	}
	return nil
}
