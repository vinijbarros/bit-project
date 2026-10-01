package seed

import (
	"strings"
	"testing"
)

func TestLoadConfigFromEnv(t *testing.T) {
	t.Setenv("DEMO_SEED_ENABLED", "true")
	t.Setenv("DEMO_USER1_USERNAME", "colaborador1")
	t.Setenv("DEMO_USER1_DISPLAY_NAME", "Colaborador 1")
	t.Setenv("DEMO_USER1_PASSWORD", "SenhaLocal#1")
	t.Setenv("DEMO_USER2_USERNAME", "colaborador2")
	t.Setenv("DEMO_USER2_DISPLAY_NAME", "Colaborador 2")
	t.Setenv("DEMO_USER2_PASSWORD", "SenhaLocal#2")

	cfg, err := LoadConfigFromEnv()
	if err != nil {
		t.Fatalf("LoadConfigFromEnv() error = %v", err)
	}
	if !cfg.Enabled || cfg.Users[0].Username != "colaborador1" || cfg.Users[1].Username != "colaborador2" {
		t.Fatalf("unexpected config: %+v", cfg)
	}
}

func TestLoadConfigFromEnvRequiresExplicitEnable(t *testing.T) {
	t.Setenv("DEMO_SEED_ENABLED", "false")
	if _, err := LoadConfigFromEnv(); err == nil {
		t.Fatal("LoadConfigFromEnv() expected an error")
	}
}

func TestValidateUserPasswordByteLimit(t *testing.T) {
	base := User{Username: "colaborador1", DisplayName: "Colaborador 1"}

	base.Password = strings.Repeat("a", 72)
	if err := validateUser(base, "DEMO_USER1_"); err != nil {
		t.Fatalf("72-byte password should be accepted: %v", err)
	}

	base.Password = strings.Repeat("á", 36) + "a"
	if err := validateUser(base, "DEMO_USER1_"); err == nil {
		t.Fatal("73-byte password should be rejected")
	}
}

func TestDemoRequestsCoverRequiredValues(t *testing.T) {
	categories := map[string]bool{}
	statuses := map[string]bool{}
	requesters := map[int]bool{}
	ages := map[int64]bool{}
	keys := map[string]bool{}

	for _, request := range demoRequests {
		categories[request.Category] = true
		statuses[request.Status] = true
		requesters[request.UserIndex] = true
		ages[int64(request.Age)] = true
		if keys[request.SeedKey] {
			t.Fatalf("duplicate seed key %q", request.SeedKey)
		}
		keys[request.SeedKey] = true
	}

	for _, category := range []string{"ti", "rh", "compras", "financeiro", "infraestrutura"} {
		if !categories[category] {
			t.Errorf("missing category %q", category)
		}
	}
	for _, status := range []string{"aberto", "em_atendimento", "concluido"} {
		if !statuses[status] {
			t.Errorf("missing status %q", status)
		}
	}
	if len(requesters) != 2 {
		t.Errorf("requester count = %d, want 2", len(requesters))
	}
	if len(ages) != len(demoRequests) {
		t.Errorf("distinct age count = %d, want %d", len(ages), len(demoRequests))
	}
}

func TestGenerateVerifiedHash(t *testing.T) {
	password := "SenhaLocal#1"
	hash, err := generateVerifiedHash(password)
	if err != nil {
		t.Fatalf("generateVerifiedHash() error = %v", err)
	}
	if hash == password || !strings.HasPrefix(hash, "$2") {
		t.Fatal("generated value is not a bcrypt hash")
	}
}
