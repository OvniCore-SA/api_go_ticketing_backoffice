package producer_config

import (
	"context"
	"errors"
	"testing"

	"github.com/OvniCore-SA/api_go_ticketing_backoffice/pkg/entities"
	"github.com/OvniCore-SA/api_go_ticketing_backoffice/pkg/filters"
)

// ─── Fakes ────────────────────────────────────────────────────────────────────

type fakeProducersRepo struct {
	producer entities.Producer
	err      error
}

func (r *fakeProducersRepo) CreateProducerRepository(entity entities.Producer) (entities.Producer, error) {
	return entity, nil
}
func (r *fakeProducersRepo) GetAllProducersRepository(filter filters.ProducerFilter, page, limit int) ([]entities.Producer, int64, error) {
	return nil, 0, nil
}
func (r *fakeProducersRepo) GetProducerRepository(filter filters.ProducerFilter) (entities.Producer, error) {
	if r.err != nil {
		return entities.Producer{}, r.err
	}
	return r.producer, nil
}
func (r *fakeProducersRepo) UpdateProducerRepository(id uint, fields map[string]interface{}) error {
	return nil
}
func (r *fakeProducersRepo) DeleteProducerRepository(id uint) error { return nil }
func (r *fakeProducersRepo) SeedProducerFromTemplateRepository(entity entities.Producer, templateID *uint) (entities.Producer, error) {
	return entity, nil
}

// spyLayer registra el orden de aplicación y adiciona su nombre a un contador
// dentro de EffectiveConfig.Meta.AppliedLayers via el Resolver.
type spyLayer struct {
	name   string
	err    error
	mutate func(cfg EffectiveConfig) EffectiveConfig
}

func (l *spyLayer) Name() string { return l.name }
func (l *spyLayer) Apply(_ context.Context, _ ResolveContext, cfg EffectiveConfig) (EffectiveConfig, error) {
	if l.err != nil {
		return cfg, l.err
	}
	if l.mutate != nil {
		cfg = l.mutate(cfg)
	}
	return cfg, nil
}

// ─── Tests ────────────────────────────────────────────────────────────────────

func TestResolver_AppliesLayersInOrder(t *testing.T) {
	repo := &fakeProducersRepo{producer: entities.Producer{
		Name:   "FiestaBresh",
		Slug:   "fiestabresh",
		Status: entities.ProducerStatusActive,
	}}
	repo.producer.ID = 42

	resolver := NewResolver(repo,
		&spyLayer{name: "one"},
		&spyLayer{name: "two"},
		&spyLayer{name: "three"},
	)

	cfg, err := resolver.Resolve(context.Background(), 42)
	if err != nil {
		t.Fatalf("Resolve returned unexpected error: %v", err)
	}

	want := []string{"one", "two", "three"}
	if len(cfg.Meta.AppliedLayers) != len(want) {
		t.Fatalf("expected %d layers applied, got %d (%v)", len(want), len(cfg.Meta.AppliedLayers), cfg.Meta.AppliedLayers)
	}
	for i, name := range want {
		if cfg.Meta.AppliedLayers[i] != name {
			t.Errorf("layer %d: want %q, got %q", i, name, cfg.Meta.AppliedLayers[i])
		}
	}

	if cfg.Meta.ProducerID != 42 {
		t.Errorf("Meta.ProducerID = %d, want 42", cfg.Meta.ProducerID)
	}
	if cfg.Meta.Suspended {
		t.Errorf("Meta.Suspended = true, want false (producer active)")
	}
}

func TestResolver_StatusLayerMarksSuspended(t *testing.T) {
	repo := &fakeProducersRepo{producer: entities.Producer{
		Name:   "Pausada",
		Slug:   "pausada",
		Status: entities.ProducerStatusSuspended,
	}}
	repo.producer.ID = 7

	resolver := NewResolver(repo, NewStatusLayer())

	cfg, err := resolver.Resolve(context.Background(), 7)
	if err != nil {
		t.Fatalf("Resolve returned unexpected error: %v", err)
	}
	if !cfg.Meta.Suspended {
		t.Errorf("Meta.Suspended = false, want true")
	}
	if cfg.Meta.Status != entities.ProducerStatusSuspended {
		t.Errorf("Meta.Status = %q, want suspended", cfg.Meta.Status)
	}
}

func TestResolver_LayerErrorSurfaces(t *testing.T) {
	repo := &fakeProducersRepo{producer: entities.Producer{Status: entities.ProducerStatusActive}}
	repo.producer.ID = 1

	boom := errors.New("db explota")
	resolver := NewResolver(repo,
		&spyLayer{name: "ok"},
		&spyLayer{name: "boom", err: boom},
		&spyLayer{name: "no-corre"}, // no debería ejecutarse
	)

	_, err := resolver.Resolve(context.Background(), 1)
	if err == nil {
		t.Fatalf("expected error from layer 'boom', got nil")
	}
	// El error viene como fiber.NewError con mensaje que incluye "boom".
	if got := err.Error(); got == "" || (contains(got, "boom") == false) {
		t.Errorf("error message should contain layer name 'boom', got %q", got)
	}
}

func TestResolver_ProducerNotFound(t *testing.T) {
	repo := &fakeProducersRepo{err: errNotFound{}}

	resolver := NewResolver(repo)
	_, err := resolver.Resolve(context.Background(), 999)
	if err == nil {
		t.Fatal("expected error for missing producer, got nil")
	}
}

// contains es un pequeño helper para evitar importar strings en tests.
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (indexOf(s, substr) >= 0)
}
func indexOf(s, substr string) int {
	for i := 0; i+len(substr) <= len(s); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

// errNotFound simula gorm.ErrRecordNotFound sin importar gorm en el test.
type errNotFound struct{}

func (errNotFound) Error() string { return "record not found" }
func (errNotFound) Is(target error) bool {
	return target != nil && target.Error() == "record not found"
}
