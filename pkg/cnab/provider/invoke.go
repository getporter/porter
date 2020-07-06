package cnabprovider

func (r *Runtime) Invoke(action string, args ActionArguments) error {
	return r.ExecuteAction(action, args)
}
