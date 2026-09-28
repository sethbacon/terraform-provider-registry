package client

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"testing"
)

// This file is the route contract between this package and the backend: every
// HTTP call the client makes must name an operation in the vendored OpenAPI
// spec, spec/openapi3.json, a verbatim copy of the backend release recorded in
// spec/BACKEND_VERSION.
//
// WHY. The client's paths are hand-written strings, and the unit tests that
// drive them against httptest servers pin the path the client sends, not that
// the backend serves it. Two routes that never existed shipped exactly that
// way and stayed broken until users hit them: GET
// /api/v1/providers/{ns}/{type}/versions/{v}, whose 404 dropped every
// registry_provider_version_deprecation from state on refresh, and DELETE
// /api/v1/admin/approvals/{id}, whose 404 Client.Delete reports as success.
//
// HOW. The package's non-test sources are parsed with go/ast, and every call
// to a transport entry point (Get, Post, Put, Delete or Do on a *Client, and
// FetchAllPages) is lowered to a method and a path template: string literals
// are kept, fmt.Sprintf verbs and every other expression become {}, a local
// path variable contributes each value assigned to it, and the query string
// is dropped. The spec's paths are lowered the same way ({name} becomes {}),
// and the two must match exactly.
//
// WHAT FAILS, so that the test cannot pass by not looking:
//   - a call whose method or path cannot be read statically;
//   - a path-shaped string literal that reaches no recognised call, such as a
//     path handed to a new wrapper the scanner does not know;
//   - a vendored spec that no longer matches the checksum fetch-spec.sh
//     recorded, so the spec cannot be edited into agreement with the client;
//   - a knownRouteGaps entry with no reason, or one that is no longer needed.

// knownRouteGaps lists client calls whose route is deliberately absent from the
// vendored spec, keyed "METHOD /path" with every path parameter written as {}
// (the form the failure message prints). Each entry must say why the call is
// right anyway and what retires the entry, typically vendoring a backend
// release whose spec documents the route.
//
// It is empty because every route the client calls is in the spec of the
// release in spec/BACKEND_VERSION.
var knownRouteGaps = map[string]string{}

func TestClientRoutesExistInVendoredSpec(t *testing.T) {
	backend := readBackendVersion(t)
	spec, err := specRoutes(readSpecFile(t, "openapi3.json"))
	if err != nil {
		t.Fatalf("spec/openapi3.json: %v", err)
	}
	fset, files := parseClientSources(t)
	scan := scanRoutes(fset, files)

	for _, u := range scan.unresolved {
		t.Errorf("cannot check %s", u)
	}
	for _, s := range scan.stray {
		t.Errorf("%s reaches no call this test recognises (Get, Post, Put, Delete or Do on a *Client, or FetchAllPages), so its route goes unchecked; call one of those directly, or teach scanRoutes the new call shape", s)
	}
	if len(scan.calls) == 0 {
		t.Fatal("found no client calls at all: scanRoutes no longer understands this package")
	}

	problems, allowed := checkRoutes(scan.calls, spec, knownRouteGaps, backend)
	for _, a := range allowed {
		t.Log(a)
	}
	for _, p := range problems {
		t.Error(p)
	}
	if len(problems) > 0 {
		t.Log("fix the path; or, if the backend really serves the route, vendor a release whose spec documents it (spec/fetch-spec.sh); or record the call in knownRouteGaps with a reason")
	}

	distinct := map[string]bool{}
	for _, c := range scan.calls {
		distinct[c.route()] = true
	}
	t.Logf("%d call sites, %d distinct routes, checked against the spec of backend %s", len(scan.calls), len(distinct), backend)
}

// TestVendoredSpecIsTheRecordedBackendRelease holds spec/openapi3.json to the
// checksum fetch-spec.sh wrote when it copied the file out of the release in
// spec/BACKEND_VERSION. CI no longer fetches the spec from a running backend,
// so without this nothing would stop a hand edit to the vendored copy, which
// is exactly how a missing route could be made to pass the test above.
func TestVendoredSpecIsTheRecordedBackendRelease(t *testing.T) {
	backend := readBackendVersion(t)
	want, err := parseSpecChecksum(string(readSpecFile(t, "openapi3.json.sha256")))
	if err != nil {
		t.Fatalf("spec/openapi3.json.sha256: %v", err)
	}
	sum := sha256.Sum256(readSpecFile(t, "openapi3.json"))
	if got := hex.EncodeToString(sum[:]); got != want {
		t.Fatalf("spec/openapi3.json has SHA-256 %s, but fetch-spec.sh recorded %s when it copied backend %s's spec; the vendored spec must stay a verbatim copy, so refresh it with spec/fetch-spec.sh instead of editing it", got, want, backend)
	}
}

