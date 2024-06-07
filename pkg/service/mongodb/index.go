package mongodb

import "context"

func (s *Service) Indexes(ctx context.Context, e Entity) error {
	col, err := s.Collection(e)
	if err != nil {
		s.log.From(Component, ActionIndex).Error(err.Error())
		return err
	}
	s.log.From(Component, ActionIndex).Debug("creating indexes", "collection", col.Name())

	for _, idx := range e.Indexes() {
		name, err := col.Indexes().CreateOne(ctx, idx)
		if err != nil {
			s.log.From(Component, ActionIndex).Error(err.Error())
			continue
		}
		s.log.From(Component, ActionIndex).Debug("index created", "name", name)
	}

	return nil
}
