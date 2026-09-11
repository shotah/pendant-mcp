package tools

import (
	"errors"
	"net/url"
	"regexp"
	"strings"
)

const (
	EnvPendantMailboxURL = "PENDANT_MAILBOX_URL"
	EnvPendantBearer     = "PENDANT_BEARER" //nolint:gosec // G101: env var name, not a credential
	EnvImageOutputDir    = "IMAGE_OUTPUT_DIR"
	EnvPendantImageDir   = "PENDANT_IMAGE_DIR"
)

const (
	AvatarEdge       = 1280
	BackdropEdge     = 1600
	AvatarMaxBytes   = 5 * 1024 * 1024
	BackdropMaxBytes = 1_500_000
	JPEGEdgeMin      = 320
	JPEGEdgeStep     = 0.75
	HTTPTimeoutSec   = 15
)

var jpegQualities = []int{90, 80, 70, 60}

var slugRE = regexp.MustCompile(`^[a-z][a-z0-9-]{0,31}$`)

// HostManifestBlurb is the yard Secrets one-liner.
const HostManifestBlurb = "Pendant face / wallpaper / mood. Crane already has PENDANT_MAILBOX_URL + PENDANT_BEARER. IMAGE_OUTPUT_DIR is the photo_generate handoff; optional PENDANT_IMAGE_DIR is a second allowed root."

// OptionalEnvKeys lists Secrets fields that must not skip the server at boot.
func OptionalEnvKeys() []string {
	return []string{
		EnvPendantMailboxURL,
		EnvPendantBearer,
		EnvImageOutputDir,
		EnvPendantImageDir,
	}
}

type serverConfig struct {
	MailboxURL string
	Bearer     string
	Slug       string
	ImageRoots []string
}

func loadConfigFromEnv(getenv func(string) string) (serverConfig, error) {
	mailbox := strings.TrimSpace(getenv(EnvPendantMailboxURL))
	if mailbox == "" {
		return serverConfig{}, errNeedMailbox()
	}
	apiBase, slug, err := mailboxToAPI(mailbox)
	if err != nil {
		return serverConfig{}, err
	}
	bearer := strings.TrimSpace(getenv(EnvPendantBearer))
	if bearer == "" {
		return serverConfig{}, errNeedBearer()
	}
	return serverConfig{
		MailboxURL: apiBase,
		Bearer:     bearer,
		Slug:       slug,
		ImageRoots: imageRoots(getenv),
	}, nil
}

func imageRoots(getenv func(string) string) []string {
	var roots []string
	for _, key := range []string{EnvImageOutputDir, EnvPendantImageDir} {
		if dir := strings.TrimSpace(getenv(key)); dir != "" {
			roots = append(roots, dir)
		}
	}
	return roots
}

// mailboxToAPI turns wss://host/ws/kit into https://host plus slug kit.
// Tests match gantry-pendant mailboxToAvatarUrl cases.
func mailboxToAPI(raw string) (base string, slug string, err error) {
	u, err := url.Parse(strings.TrimSpace(raw))
	if err != nil || u.Host == "" {
		return "", "", errBadMailbox()
	}
	switch u.Scheme {
	case "wss":
		u.Scheme = "https"
	case "ws":
		u.Scheme = "http"
	case "http", "https":
	default:
		return "", "", errBadMailbox()
	}
	parts := strings.FieldsFunc(u.Path, func(r rune) bool { return r == '/' })
	if len(parts) < 2 || parts[0] != "ws" {
		return "", "", errBadMailbox()
	}
	slug = strings.ToLower(strings.TrimSpace(parts[1]))
	if !slugRE.MatchString(slug) {
		return "", "", errBadMailbox()
	}
	u.Path = ""
	u.RawPath = ""
	u.RawQuery = ""
	u.Fragment = ""
	return strings.TrimRight(u.String(), "/"), slug, nil
}

func blobURL(cfg serverConfig, kind string) string {
	u, err := url.Parse(cfg.MailboxURL + "/api/" + kind)
	if err != nil {
		return cfg.MailboxURL + "/api/" + kind + "?slug=" + url.QueryEscape(cfg.Slug)
	}
	q := u.Query()
	q.Set("slug", cfg.Slug)
	u.RawQuery = q.Encode()
	return u.String()
}

func errNeedMailbox() error {
	return errors.New(EnvPendantMailboxURL + " is required (wss://…/ws/<slug>). " + nextThemeList)
}

func errNeedBearer() error {
	return errors.New(EnvPendantBearer + " is required (Authorization: Bearer). " + nextThemeList)
}

func errBadMailbox() error {
	return errors.New(EnvPendantMailboxURL + " must be ws(s) or http(s) with /ws/<slug>. " + nextThemeList)
}

func errNeedImageDir() error {
	return errors.New(EnvImageOutputDir + " is required for source_path (the image__photo_generate handoff). Set it on the crane, then " + nextAvatarUpdate)
}
