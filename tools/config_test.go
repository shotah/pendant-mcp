package tools

import (
	"testing"
)

func TestMailboxToAPI(t *testing.T) {
	t.Parallel()
	base, slug, err := mailboxToAPI("wss://gantry-pendant.example.workers.dev/ws/kit")
	if err != nil {
		t.Fatal(err)
	}
	if base != "https://gantry-pendant.example.workers.dev" || slug != "kit" {
		t.Fatalf("base=%q slug=%q", base, slug)
	}
	base, slug, err = mailboxToAPI("ws://127.0.0.1:3000/ws/kit")
	if err != nil || base != "http://127.0.0.1:3000" || slug != "kit" {
		t.Fatalf("local base=%q slug=%q err=%v", base, slug, err)
	}
	if _, _, err := mailboxToAPI("not a url"); err == nil {
		t.Fatal("junk should fail")
	}
	if _, _, err := mailboxToAPI("wss://x.example/ws/1bad"); err == nil {
		t.Fatal("bad slug should fail")
	}
	if _, _, err := mailboxToAPI("ftp://x.example/ws/kit"); err == nil {
		t.Fatal("ftp should fail")
	}
	if _, _, err := mailboxToAPI("https://x.example/nope"); err == nil {
		t.Fatal("missing /ws/<slug> should fail")
	}
}

func TestLoadConfigFromEnv(t *testing.T) {
	t.Parallel()
	cfg, err := loadConfigFromEnv(func(key string) string {
		switch key {
		case EnvPendantMailboxURL:
			return "wss://example.workers.dev/ws/kit"
		case EnvPendantBearer:
			return "secret"
		case EnvImageOutputDir:
			return "/tmp/out"
		default:
			return ""
		}
	})
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Slug != "kit" || cfg.Bearer != "secret" || cfg.MailboxURL != "https://example.workers.dev" {
		t.Fatalf("%+v", cfg)
	}
	if len(cfg.ImageRoots) != 1 || cfg.ImageRoots[0] != "/tmp/out" {
		t.Fatalf("roots = %v", cfg.ImageRoots)
	}
	if blobURL(cfg, "avatar") != "https://example.workers.dev/api/avatar?slug=kit" {
		t.Fatalf("avatar url = %s", blobURL(cfg, "avatar"))
	}
}

func TestLoadConfigMissing(t *testing.T) {
	t.Parallel()
	if _, err := loadConfigFromEnv(func(string) string { return "" }); err == nil {
		t.Fatal("want mailbox error")
	}
	if _, err := loadConfigFromEnv(func(key string) string {
		if key == EnvPendantMailboxURL {
			return "wss://x.example/ws/kit"
		}
		return ""
	}); err == nil {
		t.Fatal("want bearer error")
	}
}

func TestOptionalEnvKeys(t *testing.T) {
	t.Parallel()
	keys := OptionalEnvKeys()
	if len(keys) != 4 {
		t.Fatalf("keys = %v", keys)
	}
}
