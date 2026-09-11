package tools

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	mcpserver "github.com/mark3labs/mcp-go/server"
)

// KnownThemeIDs is the closed catalog (Worker is the source of truth on GET).
var KnownThemeIDs = map[string]bool{
	"boom": true, "inlay": true, "lamp": true,
	"noir": true, "ember": true, "tide": true, "bloom": true,
}

func registerTheme(s *mcpserver.MCPServer) {
	list := mcp.NewTool(ToolThemeList,
		mcp.WithDescription("List the color moods you can put on the phone app. Each card has a mood line and two hexes (canvas = background, accent = highlight). Pick one id, then call theme_update. Not for drawing a picture — that is image__photo_generate."),
	)
	registerTool(s, list, handleThemeList)

	update := mcp.NewTool(ToolThemeUpdate,
		mcp.WithDescription("Set the phone app's color mood. Pass a theme id from theme_list (boom, inlay, lamp, noir, ember, tide, bloom). Match the mood line to how you feel; do not invent hex. Humans can unfollow and keep their own theme."),
		mcp.WithString("theme", mcp.Required(), mcp.Description("Catalog id from theme_list (boom, inlay, lamp, noir, ember, tide, bloom).")),
	)
	registerTool(s, update, handleThemeUpdate)

	del := mcp.NewTool(ToolThemeDelete,
		mcp.WithDescription("Clear the room mood on the phone app. Phones that follow you fall back to the theme they picked. Does not change the face or the wallpaper."),
	)
	registerTool(s, del, handleThemeDelete)
}

func handleThemeList(ctx context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	cfg, err := loadConfigFromEnv(os.Getenv)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	got, err := pendantRequest(ctx, cfg, http.MethodGet, blobURL(cfg, "theme"), nil, "")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	if got.Status >= 400 {
		return mcp.NewToolResultError(workerError(got.Body, got.Status)), nil
	}
	return mcp.NewToolResultText(strings.TrimSpace(string(got.Body))), nil
}

func handleThemeUpdate(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	theme := strings.ToLower(strings.TrimSpace(request.GetString("theme", "")))
	if theme == "" || !KnownThemeIDs[theme] {
		return mcp.NewToolResultError("theme must be a catalog id. " + nextThemeList), nil
	}
	cfg, err := loadConfigFromEnv(os.Getenv)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	body, err := json.Marshal(map[string]string{"theme": theme})
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	got, err := pendantRequest(ctx, cfg, http.MethodPost, blobURL(cfg, "theme"), body, "application/json")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	if got.Status >= 400 {
		return mcp.NewToolResultError(workerError(got.Body, got.Status)), nil
	}
	return mcp.NewToolResultText(strings.TrimSpace(string(got.Body))), nil
}

func handleThemeDelete(ctx context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	cfg, err := loadConfigFromEnv(os.Getenv)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	got, err := pendantRequest(ctx, cfg, http.MethodDelete, blobURL(cfg, "theme"), nil, "")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	if got.Status >= 400 {
		return mcp.NewToolResultError(workerError(got.Body, got.Status)), nil
	}
	return mcp.NewToolResultText(strings.TrimSpace(string(got.Body))), nil
}
