package cigate

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

// The weekly OSV scan turns Go call-graph analysis OFF, and reachability is
// answered by a separate govulncheck job instead. This asserts the two halves
// stay joined, because today they are joined only by a comment.
//
// WHY THE SCAN HAS TO TURN IT OFF, and why that will not change.
//
// google/osv-scanner-action is a DOCKER action. Its image is built FROM a
// golang base image, so it carries exactly one Go toolchain, and the official
// golang images set GOTOOLCHAIN=local -- which forbids downloading or switching
// to another. A module whose `go` directive names a newer release than that
// toolchain cannot be built, so osv-scanner's internal govulncheck pass aborts.
//
// Measured on 2026-08-27 against the newest release, v2.5.1, published
// 2026-08-17, by pulling ghcr.io/google/osv-scanner-action:v2.5.1 and running
// the real scanner over a clean export of this repository:
//
//	image toolchain      go1.26.5, GOTOOLCHAIN=local
//	this module          go 1.26.6
//	call analysis on     "Failed to run code analysis (govulncheck) ...
//	                     go.mod requires go >= 1.26.6 (running go 1.26.5;
//	                     GOTOOLCHAIN=local)", exit 127, results array EMPTY
//	--no-call-analysis   exit 0, no abort
//
// Exit 127 carrying an empty report is the one shape that reads as both broken
// and clean at once, which is what sethbacon/terraform-registry-backend#894 was
// filed about. A newer pin does not fix it and cannot: the scanner image will
// always trail the toolchain this module builds on.
//
// WHY THIS TEST IS A COPY. The estate's replay apparatus carries the same class
// as the `osv-reachability-delegation` signature, but this repository is not in
// its SIGNATURE_SCOPE, so that signature never sees it. Adding it there is a
// scope.json change, which is a distributed migration across every host repo's
// replay copy. Until that happens the choice is a copy or no gate, and no gate
// is how the delegation became invisible in the first place.
//
// WHAT IT DOES NOT ASSERT. Not that call analysis must stay off forever: if an
// image ever ships a toolchain that can build this module, re-enabling it is a
// decision, and this test makes someone edit it rather than leaving a green run
// nobody looked at. Not that govulncheck found anything, and not that it is
// right -- only that a delegate exists, is pointed at this module with this
// module's toolchain, and is listened to.

const (
	osvAction      = "google/osv-scanner-action"
	noCallAnalysis = "--no-call-analysis=go"
)

type workflow struct {
	Jobs map[string]job `yaml:"jobs"`
}

type job struct {
	Defaults struct {
		Run struct {
			WorkingDirectory string `yaml:"working-directory"`
		} `yaml:"run"`
	} `yaml:"defaults"`
	Steps []step `yaml:"steps"`
}

type compositeAction struct {
	Runs struct {
		Steps []step `yaml:"steps"`
	} `yaml:"runs"`
}

type step struct {
	Uses             string         `yaml:"uses"`
	Run              string         `yaml:"run"`
	WorkingDirectory string         `yaml:"working-directory"`
	ContinueOnError  any            `yaml:"continue-on-error"`
	With             map[string]any `yaml:"with"`
}

func (s step) with(key string) string {
	v, ok := s.With[key]
	if !ok || v == nil {
		return ""
	}
	str, ok := v.(string)
	if !ok {
		return ""
	}
	return strings.TrimSpace(str)
}

func (s step) swallowed() bool {
	switch v := s.ContinueOnError.(type) {
	case bool:
		return v
	case string:
		return strings.TrimSpace(v) == "true"
	default:
		return false
	}
}

// repoRoot walks up from the test's directory to the tree holding the
// workflows. Resolving it by a fixed "../.." would silently start reading the
// wrong tree the moment this package moves.
func repoRoot(t *testing.T) string {
	t.Helper()
	dir, err := filepath.Abs(".")
	if err != nil {
		t.Fatalf("cannot resolve the working directory: %v", err)
	}
	for {
		if info, err := os.Stat(filepath.Join(dir, ".github", "workflows")); err == nil && info.IsDir() {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("walked to the filesystem root without finding .github/workflows")
		}
		dir = parent
	}
}

