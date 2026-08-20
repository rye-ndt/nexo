package mcp_proxy

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
	"unicode/utf8"

	"hexago/internal/helpers"
	"hexago/internal/helpers/constances"
	"hexago/internal/helpers/custom_error"
	"hexago/internal/helpers/enums"
	mcp_helpers "hexago/internal/implementation/core/mcp_proxy/helpers"
	input_itf "hexago/internal/interface/input"

	"github.com/google/uuid"
)

const chromeNotConnected = "Chrome is not connected. Open Settings, find chrome-devtools and click Enable, then try again."

const chromeGuidance = "You are driving the operator's own Chrome, already signed in to their accounts. " +
	"Never ask for a password or a one-time code: navigate to the page and read what is there. " +
	"Take a screenshot or read the page before acting on it, and prefer a css selector that is anchored " +
	"on visible text over one built from generated class names."

const (
	chromePollInterval  = 200 * time.Millisecond
	chromeMaxElements   = 200
	chromeEnabledMarker = "enabled"
	chromeEnabledTTL    = 100 * 365 * 24 * time.Hour
)

type chromeArgs struct {
	TabID      string `json:"tab_id"`
	URL        string `json:"url"`
	Selector   string `json:"selector"`
	Value      string `json:"value"`
	Key        string `json:"key"`
	Expression string `json:"expression"`
	TimeoutMS  int    `json:"timeout_ms"`
}

type chromeCall func(workflow *mcp_helpers.CDPSession, args *chromeArgs) *toolResult

type chromeBrowserCall func(args *chromeArgs) *toolResult

func (s *v1) serveChrome(w http.ResponseWriter, r *http.Request) {
	serveRPC(w, r, constances.ChromeLocalServer, s.chromeTools())
}

func (s *v1) chromeTools() []*rpcTool {
	tabID := stringProp("The tab to act on, from chrome_list_tabs. Defaults to the active tab.")
	selector := stringProp("A css selector for the element, matched in the page.")

	return []*rpcTool{
		s.chromeBrowserTool("chrome_list_tabs",
			"List the open Chrome tabs with their ids, titles and urls. "+chromeGuidance,
			objectSchema(map[string]any{}),
			s.chromeListTabs,
		),
		s.chromeBrowserTool("chrome_open_tab",
			"Open a new Chrome tab at a url and return its id.",
			objectSchema(map[string]any{"url": stringProp("The url to open.")}, "url"),
			s.chromeOpenTab,
		),
		s.chromeBrowserTool("chrome_close_tab",
			"Close one Chrome tab.",
			objectSchema(map[string]any{"tab_id": tabID}, "tab_id"),
			s.chromeCloseTab,
		),
		s.chromePageTool("chrome_navigate",
			"Point a tab at a url and wait for the page to load.",
			objectSchema(map[string]any{
				"url":    stringProp("The url to navigate to."),
				"tab_id": tabID,
			}, "url"),
			s.chromeNavigate,
		),
		s.chromePageTool("chrome_read_page",
			"Read a tab: its title, url, visible text, and the links, buttons and inputs "+
				"you can act on with their selectors. Read before you click.",
			objectSchema(map[string]any{"tab_id": tabID}),
			s.chromeReadPage,
		),
		s.chromePageTool("chrome_click",
			"Click the first element matching a selector.",
			objectSchema(map[string]any{"selector": selector, "tab_id": tabID}, "selector"),
			s.chromeClick,
		),
		s.chromePageTool("chrome_fill",
			"Type a value into the input matching a selector, replacing whatever is in it.",
			objectSchema(map[string]any{
				"selector": selector,
				"value":    stringProp("The text to put in the field."),
				"tab_id":   tabID,
			}, "selector", "value"),
			s.chromeFill,
		),
		s.chromePageTool("chrome_press_key",
			"Send one key to the focused element, such as Enter to submit a form.",
			objectSchema(map[string]any{
				"key":    stringProp("The key name, such as Enter, Tab, Escape or ArrowDown."),
				"tab_id": tabID,
			}, "key"),
			s.chromePressKey,
		),
		s.chromePageTool("chrome_wait_for",
			"Wait until an element appears in a tab. Use it after a click that loads new content.",
			objectSchema(map[string]any{
				"selector":   selector,
				"timeout_ms": map[string]any{"type": "integer", "description": "How long to wait before giving up."},
				"tab_id":     tabID,
			}, "selector"),
			s.chromeWaitFor,
		),
		s.chromePageTool("chrome_screenshot",
			"Capture what a tab currently looks like.",
			objectSchema(map[string]any{"tab_id": tabID}),
			s.chromeScreenshot,
		),
		s.chromePageTool("chrome_evaluate",
			"Run a javascript expression in a tab and return what it evaluates to. "+
				"Use it only when no other tool fits.",
			objectSchema(map[string]any{
				"expression": stringProp("The javascript expression to evaluate."),
				"tab_id":     tabID,
			}, "expression"),
			s.chromeEvaluate,
		),
	}
}