// clientCall is one HTTP call site, lowered to a route template.
type clientCall struct {
	method string // as sent, e.g. "GET"
	path   string // {} for every non-literal part, query string removed
	pos    string // e.g. "apikeys.go:21 (GetAPIKey)"
}

func (c clientCall) route() string { return c.method + " " + c.path }

// routeScan is what scanRoutes found.
type routeScan struct {
	calls      []clientCall
	unresolved []string // call sites whose method or path cannot be read statically
	stray      []string // path-shaped literals that reach no recognised call
}

// transportMethods maps each *Client method that sends a request to the HTTP
// method it uses; Do takes its method as an argument instead.
var transportMethods = map[string]string{
	"Get":    http.MethodGet,
	"Post":   http.MethodPost,
	"Put":    http.MethodPut,
	"Delete": http.MethodDelete,
	"Do":     "",
}

var (
	// pathShaped matches literals that look like a route ("/api/...",
	// "/version"); a lone "/" (a separator, TrimRight's cutset) does not.
	pathShaped = regexp.MustCompile(`^/[A-Za-z0-9._~-]`)
	// sprintfVerb matches one fmt verb with its flags and width, or "%%".
	sprintfVerb = regexp.MustCompile(`%%|%[^A-Za-z%]*[A-Za-z]`)
	// specParam matches one templated path parameter, e.g. {namespace}.
	specParam = regexp.MustCompile(`\{[^{}/]+\}`)
	// releaseTag is the form of the backend's release tags.
	releaseTag = regexp.MustCompile(`^v\d+\.\d+\.\d+(-[0-9A-Za-z.-]+)?$`)
	// sha256Hex is a lower-case hex SHA-256, as sha256sum prints it.
	sha256Hex = regexp.MustCompile(`^[0-9a-f]{64}$`)
)

// scanRoutes lowers every transport call in files to a clientCall. The
// transport entry points themselves are skipped: their bodies only forward a
// path parameter.
func scanRoutes(fset *token.FileSet, files []*ast.File) routeScan {
	var s routeScan
	consumed := map[token.Pos]bool{}
	for _, f := range files {
		for _, decl := range f.Decls {
			fd, ok := decl.(*ast.FuncDecl)
			if !ok || fd.Body == nil || isTransport(fd) {
				continue
			}
			sc := newFuncScope(fd, consumed)
			ast.Inspect(fd.Body, func(n ast.Node) bool {
				call, ok := n.(*ast.CallExpr)
				if !ok {
					return true
				}
				method, pathArg, isCall, err := sc.transportCall(call)
				if !isCall {
					return true
				}
				where := fmt.Sprintf("%s (%s)", position(fset, call.Pos()), fd.Name.Name)
				if err != nil {
					s.unresolved = append(s.unresolved, where+": "+err.Error())
					if pathArg != nil {
						_, _ = sc.lower(pathArg) // only to mark its literals seen; the call is already reported
					}
					return true
				}
				paths, err := sc.lower(pathArg)
				if err != nil {
					s.unresolved = append(s.unresolved, where+": "+err.Error())
					return true
				}
				seen := map[string]bool{}
				for _, p := range paths {
					p, _, _ = strings.Cut(p, "?")
					switch {
					case seen[p]:
					case !strings.HasPrefix(p, "/"):
						s.unresolved = append(s.unresolved, fmt.Sprintf("%s: the path does not start with a literal \"/\" (it lowers to %q)", where, p))
					default:
						s.calls = append(s.calls, clientCall{method: method, path: p, pos: where})
					}
					seen[p] = true
				}
				return true
			})
		}
	}
	for _, f := range files {
		ast.Inspect(f, func(n ast.Node) bool {
			lit, ok := n.(*ast.BasicLit)
			if !ok || lit.Kind != token.STRING || consumed[lit.Pos()] {
				return true
			}
			if v, err := strconv.Unquote(lit.Value); err == nil && pathShaped.MatchString(v) {
				s.stray = append(s.stray, fmt.Sprintf("%s: %q", position(fset, lit.Pos()), v))
			}
			return true
		})
	}
	return s
}

