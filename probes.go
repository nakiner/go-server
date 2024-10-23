package server

import (
	"encoding/json"
	"net/http"
	"time"
)

type probes struct {
	healthy      map[string]string
	ready        map[string]string
	healthChecks map[string]func() error
	readyChecks  map[string]func() error
}

type component struct {
	Name  string `json:"name,omitempty"`
	Error string `json:"error,omitempty"`
}

func (a *App) initProbes() {
	a.debugRouter.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		unhealthy := make([]component, 0, len(a.probes.healthy))
		for name, errMsg := range a.probes.healthy {
			if errMsg != "" {
				unhealthy = append(unhealthy, component{
					Name:  name,
					Error: errMsg,
				})
			}
		}
		if len(unhealthy) > 0 {
			resp, err := json.Marshal(unhealthy)
			if err != nil {
				_, _ = w.Write([]byte("COMPONENTS MARSHAL ERROR"))
				w.WriteHeader(http.StatusInternalServerError)
			}
			_, _ = w.Write(resp)
			w.WriteHeader(http.StatusInternalServerError)
		}
		_, _ = w.Write([]byte("OK"))
		w.WriteHeader(http.StatusOK)
	})

	a.debugRouter.HandleFunc("/readyz", func(w http.ResponseWriter, r *http.Request) {
		unready := make([]component, 0, len(a.probes.ready))
		for name, errMsg := range a.probes.ready {
			if errMsg != "" {
				unready = append(unready, component{
					Name:  name,
					Error: errMsg,
				})
			}
		}
		if len(unready) > 0 {
			resp, err := json.Marshal(unready)
			if err != nil {
				_, _ = w.Write([]byte("COMPONENTS MARSHAL ERROR"))
				w.WriteHeader(http.StatusInternalServerError)
			}
			_, _ = w.Write(resp)
			w.WriteHeader(http.StatusInternalServerError)
		}
		_, _ = w.Write([]byte("OK"))
		w.WriteHeader(http.StatusOK)
	})
}

func (a *App) AddHeathCheck(name string, fn func() error) {
	if a.probes.healthChecks == nil {
		a.probes.healthy = make(map[string]string)
		a.probes.healthChecks = make(map[string]func() error)
	}
	a.probes.healthChecks[name] = fn
}

func (a *App) AddReadinessCheck(name string, fn func() error) {
	if a.probes.readyChecks == nil {
		a.probes.ready = make(map[string]string)
		a.probes.readyChecks = make(map[string]func() error)
	}
	a.probes.readyChecks[name] = fn
}

func (p *probes) probesActor() (func() error, func(error)) {
	t := time.NewTicker(time.Second)

	return func() error {
			for range t.C {
				go func() {
					for name, fn := range p.readyChecks {
						if err := fn(); err != nil {
							p.ready[name] = err.Error()
						} else {
							delete(p.ready, name)
						}
					}
				}()
				go func() {
					for name, fn := range p.healthChecks {
						if err := fn(); err != nil {
							p.healthy[name] = err.Error()
						} else {
							delete(p.healthy, name)
						}
					}
				}()
			}
			return nil
		}, func(error) {
			t.Stop()
		}
}
