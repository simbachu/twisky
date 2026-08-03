// Command crapreport ranks functions by CRAP score from a Go coverprofile.
//
// CRAP = complexity² × (1 − coverage)³ + complexity
//
// Usage:
//
//	go generate ./static
//	go test ./... -coverprofile=coverage.out -covermode=atomic
//	go run ./scripts/crapreport -coverprofile=coverage.out
package main

import (
	"bufio"
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

type funcKey struct {
	pkg  string
	name string
}

type funcStats struct {
	pkg        string
	name       string
	file       string
	complexity int
	coverage   float64 // 0..1; -1 if unknown
	hasCover   bool
}

func main() {
	coverprofile := flag.String("coverprofile", "coverage.out", "path to coverprofile from go test")
	top := flag.Int("top", 25, "number of offenders to print")
	modulePath := flag.String("module", "", "module path (default: from go.mod)")
	flag.Parse()

	root, err := os.Getwd()
	if err != nil {
		fail(err)
	}
	mod := *modulePath
	if mod == "" {
		mod, err = readModulePath(filepath.Join(root, "go.mod"))
		if err != nil {
			fail(err)
		}
	}

	stats, err := collectComplexity(root, mod)
	if err != nil {
		fail(err)
	}

	cover, err := readFuncCoverage(*coverprofile, mod)
	if err != nil {
		fail(err)
	}
	for key, cov := range cover {
		matched := false
		for _, candidate := range []funcKey{
			key,
			{pkg: key.pkg, name: normalizeFuncName(key.name)},
		} {
			if s, ok := stats[candidate]; ok {
				s.coverage = cov
				s.hasCover = true
				stats[candidate] = s
				matched = true
				break
			}
		}
		if matched {
			continue
		}
		// go tool cover emits bare method names (addLabel); complexity uses Type.Method.
		for sk, s := range stats {
			if sk.pkg != key.pkg || s.hasCover {
				continue
			}
			if methodBase(sk.name) == key.name {
				s.coverage = cov
				s.hasCover = true
				stats[sk] = s
			}
		}
	}

	rows := make([]funcStats, 0, len(stats))
	for _, s := range stats {
		rows = append(rows, s)
	}
	sort.Slice(rows, func(i, j int) bool {
		ci, cj := crap(rows[i]), crap(rows[j])
		if ci != cj {
			return ci > cj
		}
		if rows[i].complexity != rows[j].complexity {
			return rows[i].complexity > rows[j].complexity
		}
		return rows[i].pkg+"."+rows[i].name < rows[j].pkg+"."+rows[j].name
	})

	n := *top
	if n > len(rows) {
		n = len(rows)
	}
	fmt.Printf("%-48s %6s %8s %8s\n", "FUNCTION", "CC", "COVER%", "CRAP")
	for i := 0; i < n; i++ {
		s := rows[i]
		coverPct := "n/a"
		if s.hasCover {
			coverPct = fmt.Sprintf("%5.1f%%", s.coverage*100)
		}
		fmt.Printf("%-48s %6d %8s %8.1f\n", s.pkg+"."+s.name, s.complexity, coverPct, crap(s))
	}
}

func crap(s funcStats) float64 {
	c := float64(s.complexity)
	cov := 0.0
	if s.hasCover {
		cov = s.coverage
		if cov < 0 {
			cov = 0
		}
		if cov > 1 {
			cov = 1
		}
	}
	uncovered := 1 - cov
	return c*c*math.Pow(uncovered, 3) + c
}

func readModulePath(goMod string) (string, error) {
	f, err := os.Open(goMod)
	if err != nil {
		return "", err
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if strings.HasPrefix(line, "module ") {
			return strings.TrimSpace(strings.TrimPrefix(line, "module ")), nil
		}
	}
	return "", fmt.Errorf("module path not found in %s", goMod)
}

func collectComplexity(root, module string) (map[funcKey]funcStats, error) {
	out := map[funcKey]funcStats{}
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			name := d.Name()
			if name == "vendor" || name == ".git" || name == "node_modules" || name == "testdata" || name == "scripts" || name == "cmd" {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)
		pkgPath := module
		dir := filepath.ToSlash(filepath.Dir(rel))
		if dir != "." {
			pkgPath = module + "/" + dir
		}

		fset := token.NewFileSet()
		file, err := parser.ParseFile(fset, path, nil, 0)
		if err != nil {
			return err
		}
		for _, decl := range file.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Body == nil {
				continue
			}
			name := fn.Name.Name
			if fn.Recv != nil && len(fn.Recv.List) > 0 {
				name = recvName(fn.Recv.List[0].Type) + "." + name
			}
			key := funcKey{pkg: pkgPath, name: name}
			out[key] = funcStats{
				pkg:        pkgPath,
				name:       name,
				file:       rel,
				complexity: cyclomatic(fn),
				coverage:   -1,
			}
		}
		return nil
	})
	return out, err
}