// isTransport reports whether fd is a transport entry point, whose path is a
// parameter by design: Get, Post, Put, Delete or Do on *Client, or
// FetchAllPages.
func isTransport(fd *ast.FuncDecl) bool {
	if fd.Recv == nil {
		return fd.Name.Name == "FetchAllPages"
	}
	_, ok := transportMethods[fd.Name.Name]
	return ok && len(fd.Recv.List) == 1 && isClientPtr(fd.Recv.List[0].Type)
}

func isClientPtr(e ast.Expr) bool {
	star, ok := e.(*ast.StarExpr)
	if !ok {
		return false
	}
	id, ok := star.X.(*ast.Ident)
	return ok && id.Name == "Client"
}

// funcScope resolves the identifiers of one function: which ones hold a
// *Client, which are parameters, and every value each variable is assigned.
type funcScope struct {
	clients   map[string]bool
	params    map[string]bool
	assigns   map[string][]assignment
	resolving map[string]bool
	consumed  map[token.Pos]bool // shared by every scope, for the stray-literal check
}

// assignment is one write to a variable. value is nil when no single
// expression is assigned (a multi-value call, a range clause).
type assignment struct {
	value ast.Expr
	tok   token.Token // token.ADD_ASSIGN appends; anything else replaces
}

func newFuncScope(fd *ast.FuncDecl, consumed map[token.Pos]bool) *funcScope {
	sc := &funcScope{
		clients:   map[string]bool{},
		params:    map[string]bool{},
		assigns:   map[string][]assignment{},
		resolving: map[string]bool{},
		consumed:  consumed,
	}
	var fields []*ast.Field
	if fd.Recv != nil {
		fields = append(fields, fd.Recv.List...)
	}
	fields = append(fields, fd.Type.Params.List...)
	for _, field := range fields {
		for _, name := range field.Names {
			sc.params[name.Name] = true
			if isClientPtr(field.Type) {
				sc.clients[name.Name] = true
			}
		}
	}
	ast.Inspect(fd.Body, func(n ast.Node) bool {
		switch n := n.(type) {
		case *ast.AssignStmt:
			for i, lhs := range n.Lhs {
				var value ast.Expr
				if len(n.Rhs) == len(n.Lhs) {
					value = n.Rhs[i]
				}
				sc.assign(lhs, value, n.Tok)
			}
		case *ast.ValueSpec:
			for i, name := range n.Names {
				if i < len(n.Values) {
					sc.assign(name, n.Values[i], token.DEFINE)
				}
			}
		case *ast.RangeStmt:
			sc.assign(n.Key, nil, token.DEFINE)
			sc.assign(n.Value, nil, token.DEFINE)
		}
		return true
	})
	return sc
}

func (sc *funcScope) assign(lhs, value ast.Expr, tok token.Token) {
	if id, ok := lhs.(*ast.Ident); ok && id.Name != "_" {
		sc.assigns[id.Name] = append(sc.assigns[id.Name], assignment{value: value, tok: tok})
	}
}

// transportCall reports whether call sends a request and, if it does, its
// HTTP method and the expression that builds its path.
func (sc *funcScope) transportCall(call *ast.CallExpr) (method string, path ast.Expr, isCall bool, err error) {
	pathArg := 1
	switch fun := call.Fun.(type) {
	case *ast.Ident:
		if fun.Name != "FetchAllPages" {
			return "", nil, false, nil
		}
		method, pathArg = http.MethodGet, 2
	case *ast.SelectorExpr:
		recv, ok := fun.X.(*ast.Ident)
		m, transport := transportMethods[fun.Sel.Name]
		if !ok || !transport || !sc.clients[recv.Name] {
			return "", nil, false, nil
		}
		method = m
		if fun.Sel.Name == "Do" {
			pathArg = 2
			if len(call.Args) > 1 {
				method, err = httpMethod(call.Args[1])
			}
		}
	default:
		return "", nil, false, nil
	}
	if len(call.Args) <= pathArg {
		return "", nil, true, errors.New("the call has no path argument")
	}
	return method, call.Args[pathArg], true, err
}

