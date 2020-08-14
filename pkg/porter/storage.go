package porter

func (p *Porter) MigrateStorage() error {
	_, err := p.Storage.Migrate()
	return err
}
