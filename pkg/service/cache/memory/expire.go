package memory

import "time"

func (s *Service) StartExpirationController() {
	go func() {
		for s.ctx.Err() == nil {
			time.Sleep(200 * time.Millisecond)

			if s.expirations.IsEmpty() {
				continue
			}

			for _, k := range s.expirations.Keys() {
				expirationAt, ok := s.expirations.Get(k)
				if !ok {
					continue
				}

				if expirationAt.IsZero() {
					s.expirations.Remove(k)
				}

				if expirationAt.Before(time.Now()) {
					s.store.Remove(k)
					s.expirations.Remove(k)
				}
			}
		}
	}()
}