func recvName(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.StarExpr:
		return recvName(t.X)
	case *ast.Ident:
		return t.Name
	case *ast.IndexExpr:
		return recvName(t.X)
	case *ast.IndexListExpr:
		return recvName(t.X)
	default:
		return "?"
	}
}

func cyclomatic(fn *ast.FuncDecl) int {
	complexity := 1
	ast.Inspect(fn.Body, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.IfStmt:
			complexity++
		case *ast.ForStmt, *ast.RangeStmt:
			complexity++
		case *ast.CaseClause:
			if x.List != nil { // non-default case
				complexity++
			}
		case *ast.CommClause:
			if x.Comm != nil {
				complexity++
			}
		case *ast.BinaryExpr:
			if x.Op.String() == "&&" || x.Op.String() == "||" {
				complexity++
			}
		}
		return true
	})
	return complexity
}

func readFuncCoverage(coverprofile, module string) (map[funcKey]float64, error) {
	if _, err := os.Stat(coverprofile); err != nil {
		return nil, fmt.Errorf("coverprofile %s: %w (run go test -coverprofile first)", coverprofile, err)
	}
	cmd := exec.Command("go", "tool", "cover", "-func="+coverprofile)
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("go tool cover: %w\n%s", err, string(out))
	}

	cover := map[funcKey]float64{}
	sc := bufio.NewScanner(strings.NewReader(string(out)))
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "total:") {
			continue
		}
		// go tool cover -func format: path.go:line:\tName\t[stmts]\tcoverage%
		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue
		}
		pctStr := strings.TrimSuffix(parts[len(parts)-1], "%")
		pct, err := strconv.ParseFloat(pctStr, 64)
		if err != nil {
			continue
		}
		funcName := parts[1]
		filePath := strings.TrimSuffix(parts[0], ":")
		if colon := strings.LastIndex(filePath, ":"); colon >= 0 {
			filePath = filePath[:colon]
		}
		pkg := packageFromFile(filePath, module)
		cover[funcKey{pkg: pkg, name: funcName}] = pct / 100
	}
	return cover, sc.Err()
}

func packageFromFile(filePath, module string) string {
	filePath = filepath.ToSlash(filePath)
	if i := strings.Index(filePath, module+"/"); i >= 0 {
		rest := filePath[i+len(module)+1:]
		dir := filepath.ToSlash(filepath.Dir(rest))
		if dir == "." {
			return module
		}
		return module + "/" + dir
	}
	// already module-relative from go tool cover in some versions
	dir := filepath.ToSlash(filepath.Dir(filePath))
	if dir == "." {
		return module
	}
	if strings.HasPrefix(dir, module) {
		return dir
	}
	return module + "/" + dir
}

// normalizeFuncName maps go tool cover names like (*T).M / (T).M to T.M.
func normalizeFuncName(name string) string {
	if !strings.HasPrefix(name, "(") {
		return name
	}
	close := strings.Index(name, ").")
	if close < 0 {
		return name
	}
	recv := name[1:close]
	method := name[close+2:]
	recv = strings.TrimPrefix(recv, "*")
	return recv + "." + method
}

func methodBase(name string) string {
	if i := strings.LastIndex(name, "."); i >= 0 {
		return name[i+1:]
	}
	return name
}

func fail(err error) {
	fmt.Fprintf(os.Stderr, "crapreport: %v\n", err)
	os.Exit(1)
}