func loadWorkflows(t *testing.T, root string) map[string]workflow {
	t.Helper()
	dir := filepath.Join(root, ".github", "workflows")
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("cannot read %s: %v", dir, err)
	}
	out := map[string]workflow{}
	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() || (!strings.HasSuffix(name, ".yml") && !strings.HasSuffix(name, ".yaml")) {
			continue
		}
		raw, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			t.Fatalf("cannot read %s: %v", name, err)
		}
		var wf workflow
		if err := yaml.Unmarshal(raw, &wf); err != nil {
			// A workflow this test cannot read is a workflow it cannot clear.
			t.Fatalf("cannot parse .github/workflows/%s: %v", name, err)
		}
		out[name] = wf
	}
	if len(out) == 0 {
		t.Fatalf("no workflow files under %s", dir)
	}
	return out
}

// goModules lists every module directory as a slash-separated path relative to
// the repository root, "." for the root one. A module added later that nothing
// scans has to be visible here, or the gate quietly stops covering the repo.
func goModules(t *testing.T, root string) []string {
	t.Helper()
	skip := map[string]bool{".git": true, "vendor": true, "testdata": true, "node_modules": true}
	var found []string
	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			if skip[entry.Name()] {
				return filepath.SkipDir
			}
			return nil
		}
		if entry.Name() != "go.mod" {
			return nil
		}
		rel, err := filepath.Rel(root, filepath.Dir(path))
		if err != nil {
			return err
		}
		found = append(found, filepath.ToSlash(rel))
		return nil
	})
	if err != nil {
		t.Fatalf("cannot walk %s: %v", root, err)
	}
	if len(found) == 0 {
		t.Fatal("no go.mod found; this test is about Go reachability, so an empty universe is a failure and not a pass")
	}
	return found
}

func covers(scanRoot, module string) bool {
	scanRoot = strings.Trim(strings.TrimSpace(scanRoot), "./")
	if scanRoot == "" {
		return true
	}
	return module == scanRoot || strings.HasPrefix(module, scanRoot+"/")
}

// goVersionFile reports the go.mod a job resolves its toolchain from, following
// a local composite action. Both of this estate's Go backends set their
// toolchain that way, and a check that read only the workflow would report a
// correctly-wired job as broken.
func goVersionFile(t *testing.T, root string, steps []step) (file string, pinned string) {
	t.Helper()
	for _, s := range steps {
		switch {
		case strings.HasPrefix(s.Uses, "actions/setup-go"):
			if f := s.with("go-version-file"); f != "" {
				return f, ""
			}
			if p := s.with("go-version"); p != "" && pinned == "" {
				pinned = p
			}
		case strings.HasPrefix(s.Uses, "./"):
			dir := filepath.Join(root, filepath.FromSlash(strings.TrimPrefix(s.Uses, "./")))
			for _, candidate := range []string{"action.yml", "action.yaml"} {
				raw, err := os.ReadFile(filepath.Join(dir, candidate))
				if err != nil {
					continue
				}
				var act compositeAction
				if err := yaml.Unmarshal(raw, &act); err != nil {
					t.Fatalf("cannot parse %s/%s: %v", s.Uses, candidate, err)
				}
				f, p := goVersionFile(t, root, act.Runs.Steps)
				if f != "" {
					return f, ""
				}
				if p != "" && pinned == "" {
					pinned = p
				}
			}
		}
	}
	return "", pinned
}

// isGovulncheckCommand reports whether a script actually RUNS govulncheck.
// The token must be followed by whitespace, so
// `python3 scripts/govulncheck_triage.py` -- which names the tool in a filename
// and computes no reachability at all -- cannot satisfy the delegation.
// Whole-line comments are dropped first: a tool named in a comment is prose.
func isGovulncheckCommand(script string) bool {
	for _, line := range strings.Split(script, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		for _, field := range strings.FieldsFunc(trimmed, func(r rune) bool {
			return r == ';' || r == '|' || r == '&' || r == '(' || r == '\n'
		}) {
			if strings.HasPrefix(strings.TrimSpace(field), "govulncheck ") {
				return true
			}
		}
	}
	return false
}

func redirectTarget(script string) string {
	idx := strings.Index(script, ">")
	if idx < 0 {
		return ""
	}
	rest := strings.TrimSpace(script[idx+1:])
	if rest == "" {
		return ""
	}
	return strings.Fields(rest)[0]
}

type delegate struct {
	where      string
	module     string
	targetsAll bool
	toolchain  string
	pinned     string
	observed   bool
}

