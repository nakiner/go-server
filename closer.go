package server

import "sync"

type closer struct {
	lock    sync.Mutex
	closers []func() error
}

func (a *closer) add(c ...func() error) {
	a.lock.Lock()
	defer a.lock.Unlock()
	a.closers = append(c, a.closers...)
}

func (a *closer) actor() (func() error, func(error)) {
	var runWg sync.WaitGroup
	runWg.Add(1)
	return func() error {
			runWg.Wait()
			return nil
		}, func(err error) {
			runWg.Done()

			a.lock.Lock()
			defer a.lock.Unlock()
			var wg sync.WaitGroup
			wg.Add(len(a.closers))
			for _, c := range a.closers {
				closer := c
				go func() {
					defer wg.Done()
					if err := closer(); err != nil {
						appLogger.Errorf("closer: %s", err)
					}
				}()
			}
			wg.Wait()
		}
}
