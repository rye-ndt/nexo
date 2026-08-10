package mcp_proxy

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"hexago/internal/helpers/constances"
	input_itf "hexago/internal/interface/input"

	"github.com/google/uuid"
)

const (
	figmaNotConnected = "Figma is not connected. Add a Figma personal access token in Settings, then try again."
	figmaReadTimeout  = 30 * time.Second
	maxFigmaBody      = 8 << 20
)

type figmaArgs struct {
	FileKey string  `json:"file_key"`
	IDs     string  `json:"ids"`
	Depth   *int    `json:"depth"`
	Format  string  `json:"format"`
	Scale   float64 `json:"scale"`
}

type figmaCall func(args *figmaArgs) (path string, query url.Values, bad *toolResult)

func (s *v1) serveFigma(w http.ResponseWriter, r *http.Request) {
	serveRPC(w, r, constances.FigmaLocalServer, s.figmaTools())
}

func (s *v1) figmaTools() []*rpcTool {
	fileKey := stringProp("The Figma file key, taken from the file URL.")

	return []*rpcTool{
		s.figmaTool("figma_whoami",
			"Return the Figma user the stored personal access token belongs to.",
			objectSchema(map[string]any{}),
			func(*figmaArgs) (string, url.Values, *toolResult) {
				return "/v1/me", nil, nil
			},
		),
		s.figmaTool("figma_get_file",
			"Fetch the document tree of a Figma file by its file key.",
			objectSchema(map[string]any{
				"file_key": fileKey,
				"depth":    map[string]any{"type": "integer", "description": "How many levels of the node tree to return."},
				"ids":      stringProp("Comma-separated node ids to limit the returned tree to."),
			}, "file_key"),
			func(args *figmaArgs) (string, url.Values, *toolResult) {
				if bad := required("file_key", args.FileKey); bad != nil {
					return "", nil, bad
				}

				query := url.Values{}
				if args.Depth != nil {
					query.Set("depth", strconv.Itoa(*args.Depth))
				}
				if ids := strings.TrimSpace(args.IDs); ids != "" {
					query.Set("ids", ids)
				}

				return "/v1/files/" + url.PathEscape(args.FileKey), query, nil
			},
		),
		s.figmaTool("figma_get_nodes",
			"Fetch specific nodes from a Figma file by their ids.",
			objectSchema(map[string]any{
				"file_key": fileKey,
				"ids":      stringProp("Comma-separated node ids to fetch."),
			}, "file_key", "ids"),
			func(args *figmaArgs) (string, url.Values, *toolResult) {
				if bad := required("file_key", args.FileKey, "ids", args.IDs); bad != nil {
					return "", nil, bad
				}

				return "/v1/files/" + url.PathEscape(args.FileKey) + "/nodes", url.Values{"ids": {args.IDs}}, nil
			},
		),
		s.figmaTool("figma_get_images",
			"Render nodes of a Figma file to image URLs.",
			objectSchema(map[string]any{
				"file_key": fileKey,
				"ids":      stringProp("Comma-separated node ids to render."),
				"format": map[string]any{
					"type":        "string",
					"enum":        []string{"png", "jpg", "svg", "pdf"},
					"description": "Image format to render, defaults to png.",
				},
				"scale": map[string]any{"type": "number", "description": "Render scale between 0.01 and 4."},
			}, "file_key", "ids"),
			func(args *figmaArgs) (string, url.Values, *toolResult) {
				if bad := required("file_key", args.FileKey, "ids", args.IDs); bad != nil {
					return "", nil, bad
				}

				format := strings.TrimSpace(args.Format)
				if format == "" {
					format = "png"
				}

				query := url.Values{"ids": {args.IDs}, "format": {format}}
				if args.Scale > 0 {
					query.Set("scale", strconv.FormatFloat(args.Scale, 'f', -1, 64))
				}

				return "/v1/images/" + url.PathEscape(args.FileKey), query, nil
			},
		),
		s.figmaTool("figma_get_components",
			"List the components published from a Figma file.",
			objectSchema(map[string]any{"file_key": fileKey}, "file_key"),
			s.figmaFileRead("/components"),
		),
		s.figmaTool("figma_get_styles",
			"List the styles published from a Figma file.",
			objectSchema(map[string]any{"file_key": fileKey}, "file_key"),
			s.figmaFileRead("/styles"),
		),
	}
}

func (s *v1) figmaTool(name, description string, input map[string]any, request figmaCall) *rpcTool {
	return &rpcTool{
		name:        name,
		description: description,
		input:       input,
		call: func(arguments json.RawMessage, _ uuid.UUID) *toolResult {
			args := &figmaArgs{}

			if len(arguments) > 0 {
				if err := json.Unmarshal(arguments, args); err != nil {
					return errorResult("cannot parse tool arguments: " + err.Error())
				}
			}

			path, query, bad := request(args)
			if bad != nil {
				return bad
			}

			return s.figmaGet(path, query)
		},
	}
}

func (s *v1) figmaFileRead(suffix string) figmaCall {
	return func(args *figmaArgs) (string, url.Values, *toolResult) {
		if bad := required("file_key", args.FileKey); bad != nil {
			return "", nil, bad
		}

		return "/v1/files/" + url.PathEscape(args.FileKey) + suffix, nil, nil
	}
}

// required takes name/value pairs and names the first one the caller left blank.
func required(pairs ...string) *toolResult {
	for i := 0; i+1 < len(pairs); i += 2 {
		if strings.TrimSpace(pairs[i+1]) == "" {
			return errorResult(pairs[i] + " is required")
		}
	}

	return nil
}

func (s *v1) figmaGet(path string, query url.Values) *toolResult {
	cred, err := s.credentials(constances.FigmaLocalServer)
	if err != nil {
		return errorResult(figmaNotConnected)
	}

	full := strings.TrimRight(s.cfg.SupportedServers[constances.FigmaLocalServer].URL, "/") + path
	if len(query) > 0 {
		full += "?" + query.Encode()
	}

	res, err := s.httpCli.Stream(&input_itf.HttpRequest{
		Method: http.MethodGet,
		URL:    full,
		Header: http.Header{
			"X-Figma-Token": {cred.token},
			"Accept":        {"application/json"},
		},
	})
	if err != nil {
		return errorResult(fmt.Sprintf("figma request failed: %v", err))
	}
	defer res.Body.Close()

	timer := time.AfterFunc(figmaReadTimeout, func() { res.Body.Close() })
	body, readErr := io.ReadAll(io.LimitReader(res.Body, maxFigmaBody))
	if !timer.Stop() {
		return errorResult("figma request timed out")
	}
	if readErr != nil {
		return errorResult(fmt.Sprintf("figma read failed: %v", readErr))
	}

	if res.StatusCode >= http.StatusBadRequest {
		return errorResult(fmt.Sprintf("figma api %d: %s", res.StatusCode, string(body)))
	}

	return textResult(string(body))
}
