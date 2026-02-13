package tools

type AddOption func(*toolAddConfig)

type toolAddConfig struct {
	overrideName string

	autoIncrement  bool
	startIncrement int
}

func WithOverrideName(name string) AddOption {
	return func(c *toolAddConfig) {
		c.overrideName = name
	}
}

func WithAutoIncrement() AddOption {
	return func(c *toolAddConfig) {
		c.autoIncrement = true
	}
}

func WithStartIncrement(start int) AddOption {
	return func(c *toolAddConfig) {
		c.startIncrement = start
	}
}