// httpMethod reads Do's method argument: a net/http Method constant or a
// string literal, used verbatim (HTTP methods are case-sensitive).
func httpMethod(e ast.Expr) (string, error) {
	switch e := e.(type) {
	case *ast.SelectorExpr:
		if pkg, ok := e.X.(*ast.Ident); ok && pkg.Name == "http" && strings.HasPrefix(e.Sel.Name, "Method") {
			return strings.ToUpper(strings.TrimPrefix(e.Sel.Name, "Method")), nil
		}
	case *ast.BasicLit:
		if e.Kind == token.STRING {
			return strconv.Unquote(e.Value)
		}
	}
	return "", errors.New("the HTTP method is neither a net/http Method constant nor a string literal")
}

// lower turns a path expression into every template it can produce. A path
// is a string, so the only literal it can hold is a string literal and the
// only operator it can use is +.
func (sc *funcScope) lower(e ast.Expr) ([]string, error) {
	switch e := e.(type) {
	case *ast.ParenExpr:
		return sc.lower(e.X)
	case *ast.BasicLit:
		sc.consumed[e.Pos()] = true
		s, err := strconv.Unquote(e.Value)
		if err != nil {
			return nil, err
		}
		return []string{s}, nil
	case *ast.BinaryExpr:
		left, err := sc.lower(e.X)
		if err != nil {
			return nil, err
		}
		right, err := sc.lower(e.Y)
		if err != nil {
			return nil, err
		}
		var out []string
		for _, l := range left {
			for _, r := range right {
				out = append(out, l+r)
			}
		}
		return out, nil
	case *ast.CallExpr:
		return sc.lowerCall(e)
	case *ast.Ident:
		return sc.resolve(e.Name)
	}
	return []string{"{}"}, nil
}

// lowerCall lowers fmt.Sprintf (its verbs become {}) and BuildQuery (a query
// string); any other call yields a single value, {}.
func (sc *funcScope) lowerCall(call *ast.CallExpr) ([]string, error) {
	switch fun := call.Fun.(type) {
	case *ast.Ident:
		if fun.Name == "BuildQuery" {
			return []string{"?{}"}, nil
		}
	case *ast.SelectorExpr:
		pkg, ok := fun.X.(*ast.Ident)
		if !ok || pkg.Name != "fmt" || fun.Sel.Name != "Sprintf" {
			break
		}
		var lit *ast.BasicLit
		if len(call.Args) > 0 {
			lit, _ = call.Args[0].(*ast.BasicLit)
		}
		if lit == nil || lit.Kind != token.STRING {
			return nil, errors.New("the fmt.Sprintf format is not a string literal")
		}
		sc.consumed[lit.Pos()] = true
		format, err := strconv.Unquote(lit.Value)
		if err != nil {
			return nil, err
		}
		return []string{sprintfVerb.ReplaceAllStringFunc(format, func(verb string) string {
			if verb == "%%" {
				return "%"
			}
			return "{}"
		})}, nil
	}
	return []string{"{}"}, nil
}

// resolve lowers an identifier. A parameter is a value, {}. A variable is
// every value assigned to it anywhere in the function, so each one is checked
// whatever the control flow. Appending (+=) is accepted only for a query
// string, which leaves the route alone; anything else would make the route
// depend on the order of statements, which this scanner does not follow.
func (sc *funcScope) resolve(name string) ([]string, error) {
	assigns := sc.assigns[name]
	if len(assigns) == 0 {
		if sc.params[name] {
			return []string{"{}"}, nil
		}
		return nil, fmt.Errorf("%s is neither a parameter nor assigned in this function", name)
	}
	if sc.resolving[name] {
		return nil, fmt.Errorf("%s is assigned from itself", name)
	}
	sc.resolving[name] = true
	defer delete(sc.resolving, name)

	var values []string
	if sc.params[name] {
		values = append(values, "{}") // a reassigned parameter may still hold its argument
	}
	for _, a := range assigns {
		if a.value == nil {
			values = append(values, "{}")
			continue
		}
		lowered, err := sc.lower(a.value)
		if err != nil {
			return nil, err
		}
		if a.tok != token.ADD_ASSIGN {
			values = append(values, lowered...)
			continue
		}
		for _, l := range lowered {
			if !strings.HasPrefix(l, "?") {
				return nil, fmt.Errorf("%s is extended with += by something other than a query string; build the route in one expression", name)
			}
		}
	}
	if len(values) == 0 {
		return nil, fmt.Errorf("%s is only ever appended to", name)
	}
	return values, nil
}

