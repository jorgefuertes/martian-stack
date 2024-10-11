package memory

import "time"

func (s *Service) StartExpirationController() {
	go func() {
		for {
			time.Sleep(200 * time.Millisecond)

			if s.expirations == nil {
				return
			}

			if len(s.expirations) == 0 {
				continue
			}

			if s.lock.TryLock() {
				for key, exp := range s.expirations {
					if exp.IsZero() {
						delete(s.expirations, key)
					} else if exp.Before(time.Now()) {
						delete(s.expirations, key)
						delete(s.store, key)
					}
				}
				s.lock.Unlock()
			}
		}
	}()
}
