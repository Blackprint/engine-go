package engine

type Config struct{}

type Option interface {
	Apply(cfg *Config) (err error)
}

//

func scaffoldOptions(opts []Option) (c Config, err error) {
	c = Config{}
	for _, opt := range opts {
		if err = opt.Apply(&c); err != nil {
			return
		}
	}
	return
}

//
