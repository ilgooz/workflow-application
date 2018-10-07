package workflow

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParser(t *testing.T) {
	expectations := []struct {
		parser        *parser
		value         interface{}
		expectedValue interface{}
		err           error
	}{
		{
			&parser{},
			"1",
			"1",
			nil,
		},
		{
			&parser{},
			1,
			1,
			nil,
		},
		{
			&parser{
				configs: []ConfigDefinition{
					{Key: "a", Value: 1},
				},
			},
			"$configs.a",
			1,
			nil,
		},
		{
			&parser{
				configs: []ConfigDefinition{
					{Key: "a", Value: map[string]interface{}{"b": 1}},
				},
			},
			"$configs.a.b",
			1,
			nil,
		},
		{
			&parser{
				configs: []ConfigDefinition{
					{Key: "a", Value: 1},
				},
			},
			"$configs.b",
			nil,
			&invalidVarErr{variable: "$configs.b"},
		},
		{
			&parser{
				data: map[string]interface{}{
					"a": 1,
				},
			},
			"$data.a",
			1,
			nil,
		},
		{
			&parser{
				data: map[string]interface{}{
					"a": map[string]interface{}{"b": 1},
				},
			},
			"$data.a.b",
			1,
			nil,
		},
		{
			&parser{
				services: []ServiceDefinition{
					{Name: "x", ID: "a"},
				},
			},
			"$services.x",
			"a",
			nil,
		},
	}

	for _, expectation := range expectations {
		expectedValue, err := expectation.parser.Parse(expectation.value)
		require.Equal(t, expectation.err, err)
		require.Equal(t, expectation.expectedValue, expectedValue)
	}
}