func (s *v1) chromeBrowserTool(name, description string, input map[string]any, call chromeBrowserCall) *rpcTool {
	return &rpcTool{
		name:        name,
		description: description,
		input:       input,
		call: func(arguments json.RawMessage, _ uuid.UUID) *toolResult {
			args, bad := chromeArguments(arguments)
			if bad != nil {
				return bad
			}

			if bad := s.chromeReady(); bad != nil {
				return bad
			}

			return call(args)
		},
	}
}

func (s *v1) chromePageTool(name, description string, input map[string]any, call chromeCall) *rpcTool {
	return &rpcTool{
		name:        name,
		description: description,
		input:       input,
		call: func(arguments json.RawMessage, _ uuid.UUID) *toolResult {
			args, bad := chromeArguments(arguments)
			if bad != nil {
				return bad
			}

			if bad := s.chromeReady(); bad != nil {
				return bad
			}

			target, bad := s.resolveTarget(args.TabID)
			if bad != nil {
				return bad
			}

			workflow, err := mcp_helpers.DialCDP(target, s.cfg.Chrome.CallTimeout)
			if err != nil {
				return errorResult("cannot attach to tab " + target.ID + ": " + err.Error())
			}
			defer workflow.Close()

			return call(workflow, args)
		},
	}
}

func (s *v1) chromeReady() *toolResult {
	if !s.chromeEnabled() {
		return errorResult(chromeNotConnected)
	}

	if err := s.chrome.Ensure(); err != nil {
		return errorResult("cannot start Chrome: " + err.Error())
	}

	return nil
}

func chromeArguments(arguments json.RawMessage) (*chromeArgs, *toolResult) {
	args := &chromeArgs{}

	if len(arguments) > 0 {
		if err := json.Unmarshal(arguments, args); err != nil {
			return nil, errorResult("cannot parse tool arguments: " + err.Error())
		}
	}

	return args, nil
}

func (s *v1) resolveTarget(tabID string) (*mcp_helpers.CDPTarget, *toolResult) {
	targets, err := mcp_helpers.CDPTargets(s.httpCli, s.chrome.Endpoint())
	if err != nil {
		return nil, errorResult("cannot list Chrome tabs: " + err.Error())
	}

	if strings.TrimSpace(tabID) == "" {
		if len(targets) == 0 {
			return nil, errorResult("no Chrome tab is open, open one with chrome_open_tab")
		}

		return targets[0], nil
	}

	for _, target := range targets {
		if target.ID == tabID {
			return target, nil
		}
	}

	return nil, errorResult("unknown tab id " + tabID + ", list the open tabs with chrome_list_tabs")
}

func (s *v1) chromeListTabs(args *chromeArgs) *toolResult {
	targets, err := mcp_helpers.CDPTargets(s.httpCli, s.chrome.Endpoint())
	if err != nil {
		return errorResult("cannot list Chrome tabs: " + err.Error())
	}

	tabs := make([]map[string]string, 0, len(targets))
	for _, target := range targets {
		tabs = append(tabs, map[string]string{
			"id":    target.ID,
			"title": target.Title,
			"url":   target.URL,
		})
	}

	return jsonResult(tabs)
}