// checkRoutes matches calls against the spec's routes. It returns one problem
// per call whose route is neither in the spec nor allowlisted, and per gaps
// entry that has no reason or is no longer needed, plus a note per call that
// is allowed only by gaps.
func checkRoutes(calls []clientCall, spec map[string]bool, gaps map[string]string, backend string) (problems, allowed []string) {
	called := map[string]bool{}
	for _, c := range calls {
		r := c.route()
		called[r] = true
		switch {
		case spec[r]:
		case strings.TrimSpace(gaps[r]) != "":
			allowed = append(allowed, fmt.Sprintf("%s: %s is not in the vendored spec; allowlisted: %s", c.pos, r, gaps[r]))
		default:
			problems = append(problems, fmt.Sprintf("%s: %s is not an operation in the vendored spec of backend %s", c.pos, r, backend))
		}
	}
	keys := make([]string, 0, len(gaps))
	for r := range gaps {
		keys = append(keys, r)
	}
	sort.Strings(keys)
	for _, r := range keys {
		switch {
		case strings.TrimSpace(gaps[r]) == "":
			problems = append(problems, fmt.Sprintf("knownRouteGaps[%q] gives no reason", r))
		case spec[r]:
			problems = append(problems, fmt.Sprintf("knownRouteGaps[%q] is stale: the vendored spec has this route now, so delete the entry", r))
		case !called[r]:
			problems = append(problems, fmt.Sprintf("knownRouteGaps[%q] is stale: no client call uses this route any more, so delete the entry", r))
		}
	}
	return problems, allowed
}

// specRoutes returns the spec's operations as "METHOD /path/{}" keys.
func specRoutes(data []byte) (map[string]bool, error) {
	var doc struct {
		Paths map[string]map[string]json.RawMessage `json:"paths"`
	}
	if err := json.Unmarshal(data, &doc); err != nil {
		return nil, err
	}
	routes := map[string]bool{}
	for path, item := range doc.Paths {
		for key := range item {
			switch key {
			case "get", "put", "post", "delete", "options", "head", "patch", "trace":
				routes[strings.ToUpper(key)+" "+specParam.ReplaceAllString(path, "{}")] = true
			}
		}
	}
	if len(routes) == 0 {
		return nil, errors.New("the spec declares no operations")
	}
	return routes, nil
}

// parseBackendVersion reads spec/BACKEND_VERSION: one line, a release tag.
func parseBackendVersion(content string) (string, error) {
	v := strings.TrimSpace(content)
	if !releaseTag.MatchString(v) {
		return "", fmt.Errorf("want one line holding a backend release tag such as v1.1.6, got %q", content)
	}
	return v, nil
}

// parseSpecChecksum reads spec/openapi3.json.sha256, one sha256sum line. The
// " *name" (binary mode) form is accepted as well as the "  name" one.
func parseSpecChecksum(content string) (string, error) {
	fields := strings.Fields(content)
	if len(fields) != 2 || strings.TrimPrefix(fields[1], "*") != "openapi3.json" || !sha256Hex.MatchString(fields[0]) {
		return "", fmt.Errorf("want one sha256sum line, \"<64 lower-case hex>  openapi3.json\", got %q", content)
	}
	return fields[0], nil
}

func position(fset *token.FileSet, pos token.Pos) string {
	p := fset.Position(pos)
	return fmt.Sprintf("%s:%d", filepath.Base(p.Filename), p.Line)
}

