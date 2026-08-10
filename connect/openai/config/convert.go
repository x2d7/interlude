package config

import "github.com/openai/openai-go/v3/packages/param"

func optFloat64(v *float64) param.Opt[float64] {
	if v == nil {
		return param.Opt[float64]{}
	}
	return param.NewOpt(*v)
}

func optInt64(v *int64) param.Opt[int64] {
	if v == nil {
		return param.Opt[int64]{}
	}
	return param.NewOpt(*v)
}

func optBool(v *bool) param.Opt[bool] {
	if v == nil {
		return param.Opt[bool]{}
	}
	return param.NewOpt(*v)
}

func optString(v *string) param.Opt[string] {
	if v == nil {
		return param.Opt[string]{}
	}
	return param.NewOpt(*v)
}