func (s *v1) chromeOpenTab(args *chromeArgs) *toolResult {
	if bad := required("url", args.URL); bad != nil {
		return bad
	}

	target, err := mcp_helpers.CDPNewTarget(s.httpCli, s.chrome.Endpoint(), args.URL)
	if err != nil {
		return errorResult("cannot open a Chrome tab: " + err.Error())
	}

	return jsonResult(map[string]string{"id": target.ID, "url": target.URL})
}

func (s *v1) chromeCloseTab(args *chromeArgs) *toolResult {
	if bad := required("tab_id", args.TabID); bad != nil {
		return bad
	}

	if err := mcp_helpers.CDPCloseTarget(s.httpCli, s.chrome.Endpoint(), args.TabID); err != nil {
		return errorResult("cannot close tab " + args.TabID + ": " + err.Error())
	}

	return textResult("closed tab " + args.TabID)
}

func (s *v1) chromeNavigate(workflow *mcp_helpers.CDPSession, args *chromeArgs) *toolResult {
	if bad := required("url", args.URL); bad != nil {
		return bad
	}

	if _, err := workflow.Call("Page.navigate", map[string]any{"url": args.URL}); err != nil {
		return errorResult("cannot navigate to " + args.URL + ": " + err.Error())
	}

	page := struct {
		Ready bool   `json:"ready"`
		URL   string `json:"url"`
		Title string `json:"title"`
	}{}

	loaded := pollFor(s.cfg.Chrome.CallTimeout, func() (bool, error) {
		if err := evalInto(workflow, chromeNavigatedJS, &page); err != nil {
			return false, err
		}

		return page.Ready, nil
	})

	return jsonResult(map[string]any{
		"url":    page.URL,
		"title":  page.Title,
		"loaded": loaded,
	})
}

func (s *v1) chromeReadPage(workflow *mcp_helpers.CDPSession, args *chromeArgs) *toolResult {
	raw, err := workflow.Eval(fmt.Sprintf(chromeReadPageJS, s.cfg.Chrome.MaxPageChars, chromeMaxElements))
	if err != nil {
		return errorResult("cannot read the page: " + err.Error())
	}

	return textResult(string(raw))
}

func (s *v1) chromeClick(workflow *mcp_helpers.CDPSession, args *chromeArgs) *toolResult {
	if bad := required("selector", args.Selector); bad != nil {
		return bad
	}

	clicked := false
	if err := evalInto(workflow, fmt.Sprintf(chromeClickJS, jsString(args.Selector)), &clicked); err != nil {
		return errorResult("cannot click " + args.Selector + ": " + err.Error())
	}

	if !clicked {
		return errorResult("no element matches " + args.Selector + ", read the page again for its selectors")
	}

	return textResult("clicked " + args.Selector)
}

func (s *v1) chromeFill(workflow *mcp_helpers.CDPSession, args *chromeArgs) *toolResult {
	if bad := required("selector", args.Selector); bad != nil {
		return bad
	}

	filled := false
	expression := fmt.Sprintf(chromeFillJS, jsString(args.Selector), jsString(args.Value))

	if err := evalInto(workflow, expression, &filled); err != nil {
		return errorResult("cannot fill " + args.Selector + ": " + err.Error())
	}

	if !filled {
		return errorResult("no element matches " + args.Selector + ", read the page again for its selectors")
	}

	return textResult("filled " + args.Selector)
}

func (s *v1) chromePressKey(workflow *mcp_helpers.CDPSession, args *chromeArgs) *toolResult {
	if bad := required("key", args.Key); bad != nil {
		return bad
	}

	event, bad := chromeKeyEvent(args.Key)
	if bad != nil {
		return bad
	}

	for _, phase := range []string{"keyDown", "keyUp"} {
		params := map[string]any{"type": phase}
		for key, value := range event {
			params[key] = value
		}

		if _, err := workflow.Call("Input.dispatchKeyEvent", params); err != nil {
			return errorResult("cannot send " + args.Key + ": " + err.Error())
		}
	}

	return textResult("pressed " + args.Key)
}