// parseClientSources parses this package's non-test Go files.
func parseClientSources(t *testing.T) (*token.FileSet, []*ast.File) {
	t.Helper()
	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatal(err)
	}
	fset := token.NewFileSet()
	var files []*ast.File
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() || !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}
		f, err := parser.ParseFile(fset, name, nil, parser.SkipObjectResolution)
		if err != nil {
			t.Fatal(err)
		}
		files = append(files, f)
	}
	return fset, files
}

func readSpecFile(t *testing.T, name string) []byte {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("spec", name)) // #nosec G304 -- name is one of this file's constant spec file names
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func readBackendVersion(t *testing.T) string {
	t.Helper()
	v, err := parseBackendVersion(string(readSpecFile(t, "BACKEND_VERSION")))
	if err != nil {
		t.Fatalf("spec/BACKEND_VERSION: %v", err)
	}
	return v
}

// The tests below pin the scanner and the matcher themselves, so that the
// contract test above cannot quietly stop checking anything.

// routeScanFixture holds one function per call shape the scanner must
// understand or refuse. It is parsed, never compiled.
const routeScanFixture = `package client

import (
	"context"
	"fmt"
	"net/http"
	"strings"
)

func (c *Client) Get(ctx context.Context, path string, result interface{}) error {
	_, err := c.Do(ctx, http.MethodGet, path, nil)
	return err
}

func FetchAllPages(ctx context.Context, c *Client, path, itemsKey string) ([]string, error) {
	_, err := c.Do(ctx, http.MethodGet, fmt.Sprintf("%s?page=%d", path, 1), nil)
	return nil, err
}

func (c *Client) Literal(ctx context.Context) error {
	return c.Get(ctx, "/api/v1/things", nil)
}

func (c *Client) Concat(ctx context.Context, id string) error {
	return c.Delete(ctx, "/api/v1/things/"+id)
}

func (c *Client) Sprintf(ctx context.Context, id string, n int) error {
	return c.Post(ctx, fmt.Sprintf("/api/v1/things/%s/items/%-3d?force=%t&max=100%%", id, n, true), nil, nil)
}

func (c *Client) FieldSegment(ctx context.Context, t Thing) error {
	return c.Get(ctx, "/api/v1/things/"+t.ID+"/items", nil)
}

func (c *Client) Alternatives(ctx context.Context, active bool, q string) error {
	var path string
	if active {
		path = "/api/v1/things/active"
	} else {
		path = "/api/v1/admin/things"
	}
	path += BuildQuery(map[string]string{"q": q})
	_, err := FetchAllPages(ctx, c, path, "things")
	return err
}

func (c *Client) DoMethods(ctx context.Context, id string) error {
	if _, err := c.Do(ctx, http.MethodPatch, "/api/v1/things/"+id, nil); err != nil {
		return err
	}
	_, err := c.Do(ctx, "POST", fmt.Sprintf("/api/v1/things/%s/refresh", id), nil)
	return err
}

func viaParam(ctx context.Context, cl *Client) error {
	return cl.Put(ctx, ("/api/v1/things"), nil, nil)
}

func (c *Client) NotAClientCall(h http.Header) string {
	return h.Get("X-Request-Id")
}

func (c *Client) UnresolvedFromCall(ctx context.Context, id string) error {
	return c.Get(ctx, pathFor(id), nil)
}

func (c *Client) UnresolvedMethod(ctx context.Context, method string) error {
	_, err := c.Do(ctx, method, "/api/v1/things", nil)
	return err
}

func (c *Client) UnresolvedAppend(ctx context.Context, id string) error {
	path := "/api/v1/things"
	path += "/" + id
	return c.Get(ctx, path, nil)
}

func (c *Client) UnresolvedSelfReference(ctx context.Context, id string) error {
	path := "/api/v1/things"
	path = path + "/" + id
	return c.Get(ctx, path, nil)
}

func (c *Client) UnresolvedReassignedParam(ctx context.Context, path string) error {
	path = strings.TrimSuffix(path, "/")
	return c.Get(ctx, path, nil)
}

func (c *Client) UnresolvedUnknownName(ctx context.Context) error {
	return c.Get(ctx, thingsPath, nil)
}

func (c *Client) UnresolvedOnlyAppended(ctx context.Context) error {
	var path string
	path += "?all=true"
	return c.Get(ctx, path, nil)
}

func (c *Client) UnresolvedFormat(ctx context.Context, format, id string) error {
	return c.Get(ctx, fmt.Sprintf(format, id), nil)
}

func (c *Client) UnresolvedNoPath(ctx context.Context) error {
	_, err := c.Do(ctx, http.MethodGet)
	return err
}

func (c *Client) UnresolvedRangeValue(ctx context.Context, paths []string) error {
	for _, p := range paths {
		if err := c.Get(ctx, p, nil); err != nil {
			return err
		}
	}
	return nil
}

func (c *Client) StrayViaWrapper(ctx context.Context) error {
	return c.fetch(ctx, "/api/v1/stray")
}

func (c *Client) fetch(ctx context.Context, path string) error {
	return c.Get(ctx, path, nil)
}
`

