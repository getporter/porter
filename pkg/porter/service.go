package porter

type ServiceOptions struct {
	Port        int64
	ServiceName string
}

func (o *ServiceOptions) Validate() error {
	return nil
}