func (s *v1) chromeWaitFor(workflow *mcp_helpers.CDPSession, args *chromeArgs) *toolResult {
	if bad := required("selector", args.Selector); bad != nil {
		return bad
	}

	timeout := s.cfg.Chrome.CallTimeout
	if args.TimeoutMS > 0 {
		timeout = time.Duration(args.TimeoutMS) * time.Millisecond
		if timeout > s.cfg.Chrome.CallTimeout {
			timeout = s.cfg.Chrome.CallTimeout
		}
	}

	expression := "!!document.querySelector(" + jsString(args.Selector) + ")"

	found := pollFor(timeout, func() (bool, error) {
		present := false
		err := evalInto(workflow, expression, &present)

		return present, err
	})

	if !found {
		return errorResult("timed out after " + timeout.String() + " waiting for " + args.Selector)
	}

	return textResult(args.Selector + " is present")
}

func (s *v1) chromeScreenshot(workflow *mcp_helpers.CDPSession, args *chromeArgs) *toolResult {
	raw, err := workflow.Call("Page.captureScreenshot", map[string]any{"format": "png"})
	if err != nil {
		return errorResult("cannot capture the tab: " + err.Error())
	}

	shot := struct {
		Data string `json:"data"`
	}{}

	if err := json.Unmarshal(raw, &shot); err != nil {
		return errorResult("cannot read the screenshot: " + err.Error())
	}

	return imageResult(shot.Data, "image/png")
}

func (s *v1) chromeEvaluate(workflow *mcp_helpers.CDPSession, args *chromeArgs) *toolResult {
	if bad := required("expression", args.Expression); bad != nil {
		return bad
	}

	raw, err := workflow.Eval(args.Expression)
	if err != nil {
		return errorResult("cannot evaluate the expression: " + err.Error())
	}

	return textResult(string(raw))
}

func (s *v1) chromeEnabled() bool {
	stored, err := s.db.GetCredentials(constances.ChromeLocalServer)

	return err == nil && stored != nil && stored.EncryptedOAuthKey != ""
}

func (s *v1) enableChrome() error {
	if err := s.chrome.Ensure(); err != nil {
		return err
	}

	encrypted, err := mcp_helpers.Encrypt(s.aead, chromeEnabledMarker)
	if err != nil {
		return err
	}

	now := helpers.NewUTC()

	if err := s.db.UpsertCredentials(&input_itf.MCPEntity{
		Name:              constances.ChromeLocalServer,
		EncryptedOAuthKey: encrypted,
		ExpiredAt:         now.Add(chromeEnabledTTL),
		CreatedAt:         now,
		UpdatedAt:         now,
	}); err != nil {
		return custom_error.TypedCritical(
			enums.ErrMcpStoreCredentials,
			"cannot store credentials for %s: %v",
			constances.ChromeLocalServer,
			err,
		)
	}

	return nil
}

func (s *v1) disableChrome() error {
	stopped := s.chrome.Stop()

	if err := s.db.DeleteCredentials(constances.ChromeLocalServer); err != nil {
		return custom_error.TypedCritical(
			enums.ErrMcpStoreCredentials,
			"cannot delete credentials for %s: %v",
			constances.ChromeLocalServer,
			err,
		)
	}

	s.uncache(constances.ChromeLocalServer)

	return stopped
}

func chromeKeyEvent(name string) (map[string]any, *toolResult) {
	key, found := chromeKeys[strings.ToLower(strings.TrimSpace(name))]

	if !found {
		printable := strings.TrimSpace(name)
		if utf8.RuneCountInString(printable) != 1 {
			return nil, errorResult("unsupported key " + name +
				", use a single character or one of Enter, Tab, Escape, Backspace, ArrowUp, ArrowDown, ArrowLeft, ArrowRight")
		}

		return map[string]any{"key": printable, "text": printable, "unmodifiedText": printable}, nil
	}

	event := map[string]any{
		"key":                   key.key,
		"code":                  key.code,
		"windowsVirtualKeyCode": key.virtualKey,
		"nativeVirtualKeyCode":  key.virtualKey,
	}

	if key.text != "" {
		event["text"] = key.text
		event["unmodifiedText"] = key.text
	}

	return event, nil
}

type chromeKey struct {
	key        string
	code       string
	virtualKey int
	text       string
}