func TestScanRoutes_LowersEveryCallShape(t *testing.T) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "fixture.go", routeScanFixture, parser.SkipObjectResolution)
	if err != nil {
		t.Fatal(err)
	}
	scan := scanRoutes(fset, []*ast.File{f})

	got := map[string]bool{}
	for _, c := range scan.calls {
		got[c.route()] = true
	}
	want := []string{
		"GET /api/v1/things",              // Literal
		"DELETE /api/v1/things/{}",        // Concat
		"POST /api/v1/things/{}/items/{}", // Sprintf, flags and query dropped
		"GET /api/v1/things/{}/items",     // FieldSegment
		"GET /api/v1/things/active",       // Alternatives, both branches...
		"GET /api/v1/admin/things",        // ...with the BuildQuery append ignored
		"PATCH /api/v1/things/{}",         // DoMethods, net/http constant
		"POST /api/v1/things/{}/refresh",  // DoMethods, string literal
		"PUT /api/v1/things",              // viaParam, *Client parameter
	}
	for _, r := range want {
		if !got[r] {
			t.Errorf("scanRoutes missed %s", r)
		}
		delete(got, r)
	}
	for r := range got {
		t.Errorf("scanRoutes reported unexpected route %s", r)
	}

	unresolved := map[string]int{}
	for _, u := range scan.unresolved {
		_, fn, _ := strings.Cut(u, " (")
		fn, _, _ = strings.Cut(fn, ")")
		unresolved[fn]++
	}
	for _, fn := range []string{
		"UnresolvedFromCall",        // path from an arbitrary call
		"UnresolvedMethod",          // method from a variable
		"UnresolvedAppend",          // += of a path segment
		"UnresolvedSelfReference",   // path = path + ...
		"UnresolvedReassignedParam", // a parameter, even after reassignment
		"UnresolvedUnknownName",     // neither parameter nor local
		"UnresolvedOnlyAppended",    // never given a route
		"UnresolvedFormat",          // non-literal Sprintf format
		"UnresolvedNoPath",          // no path argument
		"UnresolvedRangeValue",      // a range variable
		"fetch",                     // a wrapper forwarding its parameter
	} {
		if unresolved[fn] != 1 {
			t.Errorf("want exactly one unresolved call in %s, got %d (all: %q)", fn, unresolved[fn], scan.unresolved)
		}
		delete(unresolved, fn)
	}
	for fn := range unresolved {
		t.Errorf("unexpected unresolved call in %s (all: %q)", fn, scan.unresolved)
	}

	if len(scan.stray) != 1 || !strings.Contains(scan.stray[0], `"/api/v1/stray"`) {
		t.Errorf("want only the literal handed to the fetch wrapper reported as stray, got %q", scan.stray)
	}
}