func TestOSVScanDelegatesGoReachabilityToAGovulncheckJobThatWorks(t *testing.T) {
	root := repoRoot(t)
	workflows := loadWorkflows(t, root)
	modules := goModules(t, root)

	type scan struct {
		where    string
		roots    []string
		declared bool
	}
	var scans []scan
	var delegates []delegate

	for name, wf := range workflows {
		for jobName, j := range wf.Jobs {
			where := ".github/workflows/" + name + ":" + jobName
			for i, s := range j.Steps {
				if strings.Contains(s.Uses, osvAction) {
					args := strings.Fields(s.with("scan-args"))
					var roots []string
					declared := false
					for _, arg := range args {
						if arg == noCallAnalysis {
							declared = true
							continue
						}
						if !strings.HasPrefix(arg, "-") {
							roots = append(roots, arg)
						}
					}
					if len(roots) == 0 {
						roots = []string{"."}
					}
					scans = append(scans, scan{where: where, roots: roots, declared: declared})
				}
				if !isGovulncheckCommand(s.Run) {
					continue
				}
				dir := s.WorkingDirectory
				if dir == "" {
					dir = j.Defaults.Run.WorkingDirectory
				}
				module := strings.Trim(strings.TrimSpace(dir), "./")
				if module == "" {
					module = "."
				}
				observed := !s.swallowed()
				if !observed {
					if target := redirectTarget(s.Run); target != "" {
						for _, later := range j.Steps[i+1:] {
							if strings.Contains(later.Run, target) {
								observed = true
								break
							}
							for _, v := range later.With {
								if str, ok := v.(string); ok && strings.Contains(str, target) {
									observed = true
								}
							}
						}
					}
				}
				file, pinned := goVersionFile(t, root, j.Steps)
				delegates = append(delegates, delegate{
					where:      where,
					module:     module,
					targetsAll: strings.Contains(s.Run, "./..."),
					toolchain:  file,
					pinned:     pinned,
					observed:   observed,
				})
			}
		}
	}

	// An empty universe is how a hard-coded name fails silently. If the scan
	// step or the delegate is renamed out of existence, that is the failure this
	// test exists for -- not a reason for it to find nothing and pass.
	if len(scans) == 0 {
		t.Fatalf("no %s step in any workflow; the weekly OSV scan this test is about does not exist", osvAction)
	}
	if len(delegates) == 0 {
		t.Fatalf("no govulncheck invocation in any workflow; nothing in this repository computes Go reachability")
	}

	for _, module := range modules {
		goMod := "go.mod"
		if module != "." {
			goMod = module + "/go.mod"
		}

		scanning := 0
		for _, s := range scans {
			for _, r := range s.roots {
				if covers(r, module) {
					scanning++
					if !s.declared {
						t.Errorf("%s#declared: %s scans this module without %s, so whether Go call analysis runs depends on the toolchain baked into a pinned container image rather than on a decision. See this file's header for the measurement.", goMod, s.where, noCallAnalysis)
					}
					break
				}
			}
		}
		if scanning == 0 {
			t.Errorf("%s#scanned: no %s step scans this module", goMod, osvAction)
		}

		var serving []delegate
		for _, d := range delegates {
			if d.module == module && d.targetsAll {
				serving = append(serving, d)
			}
		}
		if len(serving) == 0 {
			t.Errorf("%s#delegated: no govulncheck run covers this module, so nothing here computes Go reachability for it", goMod)
			continue
		}

		wired := false
		observed := false
		for _, d := range serving {
			if filepath.ToSlash(filepath.Clean(d.toolchain)) == goMod {
				wired = true
			}
			if d.observed {
				observed = true
			}
		}
		if !wired {
			for _, d := range serving {
				detail := "no step in the job sets up Go from a go.mod"
				if d.toolchain != "" {
					detail = "it resolves its toolchain from " + d.toolchain
				} else if d.pinned != "" {
					detail = "the job pins go-version " + d.pinned + " instead of reading a go.mod"
				}
				t.Errorf("%s#delegated: %s runs govulncheck but %s, not %s. That is the same toolchain drift that broke the scanner, moved one job over.", goMod, d.where, detail, goMod)
			}
		}
		if !observed {
			t.Errorf("%s#observed: every govulncheck run covering this module is continue-on-error and no later step in its job reads the file it writes, so a failed run reads exactly like a clean one", goMod)
		}
	}
}