var chromeKeys = map[string]chromeKey{
	"enter":      {key: "Enter", code: "Enter", virtualKey: 13, text: "\r"},
	"tab":        {key: "Tab", code: "Tab", virtualKey: 9},
	"escape":     {key: "Escape", code: "Escape", virtualKey: 27},
	"backspace":  {key: "Backspace", code: "Backspace", virtualKey: 8},
	"arrowup":    {key: "ArrowUp", code: "ArrowUp", virtualKey: 38},
	"arrowdown":  {key: "ArrowDown", code: "ArrowDown", virtualKey: 40},
	"arrowleft":  {key: "ArrowLeft", code: "ArrowLeft", virtualKey: 37},
	"arrowright": {key: "ArrowRight", code: "ArrowRight", virtualKey: 39},
}

func pollFor(timeout time.Duration, check func() (bool, error)) bool {
	deadline := time.Now().Add(timeout)

	for {
		if done, err := check(); err == nil && done {
			return true
		}

		if time.Now().After(deadline) {
			return false
		}

		time.Sleep(chromePollInterval)
	}
}

func evalInto(workflow *mcp_helpers.CDPSession, expression string, value any) error {
	raw, err := workflow.Eval(expression)
	if err != nil {
		return err
	}

	return json.Unmarshal(raw, value)
}

func jsString(value string) string {
	encoded, err := json.Marshal(value)
	if err != nil {
		return `""`
	}

	return string(encoded)
}

func jsonResult(payload any) *toolResult {
	encoded, err := json.Marshal(payload)
	if err != nil {
		return errorResult("cannot encode the result: " + err.Error())
	}

	return textResult(string(encoded))
}

const chromeNavigatedJS = `({ready: document.readyState === "complete", url: location.href, title: document.title})`

const chromeClickJS = `(() => {
  const el = document.querySelector(%s);
  if (!el) return false;
  el.scrollIntoView({block: "center", inline: "center"});
  el.click();
  return true;
})()`

const chromeFillJS = `(() => {
  const el = document.querySelector(%s);
  if (!el) return false;
  el.focus();
  el.value = %s;
  el.dispatchEvent(new Event("input", {bubbles: true}));
  el.dispatchEvent(new Event("change", {bubbles: true}));
  return true;
})()`

const chromeReadPageJS = `(() => {
  const selectorFor = (el) => {
    if (el.id) return "#" + CSS.escape(el.id);
    const tag = el.tagName.toLowerCase();
    const name = el.getAttribute("name");
    if (name) return tag + "[name=" + JSON.stringify(name) + "]";
    const label = el.getAttribute("aria-label");
    if (label) return tag + "[aria-label=" + JSON.stringify(label) + "]";
    const path = [];
    let node = el;
    while (node && node.nodeType === 1 && node !== document.documentElement) {
      const parent = node.parentElement;
      const step = node.tagName.toLowerCase();
      if (!parent) {
        path.unshift(step);
        break;
      }
      const twins = Array.from(parent.children).filter((c) => c.tagName === node.tagName);
      path.unshift(twins.length > 1 ? step + ":nth-of-type(" + (twins.indexOf(node) + 1) + ")" : step);
      node = parent;
    }
    return path.join(" > ");
  };

  const maxChars = %d;
  const maxElements = %d;
  const elements = [];
  let moreElements = false;

  for (const el of document.querySelectorAll('a[href], button, input, textarea, select, [role="button"]')) {
    const box = el.getBoundingClientRect();
    if (!box.width || !box.height) continue;
    if (elements.length >= maxElements) {
      moreElements = true;
      break;
    }
    elements.push({
      selector: selectorFor(el),
      tag: el.tagName.toLowerCase(),
      type: el.getAttribute("type") || "",
      name: el.getAttribute("name") || "",
      text: ((el.innerText || el.value || el.getAttribute("aria-label") || "") + "").trim().slice(0, 120),
    });
  }

  const text = document.body ? document.body.innerText : "";

  return {
    title: document.title,
    url: location.href,
    text: text.slice(0, maxChars),
    text_truncated: text.length > maxChars,
    elements: elements,
    elements_truncated: moreElements,
  };
})()`