func TestCheckRoutes_ReportsGapsAndStaleAllowlistEntries(t *testing.T) {
	spec := map[string]bool{
		"GET /api/v1/things":       true,
		"DELETE /api/v1/things/{}": true,
		"GET /api/v1/documented":   true,
	}
	calls := []clientCall{
		{method: "GET", path: "/api/v1/things", pos: "a.go:1 (Served)"},
		{method: "DELETE", path: "/api/v1/things/{}", pos: "a.go:2 (Served)"},
		{method: "GET", path: "/api/v1/thingz", pos: "a.go:3 (Typo)"},
		{method: "PUT", path: "/api/v1/things", pos: "a.go:4 (WrongMethod)"},
		{method: "POST", path: "/api/v1/undocumented", pos: "a.go:5 (Allowlisted)"},
	}
	gaps := map[string]string{
		"POST /api/v1/undocumented": "served, but missing from this spec",
		"GET /api/v1/documented":    "the spec has it now",
		"DELETE /api/v1/uncalled":   "nothing calls it any more",
		"GET /api/v1/no-reason":     " ",
	}
	problems, allowed := checkRoutes(calls, spec, gaps, "v0.0.0-test")

	wantProblems := []string{
		`a.go:3 (Typo): GET /api/v1/thingz is not an operation in the vendored spec of backend v0.0.0-test`,
		`a.go:4 (WrongMethod): PUT /api/v1/things is not an operation in the vendored spec of backend v0.0.0-test`,
		`knownRouteGaps["DELETE /api/v1/uncalled"] is stale: no client call uses this route any more, so delete the entry`,
		`knownRouteGaps["GET /api/v1/documented"] is stale: the vendored spec has this route now, so delete the entry`,
		`knownRouteGaps["GET /api/v1/no-reason"] gives no reason`,
	}
	if strings.Join(problems, "\n") != strings.Join(wantProblems, "\n") {
		t.Errorf("problems:\n%s\nwant:\n%s", strings.Join(problems, "\n"), strings.Join(wantProblems, "\n"))
	}
	if len(allowed) != 1 || !strings.HasPrefix(allowed[0], "a.go:5 (Allowlisted): POST /api/v1/undocumented") {
		t.Errorf("want only the allowlisted call noted, got %q", allowed)
	}
}

func TestSpecRoutes(t *testing.T) {
	routes, err := specRoutes([]byte(`{"paths": {
		"/api/v1/things/{id}": {"parameters": [], "get": {}, "delete": {}},
		"/api/v1/things/{id}/items/{item_id}": {"summary": "x", "put": {}},
		"/version": {"get": {}}
	}}`))
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"DELETE /api/v1/things/{}", "GET /api/v1/things/{}", "GET /version", "PUT /api/v1/things/{}/items/{}"}
	got := make([]string, 0, len(routes))
	for r := range routes {
		got = append(got, r)
	}
	sort.Strings(got)
	if strings.Join(got, ", ") != strings.Join(want, ", ") {
		t.Errorf("specRoutes = %q, want %q", got, want)
	}

	for _, bad := range []string{`{"paths": {}}`, `{"paths": {"/x": {"parameters": []}}}`, `not json`} {
		if _, err := specRoutes([]byte(bad)); err == nil {
			t.Errorf("specRoutes(%s) accepted a spec with no operations", bad)
		}
	}
}

func TestParseBackendVersion(t *testing.T) {
	for in, want := range map[string]string{
		"v1.1.6\n":       "v1.1.6",
		"v4.27.0-rc.1\n": "v4.27.0-rc.1",
	} {
		if got, err := parseBackendVersion(in); err != nil || got != want {
			t.Errorf("parseBackendVersion(%q) = %q, %v; want %q", in, got, err, want)
		}
	}
	for _, in := range []string{"", "1.1.6\n", "main\n", "v1.1\n", "v1.1.6\nv1.1.7\n"} {
		if got, err := parseBackendVersion(in); err == nil {
			t.Errorf("parseBackendVersion(%q) = %q, want an error", in, got)
		}
	}
}

func TestParseSpecChecksum(t *testing.T) {
	const sum = "2742144ea90813ae08636cc125c7e7d50f3bf63e6008f5bc57ab23a1f40453da"
	for _, in := range []string{sum + "  openapi3.json\n", sum + " *openapi3.json\n"} {
		if got, err := parseSpecChecksum(in); err != nil || got != sum {
			t.Errorf("parseSpecChecksum(%q) = %q, %v; want %q", in, got, err, sum)
		}
	}
	for _, in := range []string{
		"",
		sum + "\n",
		sum + "  openapi3-patched.json\n",
		strings.ToUpper(sum) + "  openapi3.json\n",
		sum[:63] + "  openapi3.json\n",
		sum + "  openapi3.json\n" + sum + "  openapi3.json\n",
	} {
		if got, err := parseSpecChecksum(in); err == nil {
			t.Errorf("parseSpecChecksum(%q) = %q, want an error", in, got)
		}
	}
}
